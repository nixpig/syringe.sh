package environment

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdEnvironment() *cobra.Command {
	cmdEnvironment := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"e"},
		Short:   "Manage environments",
		Long:    "Manage your environments.",
	}

	cmdEnvironment.PersistentFlags().StringP("project", "p", "", "Project name")
	cmdEnvironment.MarkFlagRequired("project")

	return cmdEnvironment
}

func NewCmdEnvironmentAdd(handler pkg.CobraHandler) *cobra.Command {
	addCmd := &cobra.Command{
		Use:     "add [flags] ENVIRONMENT_NAME",
		Aliases: []string{"a"},
		Short:   "Add an environment",
		Example: "syringe environment add -p my_cool_project local",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return addCmd
}

func NewCmdEnvironmentRemove(handler pkg.CobraHandler) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] ENVIRONMENT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove an environment",
		Example: "syringe environment remove -p my_cool_project staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	return removeCmd
}

func NewCmdEnvironmentRename(handler pkg.CobraHandler) *cobra.Command {
	renameCmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_ENVIRONMENT_NAME NEW_ENVIRONMENT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename an environment",
		Example: "syringe environment rename -p my_cool_project staging prod",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	return renameCmd
}

func NewCmdEnvironmentList(handler pkg.CobraHandler) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List environments",
		Example: "syringe environment list -p my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    handler,
	}

	return listCmd
}