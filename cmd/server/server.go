package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

var allowedKeyTypes = []string{"ssh-rsa", "ssh-ed25519"}

type syringeServer struct {
	s *ssh.Server
}

func (s syringeServer) New(
	host string,
	hostKeyPath string,
	m ...wish.Middleware,
) (*syringeServer, error) {
	server, err := wish.NewServer(
		wish.WithAddress(host),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMaxTimeout(time.Second*60),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return slices.Contains(allowedKeyTypes, key.Type())
		}),
		wish.WithMiddleware(m...),
	)
	if err != nil && err != ssh.ErrServerClosed {
		return nil, fmt.Errorf("server stopped not gracefully: %w", err)
	}

	return &syringeServer{
		s: server,
	}, nil
}
