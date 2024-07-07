package cli

import (
	"crypto/rsa"
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
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func NewHandlerCLI(host string, port int, cmdOut io.Writer) pkg.CobraHandler {
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
			fmt.Println("unable to connect to agent, falling back to identity")
			authMethod, err = ssh.IdentityAuthMethod(identity)
			if err != nil {
				return err
			}
		} else {

			agentKeys, err := sshAgentClient.List()
			if err != nil {
				return fmt.Errorf("failed to list identities in ssh agent: %w", err)
			}

			// if the agent doesn't already contain the identity, then add it
			// iterate over the keys in the agent and compare them to ${identity}.pub to determine which key
			// pass that key's signer into the auth method further down (instead of using _all_ the keys' signers)
			var keyToUse *agent.Key
			for _, agentKey := range agentKeys {
				// compare each key to ${identity}.pub to determine if key provided by identity is in the agent
				publicKeyOnFile, err := os.ReadFile(fmt.Sprintf("%s.pub", identity))
				if err != nil {
					return fmt.Errorf("failed to read public key from filesystem: %w", err)
				}

				parsedPublicKeyOnFile, _, _, _, err := gossh.ParseAuthorizedKey(publicKeyOnFile)
				if err != nil {
					return fmt.Errorf("failed to parse public key from filesystem: %w", err)
				}

				if string(agentKey.Marshal()) == string(parsedPublicKeyOnFile.Marshal()) {
					keyToUse = agentKey
					break
				}
			}

			if keyToUse == nil {
				var privateKey *rsa.PrivateKey
				privateKey, err = ssh.GetPrivateKey(identity)
				if err != nil {
					if _, ok := err.(*gossh.PassphraseMissingError); !ok {
						return fmt.Errorf("failed to read private key from identity: %w", err)
					}

					cmd.Print(fmt.Sprintf("Enter passphrase for %s: ", identity))
					passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
					if err != nil {
						return fmt.Errorf("failed to read password: %w", err)
					}

					privateKey, err = ssh.GetPrivateKeyWithPassphrase(identity, string(passphrase))
					if err != nil {
						return fmt.Errorf("failed to read private key: %w", err)
					}
				}

				if err := sshAgentClient.Add(agent.AddedKey{PrivateKey: privateKey}); err != nil {
					fmt.Println(fmt.Errorf("failed to add key to agent: %w", err))
					fmt.Println("falling back to identity..???")
				}

				// TODO: how do we get the 'added key' back out
			}

			// if we fail to add to agent (i.e. it's still empty), then fallback to identity

			sshAgentClientSigners, err := sshAgentClient.Signers()
			if err != nil {
				return fmt.Errorf("failed to get signers from ssh client: %w", err)
			}

			var signersFunc = func() ([]gossh.Signer, error) {
				var signers []gossh.Signer

				for _, signer := range sshAgentClientSigners {

					publicKeyOnFile, err := os.ReadFile(fmt.Sprintf("%s.pub", identity))
					if err != nil {
						return nil, fmt.Errorf("failed to read public key from filesystem: %w", err)
					}

					parsedPublicKeyOnFile, _, _, _, err := gossh.ParseAuthorizedKey(publicKeyOnFile)
					if err != nil {
						return nil, fmt.Errorf("failed to parse public key from filesystem: %w", err)
					}

					if string(parsedPublicKeyOnFile.Marshal()) == string(signer.PublicKey().Marshal()) {
						fmt.Println("found signer...")
						signers = append(signers, signer)
					}
				}

				if len(signers) == 0 {
					return nil, errors.New("no valid signers in agent")
				}

				return signers, nil
			}

			authMethod, err = ssh.AgentAuthMethod(signersFunc)
			if err != nil {
				return err
			}
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
		if err := client.Run(sshcmd, cmdOut); err != nil {
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
