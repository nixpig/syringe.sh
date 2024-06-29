package project

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdProject() *cobra.Command {
	cmdProject := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Short:   "Manage projects",
		// PreRunE: auth.PreRunE,
	}

	return cmdProject
}

func NewCmdProjectAdd(handler pkg.CobraHandler) *cobra.Command {
	cmdAdd := &cobra.Command{
		Use:     "add [flags] PROJECT_NAME",
		Aliases: []string{"a"},
		Short:   "Add a project",
		Example: "syringe project add my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return cmdAdd
}

func NewCmdProjectRemove(handler pkg.CobraHandler) *cobra.Command {
	cmdRemove := &cobra.Command{
		Use:     "remove [flags] PROJECT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove a project",
		Example: "syringe project remove my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return cmdRemove
}

func NewCmdProjectRename(handler pkg.CobraHandler) *cobra.Command {
	cmdRename := &cobra.Command{
		Use:     "rename [flags] CURRENT_PROJECT_NAME NEW_PROJECT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename a project",
		Example: "syringe project rename my_cool_project my_awesome_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	return cmdRename
}

func NewCmdProjectList(handler pkg.CobraHandler) *cobra.Command {
	cmdList := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List projects",
		Args:    cobra.NoArgs,
		Example: "syringe project list",
		RunE:    handler,
	}

	return cmdList
}
