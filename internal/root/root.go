package root

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nixpig/syringe.sh/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

func New(ctx context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: config.Version,
		Use:     "syringe",
		Short:   "Distributed database-per-user encrypted secrets management over SSH protocol.",
		Long: `Distributed database-per-user encrypted secrets management over SSH protocol.

SSH is a protocol that...

How syringe.sh works...

All secrets are encrypted... Secrets are encrypted on your machine before being sent to... Nobody else, including us, can decrypt and read your secrets.

Encryption is tied to your SSH key. If you lose your SSH key, that's it... You can upload multiple SSH keys...

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
    syringe inject -p my_cool_project -e dev -- startserver

  For more examples, go to https://syringe.sh/examples`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			v := viper.New()

			if err := initialiseConfig(cmd, v); err != nil {
				return err
			}

			bindFlags(cmd, v)

			fmt.Println("viper: ", v.GetString("identity"))

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
		"Path to SSH key (optional).\nIf not provided, SSH agent is used and syringe.sh host must be configured in SSH config.",
	)

	rootCmd.SetContext(ctx)

	return rootCmd
}

func initialiseConfig(v *viper.Viper) error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	syringeConfigDir := filepath.Join(userConfigDir, "syringe")

	if err := os.MkdirAll(syringeConfigDir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(
		filepath.Join(syringeConfigDir, "settings"),
		os.O_RDWR|os.O_CREATE,
		0666,
	)
	if err != nil {
		return err
	}
	f.Close()

	v.SetConfigFile(filepath.Join(
		syringeConfigDir,
		"settings",
	))

	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return nil
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
