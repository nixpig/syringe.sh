package ssh

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/skeema/knownhosts"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const Client = "SSH-2.0-Syringe"

type SSHClient struct {
	client *gossh.Client
}

func (s *SSHClient) Close() error {
	if err := s.client.Close(); err != nil {
		return err
	}

	return nil
}

func (s *SSHClient) Run(cmd string, out io.Writer) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}

	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("%s", output)
	}

	return nil
}

func NewSSHClient(
	host string,
	port int,
	username string,
	authMethod gossh.AuthMethod,
	knownHosts string,
) (*SSHClient, error) {
	sshConfig := &gossh.ClientConfig{
		User:          username,
		ClientVersion: Client,
		Auth:          []gossh.AuthMethod{authMethod},

		HostKeyCallback: gossh.HostKeyCallback(
			func(hostname string, remote net.Addr, key gossh.PublicKey) error {
				kh, err := knownhosts.New(knownHosts)
				if err != nil {
					return fmt.Errorf("failed to open knownhosts file: %w", err)
				}

				err = kh(fmt.Sprintf("%s:%d", host, port), remote, key)

				if knownhosts.IsHostKeyChanged(err) {
					return fmt.Errorf("remote host identification has changed which may indicate a MITM attack: %w", err)
				}

				if knownhosts.IsHostUnknown(err) {
					khHandle, err := os.OpenFile(knownHosts, os.O_APPEND|os.O_WRONLY, 0600)
					if err != nil {
						return fmt.Errorf("failed to open known hosts file for writing: %w", err)
					}

					defer khHandle.Close()

					if err := knownhosts.WriteKnownHost(khHandle, hostname, remote, key); err != nil {
						return fmt.Errorf("failed to write to known hosts: %w", err)
					}
				}

				return nil
			},
		),
	}

	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("dial ssh: %w", err)
	}

	return &SSHClient{
		conn,
	}, nil
}

func NewSSHAgentClient(sshAuthSock string) (agent.ExtendedAgent, error) {
	sshAgent, err := net.Dial("unix", sshAuthSock)
	if err != nil {
		return nil, fmt.Errorf("dial agent auth sock: %w", err)
	}

	sshAgentClient := agent.NewClient(sshAgent)

	return sshAgentClient, nil
}
