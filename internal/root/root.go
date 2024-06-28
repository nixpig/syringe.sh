package root

import (
	"github.com/nixpig/syringe.sh/config"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func New(ctx context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: config.Version,
		Use:     "syringe",
		Short:   "🔐 Distributed database-per-user encrypted secrets management over SSH protocol.",
		Long:    "🔐 Distributed database-per-user encrypted secrets management over SSH protocol.",

		Example: `  # Add a project
  syringe project add my_cool_project

  # Add an environment
  syringe environment add -p my_cool_project dev

  # Add a secret
  syringe secret set -p my_cool_project -e dev SECRET_KEY secret_value

  # List secrets
  syringe secret list -p my_cool_project -e dev

  # Inject secrets into command
  syringe inject -p my_cool_project -e dev -- startserver

  For more examples, go to https://syringe.sh/examples`,
	}

	additionalHelp := `
For more help on how to use syringe.sh, go to https://syringe.sh/help

`

	rootCmd.SetHelpTemplate(rootCmd.HelpTemplate() + additionalHelp)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SetContext(ctx)

	return rootCmd
}
