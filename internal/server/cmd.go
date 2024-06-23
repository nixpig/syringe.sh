package server

import (
	"errors"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/cmd/environment"
	"github.com/nixpig/syringe.sh/server/cmd/project"
	"github.com/nixpig/syringe.sh/server/cmd/secret"
	"github.com/nixpig/syringe.sh/server/cmd/user"
	"github.com/spf13/cobra"
)

func cobraHandler(s Server) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			db, err := newUserDBConnection(sess.PublicKey())
			if err != nil {
				s.logger.Err(err).
					Str("session", sess.Context().SessionID()).
					Msg("failed to obtain user database connection")
				sess.Stderr().Write([]byte("Failed to obtain database connection using the provided public key"))
				return
			}

			defer db.Close()

			if err := cmd.Execute(
				[]*cobra.Command{
					user.UserCommand(),
					project.ProjectCommand(),
					environment.EnvironmentCommand(),
					secret.SecretCommand(),
				},
				sess.Command(),
				sess,
				sess,
				sess.Stderr(),
				db,
			); err != nil {
				s.logger.Err(errors.Unwrap(err)).
					Str("session", sess.Context().SessionID()).
					Any("command", sess.Command()).
					Msg("failed to execute command")
			}

			s.logger.Info().
				Str("session", sess.Context().SessionID()).
				Any("command", sess.Command()).
				Msg("executed command")

			next(sess)
		}
	}
}
