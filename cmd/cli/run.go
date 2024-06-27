package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gossh "golang.org/x/crypto/ssh"
)

func run(cmd *cobra.Command, args []string) error {
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
		sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
		if sshAuthSock == "" {
			return errors.New("SSH_AUTH_SOCK not set")
		}

		authMethod, err = ssh.AgentAuthMethod(sshAuthSock)
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

	var flags string

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		flags = fmt.Sprintf("%s --%s %s", flags, flag.Name, flag.Value)
	})

	scmd := []string{
		strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], " "),
		strings.Join(args, " "),
		flags,
	}

	if err := client.Run(strings.Join(scmd, " "), cmd.OutOrStdout()); err != nil {
		return err
	}

	return nil
}
