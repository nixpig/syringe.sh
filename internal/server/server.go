package server

import (
	"context"
	"errors"
	"fmt"
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

type ContextKey string

const AUTHORISED_CTX = ContextKey("AUTHORISED")

type Server struct {
	app services.AppService
	log *zerolog.Logger
}

func NewServer(
	app services.AppService,
	log *zerolog.Logger,
) Server {
	return Server{
		app: app,
		log: log,
	}
}

func (s Server) Start(host, port string) error {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(
			// exec cobra
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					if err := cmd.Execute(sess, s.app); err != nil {
						fmt.Println("error from cmd")
						os.Exit(1)
					}

					next(sess)
				}
			},
			// authenticate user
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					if _, err := s.app.AuthenticateUser(services.UserAuthRequest{
						Username:  sess.User(),
						PublicKey: sess.PublicKey(),
					}); err != nil {
						wish.Println(sess, "NOT AUTHORISED!!")
						wish.Println(sess, "prompt to register and call 'register' command if answer is 'Y', else return/exit")
						return
					}

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
