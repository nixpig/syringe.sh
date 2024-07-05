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
		Long: `🗝 Distributed database-per-user encrypted secrets management over SSH protocol.

SSH is a protocol that...

How syringe.sh works...

All secrets are encrypted... Secrets are encrypted on your machine before being sent to... Nobody else, including us, can decrypt and read your secrets.

Encryption is tied to your SSH key. If you lose your SSH key, that's it... You can upload multiple SSH keys...

Supported key formats:
  ✓ RSA
  ✓ OpenSSH
  ✗ SomeOther`,

		Example: `  • Add a project
    syringe project add my_cool_project

  • Add an environment
    syringe environment add -p my_cool_project dev

  • Add a secret
    syringe secret set -p my_cool_project -e dev SECRET_KEY secret_value

  • List secrets
    syringe secret list -p my_cool_project -e dev

  • Inject secrets into command
    syringe inject -p my_cool_project -e dev -- startserver

  For more examples, go to https://syringe.sh/examples`,
	}

	additionalHelp := `
For more help on how to use syringe.sh, go to https://syringe.sh/help`

	warningMessage :=
		"\n\n\033[31m⚠ WARNING\033[0m\n" +
			"  \033[33m~\033[0m This software is currently in development.\n" +
			"  \033[33m~\033[0m Many of the features may not work as documented, or even at all.\n" +
			"  \033[33m~\033[0m You probably (almost certainly!) don't want to use this software just yet.\033[0m\n"

	rootCmd.SetHelpTemplate(
		rootCmd.HelpTemplate() +
			additionalHelp +
			warningMessage,
	)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SetContext(ctx)

	return rootCmd
}
