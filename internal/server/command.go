package server

import (
	"database/sql"
	"errors"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/cmd/server/servercmd"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/environment"
	"github.com/nixpig/syringe.sh/internal/inject"
	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/internal/user"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewCommandHandler(
	logger *zerolog.Logger,
	appDB *sql.DB,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			authenticated, ok := sess.Context().Value(ctxkeys.Authenticated).(bool)
			if !ok {
				logger.Warn().
					Str("session", sess.Context().SessionID()).
					Msg("failed to get authentication status from context")

				sess.Stderr().Write([]byte("Failed to establish authentication status"))
				return
			}

			var commands []*cobra.Command
			var db *sql.DB
			var err error

			if !authenticated {
				db = appDB

				commands = []*cobra.Command{
					user.UserCommand(sess),
				}
			} else {
				db, err = database.NewUserDBConnection(sess.PublicKey())
				if err != nil {
					logger.Err(err).
						Str("session", sess.Context().SessionID()).
						Msg("failed to obtain user database connection")
					sess.Stderr().Write([]byte("Failed to obtain database connection using the provided public key"))
					return
				}

				defer db.Close()

				commands = []*cobra.Command{
					inject.InjectCommand(),
					project.ProjectCommand(),
					environment.EnvironmentCommand(),
					secret.SecretCommand(),
				}
			}

			if err := servercmd.Execute(
				commands,
				sess.Command(),
				sess,
				sess,
				sess.Stderr(),
				db,
			); err != nil {
				logger.Err(errors.Unwrap(err)).
					Str("session", sess.Context().SessionID()).
					Any("command", sess.Command()).
					Msg("failed to execute command")

				wish.Fatal(sess)
				next(sess)
				return
			}

			logger.Info().
				Str("session", sess.Context().SessionID()).
				Any("command", sess.Command()).
				Msg("executed command")

			next(sess)
		}
	}
}
