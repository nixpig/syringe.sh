package secret

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdSecret() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "Manage secrets",
		Long:    "Manage your secrets",
	}

	return secretCmd
}

func NewCmdSecretSet(handler pkg.CobraHandler) *cobra.Command {
	setCmd := &cobra.Command{
		Use:     "set [flags] SECRET_KEY SECRET_VALUE",
		Aliases: []string{"s"},
		Short:   "Set a secret",
		Example: "syringe secret set -p my_cool_project -e local AWS_ACCESS_KEY_ID AKIAIOSFODNN7EXAMPLE",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	setCmd.Flags().StringP("project", "p", "", "Project to use")
	setCmd.MarkFlagRequired("project")

	setCmd.Flags().StringP("environment", "e", "", "Environment to use")
	setCmd.MarkFlagRequired("environment")

	setCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return setCmd
}

func NewCmdSecretGet(handler pkg.CobraHandler) *cobra.Command {
	getCmd := &cobra.Command{
		Use:     "get [flags] SECRET_KEY",
		Aliases: []string{"g"},
		Short:   "Get a secret",
		Example: "syringe get -p my_cool_project -e local AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	getCmd.Flags().StringP("project", "p", "", "Project")
	getCmd.MarkFlagRequired("project")

	getCmd.Flags().StringP("environment", "e", "", "Environment")
	getCmd.MarkFlagRequired("environment")

	getCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return getCmd
}

func NewCmdSecretList(handler pkg.CobraHandler) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List all secrets",
		Example: "syringe secret list -p my_cool_project -e staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    handler,
	}

	listCmd.Flags().StringP("project", "p", "", "Project name")
	listCmd.MarkFlagRequired("project")

	listCmd.Flags().StringP("environment", "e", "", "Environment name")
	listCmd.MarkFlagRequired("environment")

	return listCmd
}

func NewCmdSecretRemove(handler pkg.CobraHandler) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] SECRET_KEY",
		Aliases: []string{"r"},
		Short:   "Remove a secret",
		Example: "syringe secret remove -p my_cool_project -e staging AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	removeCmd.Flags().StringP("project", "p", "", "Project name")
	removeCmd.MarkFlagRequired("project")

	removeCmd.Flags().StringP("environment", "e", "", "Environment name")
	removeCmd.MarkFlagRequired("environment")

	return removeCmd
}
