package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func NewHandlerCLI(
	hostname string,
	port int,
	out io.Writer,

	// TODO: really don't like this passing here
	newSSHClient func(
		hostname string,
		port int,
		username string,
		authMethod gossh.AuthMethod,
		knownHosts string,
	) (*ssh.SSHClient, error),
) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		currentUser, err := user.Current()
		if err != nil || currentUser.Username == "" {
			return fmt.Errorf("failed to determine username: %w", err)
		}

		identity, err := cmd.Flags().GetString("identity")
		if err != nil {
			return err
		}

		if identity == "" {
			return errors.New("no identity provided")
		}

		authMethod, err := ssh.AuthMethod(identity, cmd.OutOrStdout())
		if err != nil {
			return err
		}

		configFile, err := ssh.ConfigFile()
		if err != nil {
			return err
		}

		defer configFile.Close()

		if err := ssh.AddIdentityToSSHConfig(identity, hostname, configFile); err != nil {
			return fmt.Errorf("failed to add or update identity in ssh config file: %w", err)
		}

		client, err := newSSHClient(
			hostname,
			port,
			currentUser.Username,
			authMethod,
			filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"),
		)
		if err != nil {
			return err
		}

		defer client.Close()

		sshcmd := buildCommand(cmd, args)

		// TODO: don't like this nested if/switch stuff

		if cmd.Parent().Use == "secret" {
			switch cmd.CalledAs() {

			case "inject":
				privateKey, err := ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}

				out = NewInjectResponseParser(
					out,
					privateKey,
					ssh.Decrypt,
				)

			case "list":
				privateKey, err := ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}

				out = NewListResponseParser(
					out,
					privateKey,
					ssh.Decrypt,
				)

			case "get":
				privateKey, err := ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}

				out = NewGetResponseParser(
					out,
					privateKey,
					ssh.Decrypt,
				)
			}
		}

		if err := client.Run(sshcmd, out); err != nil {
			return err
		}

		return nil
	}
}

func buildCommand(cmd *cobra.Command, args []string) string {
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

	return strings.Join(scmd, " ")
}
