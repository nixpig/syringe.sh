package cmd

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/spf13/cobra"
)

func environmentCommand() *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:               "environment",
		Aliases:           []string{"e"},
		Short:             "Environment",
		Long:              "Environment",
		Example:           "syringe environment",
		PersistentPreRunE: initEnvironmentContext,
	}

	environmentCmd.AddCommand(environmentAddCommand())

	return environmentCmd
}

func environmentAddCommand() *cobra.Command {
	environmentAddCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "add",
		Long:    "add",
		Example: "syringe environment add []",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {

			environment := args[0]

			project, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}

			environmentService := cmd.Context().Value("ENVIRONMENT_SERVICE").(services.EnvironmentService)

			if err := environmentService.AddEnvironment(services.AddEnvironmentRequest{
				Name:        environment,
				ProjectName: project,
			}); err != nil {
				return err
			}

			return nil
		},
	}

	environmentAddCmd.Flags().StringP("project", "p", "", "Project")

	environmentAddCmd.MarkFlagRequired("project")

	return environmentAddCmd
}

func initEnvironmentContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db := ctx.Value(DB_CTX).(*sql.DB)

	environmentStore := stores.NewSqliteEnvironmentStore(db)
	environmentService := services.NewEnvironmentServiceImpl(environmentStore, validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, "ENVIRONMENT_SERVICE", environmentService)

	cmd.SetContext(ctx)

	return nil
}
