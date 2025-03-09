package server

import (
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

var allowedKeyTypes = []string{
	"ssh-rsa",
	"ssh-ed25519",
}

type server struct {
	*ssh.Server
}

func New(
	host string,
	port string,
	hostKeyPath string,
	m ...wish.Middleware,
) (*server, error) {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
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

	return &server{s}, nil
}
