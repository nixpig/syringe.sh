package secret

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdSecret() *cobra.Command {
	cmdSecret := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "Manage secrets",
		Long:    "Manage your secrets",
	}

	return cmdSecret
}

func NewCmdSecretSet(handler pkg.CobraHandler) *cobra.Command {
	cmdSet := &cobra.Command{
		Use:     "set [flags] SECRET_KEY SECRET_VALUE",
		Aliases: []string{"s"},
		Short:   "Set a secret",
		Example: "syringe secret set -p my_cool_project -e local AWS_ACCESS_KEY_ID AKIAIOSFODNN7EXAMPLE",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	cmdSet.Flags().StringP("project", "p", "", "Project to use")
	cmdSet.MarkFlagRequired("project")

	cmdSet.Flags().StringP("environment", "e", "", "Environment to use")
	cmdSet.MarkFlagRequired("environment")

	cmdSet.Flags().BoolP("secret", "s", false, "Is secret?")

	return cmdSet
}

func NewCmdSecretGet(handler pkg.CobraHandler) *cobra.Command {
	cmdGet := &cobra.Command{
		Use:     "get [flags] SECRET_KEY",
		Aliases: []string{"g"},
		Short:   "Get a secret",
		Example: "syringe get -p my_cool_project -e local AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	cmdGet.Flags().StringP("project", "p", "", "Project")
	cmdGet.MarkFlagRequired("project")

	cmdGet.Flags().StringP("environment", "e", "", "Environment")
	cmdGet.MarkFlagRequired("environment")

	cmdGet.Flags().BoolP("secret", "s", false, "Is secret?")

	return cmdGet
}

func NewCmdSecretList(handler pkg.CobraHandler) *cobra.Command {
	cmdList := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List all secrets",
		Example: "syringe secret list -p my_cool_project -e staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    handler,
	}

	cmdList.Flags().StringP("project", "p", "", "Project name")
	cmdList.MarkFlagRequired("project")

	cmdList.Flags().StringP("environment", "e", "", "Environment name")
	cmdList.MarkFlagRequired("environment")

	return cmdList
}

func NewCmdSecretRemove(handler pkg.CobraHandler) *cobra.Command {
	cmdRemove := &cobra.Command{
		Use:     "remove [flags] SECRET_KEY",
		Aliases: []string{"r"},
		Short:   "Remove a secret",
		Example: "syringe secret remove -p my_cool_project -e staging AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	cmdRemove.Flags().StringP("project", "p", "", "Project name")
	cmdRemove.MarkFlagRequired("project")

	cmdRemove.Flags().StringP("environment", "e", "", "Environment name")
	cmdRemove.MarkFlagRequired("environment")

	return cmdRemove
}
