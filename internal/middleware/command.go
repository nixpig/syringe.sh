package middleware

import (
	"context"
	"database/sql"
	"errors"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/environment"
	"github.com/nixpig/syringe.sh/internal/inject"
	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/internal/root"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/internal/user"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/helpers"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewCommandHandler(
	logger *zerolog.Logger,
	appDB *sql.DB,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {

			rootCmd := root.New(sess.Context())

			projectCmd := project.New(project.InitContext)
			projectCmd.AddCommand(project.AddCmd(project.AddCmdHandler))
			projectCmd.AddCommand(project.RemoveCmd(project.RemoveCmdHandler))
			projectCmd.AddCommand(project.RenameCmd(project.RenameCmdHandler))
			projectCmd.AddCommand(project.ListCmd(project.ListCmdHandler))
			rootCmd.AddCommand(projectCmd)

			environmentCmd := environment.New(environment.InitContext)
			environmentCmd.AddCommand(environment.AddCmd(environment.AddCmdHandler))
			environmentCmd.AddCommand(environment.RemoveCmd(environment.RemoveCmdHandler))
			environmentCmd.AddCommand(environment.RenameCmd(environment.RenameCmdHandler))
			environmentCmd.AddCommand(environment.ListCmd(environment.ListCmdHandler))
			rootCmd.AddCommand(environmentCmd)

			secretCmd := secret.New(secret.InitContext)
			secretCmd.AddCommand(secret.SetCmd(secret.SetCmdHandler))
			secretCmd.AddCommand(secret.GetCmd(secret.GetCmdHandler))
			secretCmd.AddCommand(secret.ListCmd(secret.ListCmdHandler))
			secretCmd.AddCommand(secret.RemoveCmd(secret.RemoveCmdHandler))
			rootCmd.AddCommand(secretCmd)

			rootCmd.AddCommand(inject.InjectCommand())
			rootCmd.AddCommand(user.UserCommand(sess))

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

			} else {
				db, err = database.NewUserDBConnection(sess.PublicKey())

				if err != nil {
					logger.Err(err).
						Str("session", sess.Context().SessionID()).
						Msg("failed to obtain user database connection")
					sess.Stderr().Write([]byte("Failed to obtain database connection using the provided public key"))
					return
				}

				// database connection is tightly coupled to and lasts only for the duration of the request
				defer db.Close()
				ctx = context.WithValue(ctx, ctxkeys.DB, db)
			}

			rootCmd.SetArgs(sess.Command())
			rootCmd.SetIn(sess)
			rootCmd.SetOut(sess)
			rootCmd.SetErr(sess.Stderr())
			rootCmd.CompletionOptions.DisableDefaultCmd = true

			helpers.CmdWalker(rootCmd, func(c *cobra.Command) {
				c.Flags().BoolP("help", "h", false, "Help for the "+c.Name()+" command")
			})

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
