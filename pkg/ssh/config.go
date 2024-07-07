package ssh

import (
	"fmt"
	"os"
	"strings"

	"github.com/kevinburke/ssh_config"
)

type Config struct {
	cfg  *ssh_config.Config
	file *os.File
}

func NewConfig(rw *os.File) (*Config, error) {
	cfg, err := ssh_config.Decode(rw)
	if err != nil {
		return nil, err
	}

	return &Config{
		cfg:  cfg,
		file: rw,
	}, nil
}

func (c *Config) Write() error {

	// TODO: backup config before truncating?

	if err := c.file.Truncate(0); err != nil {
		return fmt.Errorf("failed to zero out ssh config file: %w", err)
	}

	if _, err := c.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to get to beginning of config file: %w", err)
	}

	if _, err := c.file.WriteString(c.cfg.String()); err != nil {
		return fmt.Errorf("failed to write ssh config to file: %w", err)
	}

	return nil
}

func (c *Config) GetHost(pattern string, allowWildcard bool) *ssh_config.Host {
	var sshConfigHost *ssh_config.Host

	for _, h := range c.cfg.Hosts {
		if h.Matches(pattern) {
			if !allowWildcard && h.Matches("*") {
				continue
			}

			sshConfigHost = h
			break
		}
	}

	return sshConfigHost
}

func (c *Config) HostHasIdentity(host *ssh_config.Host, identity string) bool {
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

func (c *Config) AddIdentityToHost(host *ssh_config.Host, identity string) {
	identityNode := &ssh_config.KV{
		Key:   "IdentityFile",
		Value: identity,
	}

	host.Nodes = append(host.Nodes, identityNode)
}

func (c *Config) AddHost(name, identity string) error {
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

	c.cfg.Hosts = append(c.cfg.Hosts, sshConfigHost)

	return nil
}
