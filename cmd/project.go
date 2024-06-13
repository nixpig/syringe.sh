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
	projectCtxKey = contextKey("PROJECT_CTX")
)

func projectCommand() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Aliases:           []string{"p"},
		Short:             "Project",
		Long:              "Project",
		Example:           "syringe project",
		PersistentPreRunE: initProjectContext,
	}

	projectCmd.AddCommand(projectAddCommand())

	return projectCmd
}

func projectAddCommand() *cobra.Command {
	projectAddCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "add",
		Long:    "add",
		Example: "syringe project add []",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			projectService := cmd.Context().Value(projectCtxKey).(services.ProjectService)

			if err := projectService.AddProject(services.AddProjectRequest{
				Name: name,
			}); err != nil {
				return err
			}

			return nil
		},
	}

	return projectAddCmd
}

func initProjectContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db := ctx.Value(dbCtxKey).(*sql.DB)

	projectStore := stores.NewSqliteProjectStore(db)
	projectService := services.NewProjectServiceImpl(projectStore, validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, projectCtxKey, projectService)

	cmd.SetContext(ctx)

	return nil
}
