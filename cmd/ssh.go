package cmd

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
	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/rs/zerolog"
)

const (
	host = "localhost"
	port = "23234"
)

type SyringeSshServer struct {
	handlers handlers.SshHandlers
	log      *zerolog.Logger
}

func NewSyringeSshServer(
	handlers handlers.SshHandlers,
	log *zerolog.Logger,
) SyringeSshServer {
	return SyringeSshServer{
		handlers: handlers,
		log:      log,
	}
}

func (s SyringeSshServer) Start() error {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(func(next ssh.Handler) ssh.Handler {
			return func(sess ssh.Session) {
				wish.Println(sess, sess.Command())
				authed := s.handlers.AuthUser(sess.User(), sess.PublicKey())

				if authed {
					wish.Println(sess, "You are authed!!")
				} else {
					wish.Println(sess, "Hey, I don't know who you are!")
					wish.Println(sess, "Please hold on while I register you...")

					s.handlers.RegisterUser(sess.User(), sess.PublicKey())

					// args := sess.Command()
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
	defer func() { cancel() }()
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		s.log.Error().Err(err).Msg("Could not stop server")
	}

	return nil
}
