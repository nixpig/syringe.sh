package middleware

import (
	"context"
	"database/sql"
	"errors"
	"io"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/cmd/server/handlers"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/environment"
	"github.com/nixpig/syringe.sh/internal/inject"
	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/internal/root"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/internal/user"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type Executor func(
	rootCmd *cobra.Command,
	args []string,
	cmdIn io.Reader,
	cmdOut io.Writer,
	cmdErr io.ReadWriter,
	db *sql.DB,
) error

func NewCommandHandler(
	logger *zerolog.Logger,
	appDB *sql.DB,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {

			rootCmd := root.New()

			projectCmd := project.New()

			projectCmd.AddCommand(project.ProjectRemoveCommand(handlers.NewProjectRemoveHandler()))
			projectCmd.AddCommand(project.ProjectRenameCommand(handlers.NewProjectRenameHandler()))
			projectCmd.AddCommand(project.ProjectAddCommand(handlers.NewProjectAddHandler()))
			projectCmd.AddCommand(project.ProjectListCommand(handlers.NewProjectListHandler()))

			rootCmd.AddCommand(inject.InjectCommand())
			rootCmd.AddCommand(environment.EnvironmentCommand())
			rootCmd.AddCommand(secret.SecretCommand())

			rootCmd.AddCommand(projectCmd)

			authenticated, ok := sess.Context().Value(ctxkeys.Authenticated).(bool)
			if !ok {
				logger.Warn().
					Str("session", sess.Context().SessionID()).
					Msg("failed to get authentication status from context")

				sess.Stderr().Write([]byte("Failed to establish authentication status"))
				return
			}

			var db *sql.DB
			var err error
			ctx := context.Background()

			if !authenticated {
				db = appDB

				rootCmd.AddCommand(user.UserCommand(sess))
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

				ctx = context.WithValue(ctx, ctxkeys.DB, db)
			}

			rootCmd.SetArgs(sess.Command())
			rootCmd.SetIn(sess)
			rootCmd.SetOut(sess)
			rootCmd.SetErr(sess.Stderr())
			rootCmd.CompletionOptions.DisableDefaultCmd = true

			// walk(rootCmd, func(c *cobra.Command) {
			// 	c.Flags().BoolP("help", "h", false, "Help for the "+c.Name()+" command")
			// })

			if err := rootCmd.ExecuteContext(ctx); err != nil {
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

func walk(c *cobra.Command, f func(*cobra.Command)) {
	f(c)
	for _, c := range c.Commands() {
		walk(c, f)
	}
}
