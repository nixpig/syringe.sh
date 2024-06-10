package server

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/rs/zerolog"
)

type contextKey string

const AUTHORISED = contextKey("AUTHORISED")

type SyringeSshServer struct {
	appService services.AppService
	log        *zerolog.Logger
}

func NewSyringeSshServer(
	appService services.AppService,
	log *zerolog.Logger,
) SyringeSshServer {
	return SyringeSshServer{
		appService: appService,
		log:        log,
	}
}

func (s SyringeSshServer) Start(host, port string) error {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					isAuthorised := sess.Context().Value(AUTHORISED)

					if !isAuthorised.(bool) {
						wish.Println(sess, "NOT AUTHORISED!!")
						return
					}

					err := cmd.Execute(sess, s.appService)
					if err != nil {
						os.Exit(1)
					}

					next(sess)
				}
			},
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					isAuthorised, err := s.appService.AuthenticateUser(services.UserAuthRequest{
						Username:  sess.User(),
						PublicKey: sess.PublicKey(),
					})
					if err != nil {
						return
					}

					sess.Context().SetValue(
						AUTHORISED,
						isAuthorised.Auth,
					)

					next(sess)
				}
			},
			logging.Middleware(),
		),
	)
	if err != nil {
		return err
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	s.log.Info().Msg("Starting SSH server")

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			s.log.Error().Err(err).Msg("Could not start server")
			done <- nil
		}
	}()

	<-done

	s.log.Info().Msg("Stopping SSH server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		s.log.Error().Err(err).Msg("Could not stop server")
	}

	return nil
}
