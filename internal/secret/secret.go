package secret

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdSecret() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "Manage secrets",
		Long:    "Manage your secrets",
	}

	return cmd
}

func NewCmdSecretSet(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set [flags] SECRET_KEY SECRET_VALUE",
		Aliases: []string{"s"},
		Short:   "Set a secret",
		Example: "syringe secret set -p my_cool_project -e local AWS_ACCESS_KEY_ID AKIAIOSFODNN7EXAMPLE",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdSecretGet(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get [flags] SECRET_KEY",
		Aliases: []string{"g"},
		Short:   "Get a secret",
		Example: "syringe get -p my_cool_project -e local AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdSecretList(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List all secrets",
		Example: "syringe secret list -p my_cool_project -e staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdSecretRemove(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [flags] SECRET_KEY",
		Aliases: []string{"r"},
		Short:   "Remove a secret",
		Example: "syringe secret remove -p my_cool_project -e staging AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	addFlags(cmd)

	return cmd
}

func NewCmdSecretInject(handler pkg.CobraHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inject [flags] -- SUBCOMMAND",
		Aliases: []string{"i"},
		Short:   "Inject secrets",
		Long:    "Inject secrets into the specified subcommand.",
		Example: `  • Inject secrets from 'dev' environment in 'my_cool_project' project into 'startserver' command
    syringe secret inject -p my_cool_project -e dev -- startserver`,
		Args: cobra.MinimumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: handler,
	}

	addFlags(cmd)

	return cmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("project", "p", "", "Project name")
	cmd.MarkFlagRequired("project")

	cmd.Flags().StringP("environment", "e", "", "Environment name")
	cmd.MarkFlagRequired("environment")
}
