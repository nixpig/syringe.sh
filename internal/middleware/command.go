package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/environment"
	"github.com/nixpig/syringe.sh/internal/inject"
	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/internal/root"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/internal/user"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/helpers"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewCommandHandler(
	logger *zerolog.Logger,
	appDB *sql.DB,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			var userDB *sql.DB
			var err error

			ctx, ok := sess.Context().(context.Context)
			if !ok {
				logger.Error().Err(errors.New("context error")).Msg("failed to get session context")
				sess.Stderr().Write([]byte("failed to get context from session"))
				return
			}

			ctx = context.WithValue(ctx, ctxkeys.APP_DB, appDB)
			ctx = context.WithValue(ctx, ctxkeys.Username, sess.User())
			ctx = context.WithValue(ctx, ctxkeys.PublicKey, sess.PublicKey())

			rootCmd := root.New(ctx)

			environmentCmd := environment.NewCmdEnvironment(environment.InitContext)
			environmentCmd.AddCommand(environment.NewCmdEnvironmentAdd(environment.AddCmdHandler))
			environmentCmd.AddCommand(environment.NewCmdEnvironmentRemove(environment.RemoveCmdHandler))
			environmentCmd.AddCommand(environment.NewCmdEnvironmentRename(environment.RenameCmdHandler))
			environmentCmd.AddCommand(environment.NewCmdEnvironmentList(environment.ListCmdHandler))
			rootCmd.AddCommand(environmentCmd)

			secretCmd := secret.NewCmdSecret(secret.InitContext)
			secretCmd.AddCommand(secret.NewCmdSecretSet(secret.SetCmdHandler))
			secretCmd.AddCommand(secret.NewCmdSecretGet(secret.GetCmdHandler))
			secretCmd.AddCommand(secret.NewCmdSecretList(secret.ListCmdHandler))
			secretCmd.AddCommand(secret.NewCmdSecretRemove(secret.RemoveCmdHandler))
			rootCmd.AddCommand(secretCmd)

			userCmd := user.NewCmdUser(user.InitContext)
			userCmd.AddCommand(user.NewCmdUserRegister(user.RegisterCmdHandler))
			rootCmd.AddCommand(userCmd)

			injectCmd := inject.NewCmdInjectWithHandler(inject.InitContext, inject.InjectCmdHandler)
			rootCmd.AddCommand(injectCmd)

			authenticated, ok := sess.Context().Value(ctxkeys.Authenticated).(bool)
			if !ok {
				logger.Warn().
					Str("session", sess.Context().SessionID()).
					Msg("failed to get authentication status from context")
				sess.Stderr().Write([]byte("Failed to establish authentication status"))
				return
			}

			if authenticated {
				userDB, err = database.NewUserDBConnection(sess.PublicKey())
				if err != nil {
					logger.Error().Err(err).
						Str("session", sess.Context().SessionID()).
						Msg("failed to obtain user database connection")
					sess.Stderr().Write([]byte("Failed to obtain database connection using the provided public key"))
					return
				}

				// database connection is tightly coupled to and lasts only for the duration of the request
				defer userDB.Close()
				ctx = context.WithValue(ctx, ctxkeys.USER_DB, userDB)
			}

			// --------------------------------------
			validate := validation.NewValidator()

			cmdProject := project.NewCmdProject()

			projectService := project.NewProjectServiceImpl(
				project.NewSqliteProjectStore(userDB),
				validate,
			)

			handlerProjectAdd := project.NewHandlerProjectAdd(projectService)
			cmdProjectAdd := project.NewCmdProjectAdd(handlerProjectAdd)
			cmdProject.AddCommand(cmdProjectAdd)

			handlerProjectRemove := project.NewHandlerProjectRemove(projectService)
			cmdProjectRemove := project.NewCmdProjectRemove(handlerProjectRemove)
			cmdProject.AddCommand(cmdProjectRemove)

			handlerProjectRename := project.NewHandlerProjectRename(projectService)
			cmdProjectRename := project.NewCmdProjectRename(handlerProjectRename)
			cmdProject.AddCommand(cmdProjectRename)

			handlerProjectList := project.NewHandlerProjectList(projectService)
			cmdProjectList := project.NewCmdProjectList(handlerProjectList)
			cmdProject.AddCommand(cmdProjectList)

			rootCmd.AddCommand(cmdProject)

			// --------------------------------------

			rootCmd.SetArgs(sess.Command())
			rootCmd.SetIn(sess)
			rootCmd.SetOut(sess)
			rootCmd.SetErr(sess.Stderr())
			rootCmd.CompletionOptions.DisableDefaultCmd = true

			helpers.WalkCmd(rootCmd, func(c *cobra.Command) {
				c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
				c.Flags().BoolP("version", "v", false, "Print version information")
			})

			if err := rootCmd.ExecuteContext(ctx); err != nil {
				logger.Error().
					Err(err).
					Str("session", sess.Context().SessionID()).
					Any("command", sess.Command()).
					Msg("failed to execute command")

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
