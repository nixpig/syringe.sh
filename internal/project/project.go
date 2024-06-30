package project

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdProject() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Short:   "Manage projects",
		Long:    "Mange your projects.",
	}

	return cmd
}

func NewCmdProjectAdd(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [flags] PROJECT_NAME",
		Aliases: []string{"a"},
		Short:   "Add a project",
		Example: "syringe project add my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return cmd
}

func NewCmdProjectRemove(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [flags] PROJECT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove a project",
		Example: "syringe project remove my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return cmd
}

func NewCmdProjectRename(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_PROJECT_NAME NEW_PROJECT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename a project",
		Example: "syringe project rename my_cool_project my_awesome_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	return cmd
}

func NewCmdProjectList(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List projects",
		Args:    cobra.NoArgs,
		Example: "syringe project list",
		RunE:    handler,
	}

	return cmd
}
