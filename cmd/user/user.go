package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/server/pkg/validation"
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
	// userCmd.AddCommand(UserDeleteCommand())

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

func UserDeleteCommand() *cobra.Command {
	userDeleteCmd := &cobra.Command{
		Use:     "delete [flags] [USERNAME]",
		Aliases: []string{"d"},
		Short:   "Delete a user",
		Example: "syringe user delete -i ~/.ssh/id_rsa",
		Args:    cobra.NoArgs,
		RunE:    userDeleteRunE,
	}

	return userDeleteCmd
}

func userDeleteRunE(cmd *cobra.Command, args []string) error {
	fmt.Println("delete user...")

	return nil
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
