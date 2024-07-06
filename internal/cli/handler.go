package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	gossh "golang.org/x/crypto/ssh"
)

func NewHandlerCLI(host string, port int, cmdOut io.Writer) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		var authMethod gossh.AuthMethod
		var identity string

		// don't care if errors, since will fallback to using ssh agent in case of empty identity
		identity, _ = cmd.Flags().GetString("identity")

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

			// TODO: get identity file path from config (if it exists); if it doesn't, then exit with error

			// TODO: check that identity file from config matches the identity in the host??

			authMethod, err = ssh.AgentAuthMethod(sshAuthSock)
			if err != nil {
				return err
			}
		}

		// get the host from ssh config, if it exists
		f, err := os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"), os.O_RDWR, 0600)
		if err != nil {
			return fmt.Errorf("unable to open ssh config: %w", err)
		}

		defer f.Close()

		cfg, err := ssh_config.Decode(f)
		if err != nil {
			return err
		}

		// get the closest matching host from ssh config file
		var sshConfigHost *ssh_config.Host

		for _, h := range cfg.Hosts {
			if h.Matches(os.Getenv("APP_HOST")) {
				if h.Matches("*") {
					continue
				}

				sshConfigHost = h
				break
			}
		}

		if sshConfigHost != nil {
			// check if host contains identity, if not then add to host
			var identityNodes []ssh_config.Node

			for _, n := range sshConfigHost.Nodes {
				switch t := n.(type) {
				case *ssh_config.KV:
					if strings.ToLower(t.Key) == "identityfile" {
						identityNodes = append(identityNodes, n)
					}
				default:
					continue
				}

			}

			fmt.Println(identityNodes)

			var hasIdentity bool

			for _, id := range identityNodes {
				ip := id.(*ssh_config.KV).Value

				if strings.HasPrefix(ip, "~/") {
					var home string
					home, err := os.UserHomeDir()
					if err != nil {
						home = os.Getenv("HOME")
					}

					ip = home + ip[1:]
				}

				if ip == identity {
					hasIdentity = true
				}
			}

			if !hasIdentity {
				// add identity to existing host
				fmt.Println("add to host...")
				identityNode := &ssh_config.KV{
					Key:   "IdentityFile",
					Value: identity,
				}

				sshConfigHost.Nodes = append(sshConfigHost.Nodes, identityNode)
			}

			fmt.Println("sshConfigHost: ", sshConfigHost)

			fmt.Println("host found and updated")

		} else {
			// add new host with identity
			fmt.Println("not found; creating new host")

			pattern, err := ssh_config.NewPattern(os.Getenv("APP_HOST"))
			if err != nil {
				return err
			}

			nodes := []ssh_config.Node{
				&ssh_config.KV{
					Key:   "AddKeysToAgent",
					Value: "yes",
				},
				&ssh_config.KV{
					Key:   "IgnoreUnknown",
					Value: "UseKeychain",
				},
				&ssh_config.KV{
					Key:   "UseKeychain",
					Value: "yes",
				},
				&ssh_config.KV{
					Key:   "IdentityFile",
					Value: identity,
				},
			}

			sshConfigHost = &ssh_config.Host{
				Patterns: []*ssh_config.Pattern{pattern},
				Nodes:    nodes,
			}

			cfg.Hosts = append(cfg.Hosts, sshConfigHost)
		}

		// TODO: backup config before truncating

		// save sshConfigHost back to ~/.ssh/config
		fmt.Println("host: ", host)
		if err := f.Truncate(0); err != nil {
			return fmt.Errorf("unable to zero out ssh config file: %w", err)
		}

		if _, err := f.Seek(0, 0); err != nil {
			return fmt.Errorf("unable to get to beginning of config file: %w", err)
		}

		_, err = f.WriteString(cfg.String())
		if err != nil {
			return fmt.Errorf("error writing ssh config to file: %w", err)
		}

		// TODO: prompt to add to agent??

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

		// TODO: probably need to use a channel to close the client once done
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
