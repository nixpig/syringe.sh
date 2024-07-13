package ssh

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
)

func AddIdentityToSSHConfig(identity string, f *os.File) error {
	var err error

	cfg, err := configFromFile(f)
	if err != nil {
		return err
	}

	sshConfigHost, err := hostFromConfig(cfg, os.Getenv("APP_HOST"), false)
	if err == nil {
		if hostHasIdentity(sshConfigHost, identity) {
			return nil
		}

		addIdentityToHost(sshConfigHost, identity)
	} else {
		if err := addHost(
			cfg,
			os.Getenv("APP_HOST"),
			identity,
		); err != nil {
			return err
		}
	}

	if err := writeConfig(cfg, f); err != nil {
		return fmt.Errorf("failed to write ssh config to file: %w", err)
	}

	return nil
}

func ConfigFile() (*os.File, error) {
	var f *os.File
	var err error

	f, err = os.OpenFile(filepath.Join(os.Getenv("HOME"), ".ssh", "config"), os.O_RDWR, 0600)
	if err != nil {
		f, err = os.OpenFile(filepath.Join("/etc", "ssh", "ssh_config"), os.O_RDWR, 0600)
		if err != nil {
			return nil, errors.New("failed to open ssh config file")
		}
	}

	return f, nil
}

func writeConfig(cfg *ssh_config.Config, f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("failed to zero out ssh config file: %w", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to get to beginning of config file: %w", err)
	}

	if _, err := f.WriteString(cfg.String()); err != nil {
		return fmt.Errorf("failed to write ssh config to file: %w", err)
	}

	return nil
}

func configFromFile(f *os.File) (*ssh_config.Config, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to get to beginning of config file: %w", err)
	}

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func hostFromConfig(
	cfg *ssh_config.Config,
	pattern string,
	allowWildcard bool,
) (*ssh_config.Host, error) {
	var sshConfigHost *ssh_config.Host

	for _, h := range cfg.Hosts {
		if h.Matches(pattern) {
			if !allowWildcard && h.Matches("*") {
				continue
			}

			sshConfigHost = h
			break
		}
	}

	if sshConfigHost != nil {
		return sshConfigHost, nil
	}

	return nil, errors.New("host not in config")
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

func addIdentityToHost(host *ssh_config.Host, identity string) {
	identityNode := &ssh_config.KV{
		Key:   "IdentityFile",
		Value: identity,
	}

	host.Nodes = append(host.Nodes, identityNode)
}

func addHost(cfg *ssh_config.Config, name, identity string) error {
	pattern, err := ssh_config.NewPattern(name)
	if err != nil {
		return err
	}

	nodes := []ssh_config.Node{
		&ssh_config.KV{Key: "AddKeysToAgent", Value: "yes"},
		&ssh_config.KV{Key: "IgnoreUnknown", Value: "UseKeychain"},
		&ssh_config.KV{Key: "UseKeychain", Value: "yes"},
		&ssh_config.KV{Key: "IdentityFile", Value: identity},
	}

	sshConfigHost := &ssh_config.Host{
		Patterns: []*ssh_config.Pattern{pattern},
		Nodes:    nodes,
	}

	cfg.Hosts = append(cfg.Hosts, sshConfigHost)

	return nil
}
