package environment

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdEnvironment() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"e"},
		Short:   "Manage environments",
		Long:    "Manage your environments.",
	}

	return cmd
}

func NewCmdEnvironmentAdd(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [flags] ENVIRONMENT_NAME",
		Aliases: []string{"a"},
		Short:   "Add an environment",
		Example: "syringe environment add -p my_cool_project local",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdEnvironmentRemove(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [flags] ENVIRONMENT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove an environment",
		Example: "syringe environment remove -p my_cool_project staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdEnvironmentRename(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_ENVIRONMENT_NAME NEW_ENVIRONMENT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename an environment",
		Example: "syringe environment rename -p my_cool_project staging prod",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdEnvironmentList(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List environments",
		Example: "syringe environment list -p my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("project", "p", "", "Project name")
	cmd.MarkFlagRequired("project")
}
