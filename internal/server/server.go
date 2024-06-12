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
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/rs/zerolog"
)

type ContextKey string

const AUTHORISED_CTX = ContextKey("AUTHORISED")

type Server struct {
	app    services.AppService
	logger *zerolog.Logger
}

func NewServer(
	app services.AppService,
	logger *zerolog.Logger,
) Server {
	return Server{
		app:    app,
		logger: logger,
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
						s.logger.Err(err).
							Str("session", sess.Context().SessionID()).
							Msg("failed to execute command")
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
						s.logger.Warn().Msg("user not logged in")
						s.logger.Warn().Msg("prompt to register and call 'register' command if answer is 'Y', else return/exit")
						return
					}
					next(sess)
				}
			},
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					s.logger.Info().
						Str("session", sess.Context().SessionID()).
						Str("user", sess.User()).
						Str("address", sess.RemoteAddr().String()).
						Bool("publickey", sess.PublicKey() != nil).
						Str("client", sess.Context().ClientVersion()).
						Msg("connect")

					next(sess)

					s.logger.Info().
						Str("session", sess.Context().SessionID()).
						Msg("disconnect")
				}
			},
		),
	)
	if err != nil {
		return err
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	s.logger.Info().
		Str("host", host).
		Str("port", port).
		Msg("starting server")

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("failed to start server")
			done <- nil
		}
	}()

	<-done

	s.logger.Info().Msg("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		s.logger.Error().Err(err).Msg("failed to stop server")
		return err
	}

	return nil
}
