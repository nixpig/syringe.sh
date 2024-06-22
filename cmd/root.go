package cmd

import (
	"context"
	"os"
	"os/user"
	"strings"

	"github.com/nixpig/syringe.sh/cli/pkg/ssh"

	"github.com/spf13/cobra"
	gossh "golang.org/x/crypto/ssh"
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
	var err error
	var authMethod gossh.AuthMethod

	// identity := "/home/nixpig/.ssh/id_rsa"
	identity := ""

	currentUser, err := user.Current()
	if err != nil || currentUser.Username == "" {
		return err
	}

	if identity != "" {
		authMethod, err = ssh.IdentityAuthMethod(identity)
		if err != nil {
			return err
		}
	} else {
		authMethod, err = ssh.AgentAuthMethod(os.Getenv("SSH_AUTH_SOCK"))
		if err != nil {
			return err
		}
	}

	client, err := ssh.NewSSHClient(
		"localhost",
		23234,
		currentUser.Username,
		authMethod,
	)
	if err != nil {
		return err
	}

	defer client.Close()

	client.Run(strings.Join(args, " "), cmd.OutOrStdout())

	return nil
}
