package main

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
	"github.com/rs/zerolog"
)

type Server struct {
	logger      *zerolog.Logger
	middleware  []wish.Middleware
	timeout     time.Duration
	hostKeyPath string
}

func newServer(
	logger *zerolog.Logger,
	middleware []wish.Middleware,
	timeout time.Duration,
	hostKeyPath string,
) Server {
	return Server{
		logger:      logger,
		middleware:  middleware,
		timeout:     timeout,
		hostKeyPath: hostKeyPath,
	}
}

func (s Server) Start(host, port string) error {
	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(s.hostKeyPath),
		wish.WithMaxTimeout(s.timeout),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519" || key.Type() == "ssh-rsa"
		}),
		wish.WithMiddleware(
			s.middleware...,
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
