package project

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdProject() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Short:   "Manage projects",
	}

	return projectCmd
}

func NewCmdProjectAdd(handler pkg.CobraHandler) *cobra.Command {
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

func NewCmdProjectRemove(handler pkg.CobraHandler) *cobra.Command {
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

func NewCmdProjectRename(handler pkg.CobraHandler) *cobra.Command {
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

func NewCmdProjectList(handler pkg.CobraHandler) *cobra.Command {
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
