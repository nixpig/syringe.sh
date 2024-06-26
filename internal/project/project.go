package project

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

type CobraHandler func(cmd *cobra.Command, args []string) error

func New() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Aliases:           []string{"p"},
		Short:             "Manage projects",
		PersistentPreRunE: initProjectContext,
	}

	return projectCmd
}

func ProjectAddCommand(handler CobraHandler) *cobra.Command {
	projectAddCmd := &cobra.Command{
		Use:     "add [flags] PROJECT_NAME",
		Aliases: []string{"a"},
		Short:   "Add a project",
		Example: "syringe project add my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return projectAddCmd
}

func ProjectRemoveCommand(handler CobraHandler) *cobra.Command {
	projectRemoveCmd := &cobra.Command{
		Use:     "remove [flags] PROJECT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove a project",
		Example: "syringe project remove my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return projectRemoveCmd
}

func ProjectRenameCommand(handler CobraHandler) *cobra.Command {
	projectRenameCmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_PROJECT_NAME NEW_PROJECT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename a project",
		Example: "syringe project rename my_cool_project my_awesome_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	return projectRenameCmd
}

func ProjectListCommand(handler CobraHandler) *cobra.Command {
	projectListCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List projects",
		Args:    cobra.NoArgs,
		Example: "syringe project list",
		RunE:    handler,
	}

	return projectListCmd
}

func initProjectContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.DB).(*sql.DB)
	if !ok {
		return fmt.Errorf("unable to get database from context")
	}

	projectService := NewProjectServiceImpl(
		NewSqliteProjectStore(db),
		validation.NewValidator(),
	)

	ctx = context.WithValue(ctx, ctxkeys.ProjectService, projectService)

	cmd.SetContext(ctx)

	return nil
}
