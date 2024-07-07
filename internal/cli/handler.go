package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func NewHandlerCLI(host string, port int, out io.Writer) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		var authMethod gossh.AuthMethod

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

		sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
		if sshAuthSock == "" {
			return errors.New("SSH_AUTH_SOCK not set")
		}

		sshAgentClient, err := ssh.NewSSHAgentClient(sshAuthSock)
		if err != nil {
			cmd.Println("unable to connect to agent, falling back to identity")

			signer, err := ssh.GetSigner(identity, cmd.OutOrStderr())
			if err != nil {
				return err
			}

			authMethod = gossh.PublicKeys(signer)

		} else {
			agentKeys, err := sshAgentClient.List()
			if err != nil {
				return fmt.Errorf("failed to get identities from ssh agent: %w", err)
			}

			publicKey, err := ssh.GetPublicKey(fmt.Sprintf("%s.pub", identity))
			if err != nil {
				return fmt.Errorf("failed to get public key: %w", err)
			}

			// if the agent doesn't already contain the identity, then add it
			if i := slices.IndexFunc(agentKeys, func(agentKey *agent.Key) bool {
				return string(agentKey.Marshal()) == string(publicKey.Marshal())
			}); i == -1 {
				privateKey, err := ssh.GetPrivateKey(identity, cmd.OutOrStderr())
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}

				if err := sshAgentClient.Add(agent.AddedKey{PrivateKey: privateKey}); err != nil {
					return fmt.Errorf("failed to add key to agent: %w", err)
				}
			}

			sshAgentClientSigners, err := sshAgentClient.Signers()
			if err != nil {
				return fmt.Errorf("failed to get signers from ssh client: %w", err)
			}

			authMethod = gossh.PublicKeysCallback(
				// use only signer for the specified identity key
				newSignersFunc(publicKey, sshAgentClientSigners),
			)
		}

		if err := addIdentityToSSHConfig(identity); err != nil {
			return fmt.Errorf("failed to add/update identity in ssh config file: %w", err)
		}

		// TODO: pull this out and pass in as a dependency
		client, err := ssh.NewSSHClient(
			host,
			port,
			currentUser.Username,
			authMethod,
		)
		if err != nil {
			return err
		}

		defer client.Close()

		sshcmd := buildCommand(cmd, args)

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

func addIdentityToSSHConfig(identity string) error {
	var err error

	var f *os.File
	f, err = os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"), os.O_RDWR, 0600)
	if err != nil {
		f, err = os.OpenFile(filepath.Join("/etc", "ssh", "ssh_config"), os.O_RDWR, 0600)
		if err != nil {
			return errors.New("failed to open ssh config file")
		}
	}

	defer f.Close()

	config, err := ssh.NewConfig(f)
	if err != nil {
		return err
	}

	sshConfigHost := config.GetHost(os.Getenv("APP_HOST"), false)
	if sshConfigHost == nil {
		config.AddHost(os.Getenv("APP_HOST"), identity)
		if err := config.Write(); err != nil {
			return err
		}
	} else {
		if !config.HostHasIdentity(sshConfigHost, identity) {
			config.AddIdentityToHost(sshConfigHost, identity)
			if err := config.Write(); err != nil {
				return err
			}
		}
	}

	return nil
}

func newSignersFunc(publicKey gossh.PublicKey, agentSigners []gossh.Signer) func() ([]gossh.Signer, error) {
	return func() ([]gossh.Signer, error) {
		var signers []gossh.Signer

		for _, signer := range agentSigners {
			if string(publicKey.Marshal()) == string(signer.PublicKey().Marshal()) {
				signers = append(signers, signer)
			}
		}

		if len(signers) == 0 {
			return nil, errors.New("no valid signers in agent")
		}

		return signers, nil
	}
}
