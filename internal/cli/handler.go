package cli

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"slices"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

func NewHandlerCLI(host string, port int, out io.Writer) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		var authMethod gossh.AuthMethod
		var privateKey *rsa.PrivateKey

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

		publicKey, err := ssh.GetPublicKey(fmt.Sprintf("%s.pub", identity))
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}

		sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
		if sshAuthSock == "" {
			return errors.New("SSH_AUTH_SOCK not set")
		}

		sshAgentClient, err := ssh.NewSSHAgentClient(sshAuthSock)
		if err != nil {
			cmd.Println("unable to connect to agent, falling back to identity")

			signer, err := ssh.GetSigner(identity, cmd.OutOrStderr(), term.ReadPassword)
			if err != nil {
				return err
			}

			authMethod = gossh.PublicKeys(signer)

		} else {
			agentKeys, err := sshAgentClient.List()
			if err != nil {
				return fmt.Errorf("failed to get identities from ssh agent: %w", err)
			}

			// if the agent doesn't already contain the identity, then add it
			if i := slices.IndexFunc(agentKeys, func(agentKey *agent.Key) bool {
				return string(agentKey.Marshal()) == string(publicKey.Marshal())
			}); i == -1 {
				privateKey, err = ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
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
				ssh.NewSignersFunc(publicKey, sshAgentClientSigners),
			)
		}

		configFile, err := ssh.ConfigFile()
		if err != nil {
			return err
		}

		defer configFile.Close()

		if err := ssh.AddIdentityToSSHConfig(identity, configFile); err != nil {
			return fmt.Errorf("failed to add or update identity in ssh config file: %w", err)
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

		if cmd.CalledAs() == "inject" {
			sshcmd := buildCommand(cmd, args)

			privateKey, err = ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
			if err != nil {
				return fmt.Errorf("failed to read private key: %w", err)
			}

			dout := DecryptedOutput{
				out:        out,
				decrypt:    injectDecryptor,
				privateKey: privateKey,
			}

			if err := client.Run(sshcmd, dout); err != nil {
				return err
			}

			return nil
		}

		if cmd.Parent().Use == "secret" {
			switch cmd.CalledAs() {

			case "list":
				sshcmd := buildCommand(cmd, args)

				privateKey, err = ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}

				dout := DecryptedOutput{
					out:        out,
					decrypt:    listDecryptor,
					privateKey: privateKey,
				}

				if err := client.Run(sshcmd, dout); err != nil {
					return err
				}

				return nil

			case "get":
				sshcmd := buildCommand(cmd, args)

				privateKey, err = ssh.GetPrivateKey(identity, cmd.OutOrStderr(), term.ReadPassword)
				if err != nil {
					return fmt.Errorf("failed to read private key: %w", err)
				}

				dout := DecryptedOutput{
					out:        out,
					decrypt:    ssh.Decrypt,
					privateKey: privateKey,
				}

				if err := client.Run(sshcmd, dout); err != nil {
					return err
				}

				return nil
			}
		}

		sshcmd := buildCommand(cmd, args)
		if err := client.Run(sshcmd, out); err != nil {
			return err
		}

		return nil
	}
}

func listDecryptor(cypherText string, privateKey *rsa.PrivateKey) (string, error) {
	var err error

	lines := strings.Split(cypherText, "\n")
	for i, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		parts[1], err = ssh.Decrypt(parts[1], privateKey)
		if err != nil {
			return "", err
		}

		lines[i] = strings.Join(parts, "=")
	}

	return strings.Join(lines, "\n"), nil
}

func injectDecryptor(cypherText string, privateKey *rsa.PrivateKey) (string, error) {
	var err error

	lines := strings.Split(cypherText, " ")
	for i, l := range lines {
		parts := strings.SplitN(l, "=", 2)
		parts[1], err = ssh.Decrypt(parts[1], privateKey)
		if err != nil {
			return "", err
		}

		lines[i] = strings.Join(parts, "=")
	}

	return strings.Join(lines, " "), nil
}

type Decryptor func(cypherText string, privateKey *rsa.PrivateKey) (string, error)

type DecryptedOutput struct {
	out        io.Writer
	privateKey *rsa.PrivateKey
	decrypt    Decryptor
}

func (d DecryptedOutput) Write(p []byte) (int, error) {
	decrypted, err := d.decrypt(string(p), d.privateKey)
	if err != nil {
		return 0, err
	}

	b, err := d.out.Write([]byte(decrypted))
	if err != nil {
		return b, err
	}

	if len([]byte(decrypted)) != b {
		return b, io.ErrShortWrite
	}

	return b, nil
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
