package root

import (
	"github.com/nixpig/syringe.sh/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

func New(ctx context.Context, v *viper.Viper) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: config.Version,
		Use:     "syringe",
		Short:   "Distributed database-per-user encrypted secrets management over SSH protocol.",
		Long: `Distributed database-per-user encrypted secrets management over SSH protocol.

SSH (Secure Shell) is a cryptographic network protocol for secure communication between computers over an unsecured network that uses keys for secure authentication. If you've ever ssh'd into a remote machine or used CLI tools like git then you've used SSH.

syringe.sh uses SSH as the protocol for communication between the client (your machine) and the server (in the cloud).

Your public key is uploaded to the server. Your private key is then used to authenticate when you connect.

Secrets are encrypted locally using your key before being sent to the server and stored in a separate database tied to your SSH key.

Secrets can only be decrypted locally using your private key. Without your private key, nobody can decrypt and read your secrets. It's important you don't lose this, else your secrets will be lost forever.

Supported key formats:
  ✓ RSA
  ...more soon!`,

		Example: `  • Add a project
    syringe project add my_cool_project

  • Add an environment
    syringe environment add -p my_cool_project dev

  • Add a secret
    syringe secret set -p my_cool_project -e dev SECRET_KEY secret_value

  • List secrets
    syringe secret list -p my_cool_project -e dev

  • Inject secrets into command
    syringe secret inject -p my_cool_project -e dev -- startserver

  For more examples, go to https://syringe.sh/examples`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if v != nil {
				bindFlags(cmd, v)
			}

			return nil
		},
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

	rootCmd.PersistentFlags().StringP(
		"identity",
		"i",
		"",
		`Path to SSH key (optional).
If not provided, 'identity' is read from settings file.
If no identity specified as flag or in settings file, SSH agent is used and syringe.sh host must be configured in SSH config.`,
	)
	rootCmd.SetContext(ctx)

	return rootCmd
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	if err := v.BindPFlag("identity", cmd.Flags().Lookup("identity")); err != nil {
		return err
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if v.IsSet(f.Name) {
			cmd.Flags().Set(f.Name, v.GetString(f.Name))
		}
	})

	return nil
}
