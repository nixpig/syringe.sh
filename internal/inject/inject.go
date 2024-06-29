package inject

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

func NewCmdInjectWithHandler(init pkg.CobraHandler, run pkg.CobraHandler) *cobra.Command {
	injectCmd := NewCmdInject(init)
	injectCmd.RunE = run

	return injectCmd
}

func NewCmdInject(init pkg.CobraHandler) *cobra.Command {
	injectCmd := &cobra.Command{
		Use:     "inject [flags] -- SUBCOMMAND",
		Aliases: []string{"i"},
		Short:   "Inject secrets",
		Long:    "Inject secrets into the specified subcommand.",
		Example: `  # Inject secrets from 'dev' environment in 'my_cool_project' project into 'startserver' command
    syringe inject -p my_cool_project -e dev -- startserver`,
		PersistentPreRunE: init,
		Args:              cobra.MinimumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	injectCmd.Flags().StringP("project", "p", "", "Project name")
	injectCmd.MarkFlagRequired("project")

	injectCmd.Flags().StringP("environment", "e", "", "Environment name")
	injectCmd.MarkFlagRequired("environment")

	return injectCmd
}

func InitContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.USER_DB).(*sql.DB)
	if !ok {
		return fmt.Errorf("failed to get database from context")
	}

	secretStore := secret.NewSqliteSecretStore(db)
	secretService := secret.NewSecretServiceImpl(
		secretStore,
		validation.NewValidator(),
	)

	ctx = context.WithValue(ctx, ctxkeys.SecretService, secretService)

	cmd.SetContext(ctx)

	return nil
}
