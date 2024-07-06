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

		currentUser, err := user.Current()
		if err != nil || currentUser.Username == "" {
			return fmt.Errorf("failed to determine username: %w", err)
		}

		identity, err := cmd.Flags().GetString("identity")
		if err != nil {
			return err
		}
		if identity == "" {
			return errors.New("no identity specified")
			// sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
			// if sshAuthSock == "" {
			// 	return errors.New("SSH_AUTH_SOCK not set")
			// }
			//
			// // TODO: get identity file path from config (if it exists); if it doesn't, then exit with error
			// // TODO: check that identity file from config matches the identity in the host??
			// authMethod, err = ssh.AgentAuthMethod(sshAuthSock)
			// if err != nil {
			// 	return err
			// }
		}

		if err := addIdentityToSSHConfig(identity); err != nil {
			return fmt.Errorf("failed to add/update identity in ssh config file: %w", err)
		}

		// TODO: check if provided identity is already available on ssh agent
		// if it isn't then add it
		// then use ssh agent rather than identityauthmethod
		// fallback to identityauthmethod

		authMethod, err = ssh.IdentityAuthMethod(identity)
		if err != nil {
			return err
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

func hostHasIdentity(host *ssh_config.Host, identity string) bool {
	var hasIdentity bool

	// check if host contains identity, if not then add to host
	var identityNodes []ssh_config.Node

	for _, n := range host.Nodes {
		switch t := n.(type) {
		case *ssh_config.KV:
			if strings.ToLower(t.Key) == "identityfile" {
				identityNodes = append(identityNodes, n)
			}
		default:
			continue
		}
	}

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

	return hasIdentity
}

func addIdentityToSSHConfig(identity string) error {
	var err error

	// get the host from ssh config, if it exists
	var f *os.File
	f, err = os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"), os.O_RDWR, 0600)
	if err != nil {
		f, err = os.OpenFile(filepath.Join("/etc", "ssh", "ssh_config"), os.O_RDWR, 0600)
		return errors.New("failed to open ssh config file")
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
		hasIdentity := hostHasIdentity(sshConfigHost, identity)

		if !hasIdentity {
			// add identity to existing host
			identityNode := &ssh_config.KV{
				Key:   "IdentityFile",
				Value: identity,
			}

			sshConfigHost.Nodes = append(sshConfigHost.Nodes, identityNode)
		}
	} else {
		// add new host with identity

		pattern, err := ssh_config.NewPattern(os.Getenv("APP_HOST"))
		if err != nil {
			return err
		}

		nodes := []ssh_config.Node{
			&ssh_config.KV{Key: "AddKeysToAgent", Value: "yes"},
			&ssh_config.KV{Key: "IgnoreUnknown", Value: "UseKeychain"},
			&ssh_config.KV{Key: "UseKeychain", Value: "yes"},
			&ssh_config.KV{Key: "IdentityFile", Value: identity},
		}

		sshConfigHost = &ssh_config.Host{
			Patterns: []*ssh_config.Pattern{pattern},
			Nodes:    nodes,
		}

		cfg.Hosts = append(cfg.Hosts, sshConfigHost)
	}

	// TODO: backup config before truncating?

	// FIXME: why would we write this every time the user runs the app
	// surely we want to avoid this if the key already exists in config

	// save sshConfigHost back to ~/.ssh/config
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("failed to zero out ssh config file: %w", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to get to beginning of config file: %w", err)
	}

	_, err = f.WriteString(cfg.String())
	if err != nil {
		return fmt.Errorf("failed to write ssh config to file: %w", err)
	}

	return nil
}
