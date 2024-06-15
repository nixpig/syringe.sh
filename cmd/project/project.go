package project

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg"
	"github.com/spf13/cobra"
)

type contextKey string

const (
	dbCtxKey      = pkg.DBCtxKey
	projectCtxKey = pkg.ProjectCtxKey
)

func ProjectCommand() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Aliases:           []string{"p"},
		Short:             "Manage projects",
		PersistentPreRunE: initProjectContext,
	}

	projectCmd.AddCommand(projectAddCommand())
	projectCmd.AddCommand(projectRemoveCommand())
	projectCmd.AddCommand(projectRenameCommand())
	projectCmd.AddCommand(projectListCommand())

	return projectCmd
}

func projectAddCommand() *cobra.Command {
	projectAddCmd := &cobra.Command{
		Use:     "add [flags] PROJECT_NAME",
		Aliases: []string{"a"},
		Short:   "Add a project",
		Example: "syringe project add my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    projectAddRunE,
	}

	return projectAddCmd
}

func projectAddRunE(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	projectService := cmd.Context().Value(projectCtxKey).(services.ProjectService)

	if err := projectService.Add(services.AddProjectRequest{
		Name: projectName,
	}); err != nil {
		return err
	}

	return nil
}

func projectRemoveCommand() *cobra.Command {
	projectRemoveCmd := &cobra.Command{
		Use:     "remove [flags] PROJECT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove a project",
		Example: "syringe project remove my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    projectRemoveRunE,
	}

	return projectRemoveCmd
}

func projectRemoveRunE(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	projectService := cmd.Context().Value(projectCtxKey).(services.ProjectService)

	if err := projectService.Remove(services.RemoveProjectRequest{
		Name: projectName,
	}); err != nil {
		return err
	}

	return nil
}

func projectRenameCommand() *cobra.Command {
	projectRenameCmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_PROJECT_NAME NEW_PROJECT_NAME",
		Aliases: []string{"r"},
		Short:   "Rename a project",
		Example: "syringe project rename my_cool_project my_awesome_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    projectRenameRunE,
	}

	return projectRenameCmd
}

func projectRenameRunE(cmd *cobra.Command, args []string) error {
	name := args[0]
	newName := args[1]

	projectService := cmd.Context().Value(projectCtxKey).(services.ProjectService)

	if err := projectService.Rename(services.RenameProjectRequest{
		Name:    name,
		NewName: newName,
	}); err != nil {
		return err
	}

	return nil
}

func projectListCommand() *cobra.Command {
	projectListCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List projects",
		Args:    cobra.NoArgs,
		Example: "syringe project list",
		RunE:    projectListRunE,
	}

	return projectListCmd
}

func projectListRunE(cmd *cobra.Command, args []string) error {
	projectService := cmd.Context().Value(projectCtxKey).(services.ProjectService)

	projects, err := projectService.List()
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		cmd.Println("No projects found!")
		cmd.Println("Try adding one with `syringe project add PROJECT_NAME`")
		return nil
	}

	for _, project := range projects {
		cmd.Println(project)
	}

	return nil
}

func initProjectContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(dbCtxKey).(*sql.DB)
	if !ok {
		return fmt.Errorf("unable to get database from context")
	}

	projectService := services.NewProjectServiceImpl(
		stores.NewSqliteProjectStore(db),
		validator.New(validator.WithRequiredStructEnabled()),
	)

	ctx = context.WithValue(ctx, projectCtxKey, projectService)

	cmd.SetContext(ctx)

	return nil
}
