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

func New(init CobraHandler) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Aliases:           []string{"p"},
		Short:             "Manage projects",
		PersistentPreRunE: init,
	}

	return projectCmd
}

func AddCmd(handler CobraHandler) *cobra.Command {
	addCmd := &cobra.Command{
		Use:     "add [flags] PROJECT_NAME",
		Aliases: []string{"a"},
		Short:   "Add a project",
		Example: "syringe project add my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return addCmd
}

func RemoveCmd(handler CobraHandler) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] PROJECT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove a project",
		Example: "syringe project remove my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return removeCmd
}

func RenameCmd(handler CobraHandler) *cobra.Command {
	renameCmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_PROJECT_NAME NEW_PROJECT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename a project",
		Example: "syringe project rename my_cool_project my_awesome_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	return renameCmd
}

func ListCmd(handler CobraHandler) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List projects",
		Args:    cobra.NoArgs,
		Example: "syringe project list",
		RunE:    handler,
	}

	return listCmd
}

func InitContext(cmd *cobra.Command, args []string) error {
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
