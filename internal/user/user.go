package user

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

func NewCmdUser(init pkg.CobraHandler) *cobra.Command {
	userCmd := &cobra.Command{
		Use:               "user",
		Aliases:           []string{"u"},
		Short:             "Manage users",
		PersistentPreRunE: init,
	}

	return userCmd
}

func NewCmdUserRegister(handler pkg.CobraHandler) *cobra.Command {
	registerCmd := &cobra.Command{
		Use:     "register [flags] [USERNAME]",
		Aliases: []string{"r"},
		Short:   "Register user",
		Example: "syringe user register",
		Args:    cobra.NoArgs,
		RunE:    handler,
	}

	return registerCmd
}

func InitContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.APP_DB).(*sql.DB)
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
