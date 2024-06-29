package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/go-playground/validator/v10"
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

func NewMiddlewareCommand(
	logger *zerolog.Logger,
	appDB *sql.DB,
	validate *validator.Validate,
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

			ctx = context.WithValue(ctx, ctxkeys.Username, sess.User())
			ctx = context.WithValue(ctx, ctxkeys.PublicKey, sess.PublicKey())

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
			}

			// -- COMMANDS
			cmdRoot := root.New(ctx)

			// -- USER CMD
			cmdUser := user.NewCmdUser()

			userService := user.NewUserServiceImpl(
				user.NewSqliteUserStore(appDB),
				validate,
				http.Client{},
				user.TursoAPISettings{
					URL:   os.Getenv("API_BASE_URL"),
					Token: os.Getenv("API_TOKEN"),
				},
			)

			handlerUserRegister := user.NewHandlerUserRegister(userService)
			cmdUser.AddCommand(user.NewCmdUserRegister(handlerUserRegister))
			cmdRoot.AddCommand(cmdUser)

			// -- PROJECT CMD
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

			cmdRoot.AddCommand(cmdProject)

			// -- ENVIRONMENT CMD
			cmdEnvironment := environment.NewCmdEnvironment()

			environmentService := environment.NewEnvironmentServiceImpl(
				environment.NewSqliteEnvironmentStore(userDB),
				validate,
			)

			handlerEnvironmentAdd := environment.NewHandlerEnvironmentAdd(environmentService)
			cmdEnvironmentAdd := environment.NewCmdEnvironmentAdd(handlerEnvironmentAdd)
			cmdEnvironment.AddCommand(cmdEnvironmentAdd)

			handlerEnvironmentRemove := environment.NewHandlerEnvironmentRemove(environmentService)
			cmdEnvironmentRemove := environment.NewCmdEnvironmentRemove(handlerEnvironmentRemove)
			cmdEnvironment.AddCommand(cmdEnvironmentRemove)

			handlerEnvironmentRename := environment.NewHandlerEnvironmentRename(environmentService)
			cmdEnvironmentRename := environment.NewCmdEnvironmentRename(handlerEnvironmentRename)
			cmdEnvironment.AddCommand(cmdEnvironmentRename)

			handlerEnvironmentList := environment.NewHandlerEnvironmentList(environmentService)
			cmdEnvironmentList := environment.NewCmdEnvironmentList(handlerEnvironmentList)
			cmdEnvironment.AddCommand(cmdEnvironmentList)

			cmdRoot.AddCommand(cmdEnvironment)

			// -- SECRET CMD
			cmdSecret := secret.NewCmdSecret()

			secretService := secret.NewSecretServiceImpl(
				secret.NewSqliteSecretStore(userDB),
				validate,
			)

			handlerSecretSet := secret.NewHandlerSecretSet(secretService)
			cmdSecretSet := secret.NewCmdSecretSet(handlerSecretSet)
			cmdSecret.AddCommand(cmdSecretSet)

			handlerSecretGet := secret.NewHandlerSecretGet(secretService)
			cmdSecretGet := secret.NewCmdSecretGet(handlerSecretGet)
			cmdSecret.AddCommand(cmdSecretGet)

			handlerSecretList := secret.NewHandlerSecretList(secretService)
			cmdSecretList := secret.NewCmdSecretList(handlerSecretList)
			cmdSecret.AddCommand(cmdSecretList)

			handlerSecretRemove := secret.NewHandlerSecretRemove(secretService)
			cmdSecretRemove := secret.NewCmdSecretRemove(handlerSecretRemove)
			cmdSecret.AddCommand(cmdSecretRemove)

			cmdRoot.AddCommand(cmdSecret)

			// -- INJECT CMD
			handlerInject := inject.NewHandlerInject(secretService)
			cmdInject := inject.NewCmdInject(handlerInject)
			cmdRoot.AddCommand(cmdInject)

			helpers.WalkCmd(cmdRoot, func(c *cobra.Command) {
				c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
				c.Flags().BoolP("version", "v", false, "Print version information")
			})

			// --------------------------------------

			cmdRoot.SetArgs(sess.Command())
			cmdRoot.SetIn(sess)
			cmdRoot.SetOut(sess)
			cmdRoot.SetErr(sess.Stderr())
			cmdRoot.CompletionOptions.DisableDefaultCmd = true

			if err := cmdRoot.ExecuteContext(ctx); err != nil {
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
