package middleware

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"os"

	gossh "golang.org/x/crypto/ssh"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/internal/auth"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/environment"
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

func NewMiddlewareCommand(
	logger *zerolog.Logger,
	appDB *sql.DB,
	validate validation.Validator,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			var userDB *sql.DB
			var err error

			ctx, ok := sess.Context().(context.Context)
			if !ok {
				logger.Error().Err(errors.New("context error")).Msg("failed to get session context")
				sess.Stderr().Write([]byte("Error: failed to get context from session"))
				return
			}

			ctx = context.WithValue(ctx, ctxkeys.Username, sess.User())
			ctx = context.WithValue(ctx, ctxkeys.PublicKey, sess.PublicKey())

			authenticated, ok := sess.Context().Value(ctxkeys.Authenticated).(bool)
			if !ok {
				logger.Error().
					Str("session", sess.Context().SessionID()).
					Msg("failed to get authentication status from context")
				sess.Stderr().Write([]byte("Error: failed to establish authentication status"))
				return
			}

			if authenticated {
				marshalledKey := gossh.MarshalAuthorizedKey(sess.PublicKey())
				userDB, err = database.NewConnection(
					database.GetDatabasePath(
						fmt.Sprintf("%x.db", sha1.Sum(marshalledKey)),
					),
					os.Getenv("DB_USER"),
					os.Getenv("DB_PASSWORD"),
				)
				if err != nil {
					logger.Error().Err(err).
						Str("session", sess.Context().SessionID()).
						Msg("failed to obtain user database connection")
					sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to obtain database connection using the provided public key: %s", err)))
					return
				}

				// database connection is tightly coupled to, and lasts only for the duration of, the request
				defer userDB.Close()
			}

			cmdRoot := root.New(ctx, nil)

			// -- user
			cmdUser := user.NewCmdUser()

			userService := user.NewUserServiceImpl(
				user.NewSqliteUserStore(appDB),
				validate,
			)

			handlerUserRegister := user.NewHandlerUserRegister(userService)
			cmdUser.AddCommand(user.NewCmdUserRegister(handlerUserRegister))

			// -- project
			cmdProject := project.NewCmdProject()
			cmdProject.PersistentPreRunE = auth.PreRunEAuth

			projectService := project.NewProjectServiceImpl(
				project.NewSqliteProjectStore(userDB),
				validate,
			)

			cmdProjectAdd := project.NewCmdProjectAdd(
				project.NewHandlerProjectAdd(projectService),
			)

			cmdProjectRemove := project.NewCmdProjectRemove(
				project.NewHandlerProjectRemove(projectService),
			)

			cmdProjectRename := project.NewCmdProjectRename(
				project.NewHandlerProjectRename(projectService),
			)

			cmdProjectList := project.NewCmdProjectList(
				project.NewHandlerProjectList(projectService),
			)

			cmdProject.AddCommand(
				cmdProjectAdd,
				cmdProjectRemove,
				cmdProjectRename,
				cmdProjectList,
			)

			// -- environment
			cmdEnvironment := environment.NewCmdEnvironment()
			cmdEnvironment.PersistentPreRunE = auth.PreRunEAuth

			environmentService := environment.NewEnvironmentServiceImpl(
				environment.NewSqliteEnvironmentStore(userDB),
				validate,
			)

			cmdEnvironmentAdd := environment.NewCmdEnvironmentAdd(
				environment.NewHandlerEnvironmentAdd(environmentService),
			)

			cmdEnvironmentRemove := environment.NewCmdEnvironmentRemove(
				environment.NewHandlerEnvironmentRemove(environmentService),
			)

			cmdEnvironmentRename := environment.NewCmdEnvironmentRename(
				environment.NewHandlerEnvironmentRename(environmentService),
			)

			cmdEnvironmentList := environment.NewCmdEnvironmentList(
				environment.NewHandlerEnvironmentList(environmentService),
			)

			cmdEnvironment.AddCommand(
				cmdEnvironmentAdd,
				cmdEnvironmentRemove,
				cmdEnvironmentRename,
				cmdEnvironmentList,
			)

			// -- SECRET CMD
			cmdSecret := secret.NewCmdSecret()
			cmdSecret.PersistentPreRunE = auth.PreRunEAuth

			secretService := secret.NewSecretServiceImpl(
				secret.NewSqliteSecretStore(userDB),
				validate,
			)

			cmdSecretSet := secret.NewCmdSecretSet(
				secret.NewSSHHandlerSecretSet(secretService),
			)

			cmdSecretGet := secret.NewCmdSecretGet(
				secret.NewSSHHandlerSecretGet(secretService),
			)

			cmdSecretList := secret.NewCmdSecretList(
				secret.NewSSHHandlerSecretList(secretService),
			)

			cmdSecretRemove := secret.NewCmdSecretRemove(
				secret.NewSSHHandlerSecretRemove(secretService),
			)

			cmdSecretInject := secret.NewCmdSecretInject(
				secret.NewSSHHandlerSecretInject(secretService),
			)
			cmdSecretInject.PersistentPreRunE = auth.PreRunEAuth

			cmdSecret.AddCommand(
				cmdSecretSet,
				cmdSecretGet,
				cmdSecretList,
				cmdSecretRemove,
				cmdSecretInject,
			)

			cmdRoot.AddCommand(
				cmdUser,
				cmdProject,
				cmdEnvironment,
				cmdSecret,
			)

			helpers.WalkCmd(cmdRoot, func(c *cobra.Command) {
				c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
				c.Flags().BoolP("version", "v", false, "Print version information")
			})

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
