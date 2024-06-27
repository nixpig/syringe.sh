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
			ctx, ok := sess.Context().(context.Context)
			if !ok {
				logger.Err(errors.New("context error")).Msg("failed to get session context")
				sess.Stderr().Write([]byte("failed to get context from session"))
				return
			}

			ctx = context.WithValue(ctx, ctxkeys.APP_DB, appDB)
			ctx = context.WithValue(ctx, ctxkeys.Username, sess.User())
			ctx = context.WithValue(ctx, ctxkeys.PublicKey, sess.PublicKey())

			rootCmd := root.New(ctx)

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

			userCmd := user.New(user.InitContext)
			userCmd.AddCommand(user.RegisterCmd(user.RegisterCmdHandler))
			rootCmd.AddCommand(userCmd)

			rootCmd.AddCommand(inject.InjectCommand())

			authenticated, ok := sess.Context().Value(ctxkeys.Authenticated).(bool)
			if !ok {
				logger.Warn().
					Str("session", sess.Context().SessionID()).
					Msg("failed to get authentication status from context")
				sess.Stderr().Write([]byte("Failed to establish authentication status"))
				return
			}

			if authenticated {
				userDB, err := database.NewUserDBConnection(sess.PublicKey())
				if err != nil {
					logger.Err(err).
						Str("session", sess.Context().SessionID()).
						Msg("failed to obtain user database connection")
					sess.Stderr().Write([]byte("Failed to obtain database connection using the provided public key"))
					return
				}

				// database connection is tightly coupled to and lasts only for the duration of the request
				defer userDB.Close()
				ctx = context.WithValue(ctx, ctxkeys.USER_DB, userDB)
			}

			rootCmd.SetArgs(sess.Command())
			rootCmd.SetIn(sess)
			rootCmd.SetOut(sess)
			rootCmd.SetErr(sess.Stderr())
			rootCmd.CompletionOptions.DisableDefaultCmd = true

			helpers.WalkCmd(rootCmd, func(c *cobra.Command) {
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
