package cmd

import (
	"context"
	"strings"

	client "github.com/nixpig/syringe.sh/cli/pkg/ssh"
	"github.com/spf13/cobra"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use: "syringe",
		// defers to commands defined on server, therefore these values should never be displayed
		Short:              "",
		Long:               "",
		Example:            "",
		DisableFlagParsing: true,
		Hidden:             true,
		SilenceUsage:       true,
		DisableSuggestions: true,
		RunE:               rootRunE,
	}

	ctx := context.Background()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

func rootRunE(cmd *cobra.Command, args []string) error {
	identity := "/home/nixpig/.ssh/id_rsa"

	output, err := client.SSHClient(
		"localhost",
		23234,
		identity,
		strings.Join(args, " "),
	)
	if err != nil {
		return err
	}

	cmd.Println(string(output))
	return nil
}
