package inject

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdInject(handler pkg.CobraHandler) *cobra.Command {
	injectCmd := &cobra.Command{
		Use:     "inject [flags] -- SUBCOMMAND",
		Aliases: []string{"i"},
		Short:   "Inject secrets",
		Long:    "Inject secrets into the specified subcommand.",
		Example: `  # Inject secrets from 'dev' environment in 'my_cool_project' project into 'startserver' command
    syringe inject -p my_cool_project -e dev -- startserver`,
		Args: cobra.MinimumNArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: handler,
	}

	injectCmd.Flags().StringP("project", "p", "", "Project name")
	injectCmd.MarkFlagRequired("project")

	injectCmd.Flags().StringP("environment", "e", "", "Environment name")
	injectCmd.MarkFlagRequired("environment")

	return injectCmd
}
