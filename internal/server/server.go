package server

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/pkg/turso"
	"github.com/rs/zerolog"
	gossh "golang.org/x/crypto/ssh"
)

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
			cobraHandler(s),
			authAndRegisterHandler(s),
			loggingHandler(s),
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

// TODO: really don't like this!!
func newUserDBConnection(publicKey ssh.PublicKey) (*sql.DB, error) {
	api := turso.New(
		os.Getenv("DATABASE_ORG"),
		os.Getenv("API_TOKEN"),
		http.Client{},
	)

	marshalledKey := gossh.MarshalAuthorizedKey(publicKey)

	hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))
	expiration := "30s"

	token, err := api.CreateToken(hashedKey, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to create token:\n%s", err)
	}

	fmt.Println("creating new user-specific db connection")
	db, err := database.Connection(
		"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
		string(token.Jwt),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection:\n%s", err)
	}

	return db, nil
}
