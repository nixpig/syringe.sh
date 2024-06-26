package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

func UserCommand(sess ssh.Session) *cobra.Command {
	userCmd := &cobra.Command{
		Use:               "user",
		Aliases:           []string{"u"},
		Short:             "Manage users",
		PersistentPreRunE: initUserContext,
	}

	userCmd.AddCommand(userRegisterCommand(sess))

	return userCmd
}

func userRegisterCommand(sess ssh.Session) *cobra.Command {
	userRegisterCmd := &cobra.Command{
		Use:     "register [flags] [USERNAME]",
		Aliases: []string{"r"},
		Short:   "Register user",
		Example: "syringe user register",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			userService, ok := cmd.Context().Value(ctxkeys.UserService).(UserService)
			if !ok {
				return fmt.Errorf("unable to get user service from context")
			}

			user, err := userService.RegisterUser(RegisterUserRequest{
				Username:  sess.User(),
				Email:     "not_used_yet@example.org",
				PublicKey: sess.PublicKey(),
			})
			if err != nil {
				return fmt.Errorf("unable to register user: %w", err)
			}

			cmd.Print(user)

			return nil
		},
	}

	return userRegisterCmd
}

func initUserContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.DB).(*sql.DB)
	if !ok {
		return fmt.Errorf("unable to get database from context")
	}

	userService := NewUserServiceImpl(
		NewSqliteUserStore(db),
		validation.NewValidator(),
		http.Client{},
		TursoAPISettings{
			URL:   os.Getenv("API_BASE_URL"),
			Token: os.Getenv("API_TOKEN"),
		},
	)

	ctx = context.WithValue(ctx, ctxkeys.UserService, userService)

	cmd.SetContext(ctx)

	return nil
}
