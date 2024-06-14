package cmd

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/spf13/cobra"
)

const (
	environmentCtxKey = contextKey("ENVIRONMENT_CTX")
)

func environmentCommand() *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:               "environment",
		Aliases:           []string{"e"},
		Short:             "Manage environments",
		PersistentPreRunE: initEnvironmentContext,
	}

	environmentCmd.AddCommand(environmentAddCommand())

	return environmentCmd
}

func environmentAddCommand() *cobra.Command {
	environmentAddCmd := &cobra.Command{
		Use:     "add [flags] ENVIRONMENT_NAME",
		Aliases: []string{"a"},
		Short:   "Add an environment",
		Example: "syringe environment add -p my_cool_project local",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {

			environment := args[0]

			project, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}

			environmentService := cmd.Context().Value(environmentCtxKey).(services.EnvironmentService)

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

	db := ctx.Value(dbCtxKey).(*sql.DB)

	environmentStore := stores.NewSqliteEnvironmentStore(db)
	environmentService := services.NewEnvironmentServiceImpl(environmentStore, validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, environmentCtxKey, environmentService)

	cmd.SetContext(ctx)

	return nil
}
