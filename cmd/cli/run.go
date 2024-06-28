package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gossh "golang.org/x/crypto/ssh"
)

func newCliHandler(cmdOut io.Writer) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		var authMethod gossh.AuthMethod

		// don't care if errors, since will fallback to using ssh agent in case of empty identity
		identity, _ := cmd.Flags().GetString("identity")

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
			if flag.Name == "identity" {
				return
			}

			flags = fmt.Sprintf("%s --%s %s", flags, flag.Name, flag.Value)
		})

		scmd := []string{
			strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], " "),
			strings.Join(args, " "),
			flags,
		}

		if err := client.Run(strings.Join(scmd, " "), cmdOut); err != nil {
			return err
		}

		return nil
	}
}
