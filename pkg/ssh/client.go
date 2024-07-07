package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/skeema/knownhosts"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

type SSHClient struct {
	client *gossh.Client
}

func (s *SSHClient) Close() error {
	if err := s.client.Close(); err != nil {
		return err
	}

	return nil
}

func (s *SSHClient) Run(cmd string, w io.Writer) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		if _, err := w.Write(output); err != nil {
			return err
		}
		return err
	}

	// var privateKey *rsa.PrivateKey
	//
	// privateKey, err = GetPrivateKey("/home/nixpig/.ssh/id_rsa_test2")
	// if err != nil {
	// 	_, ok := err.(*gossh.PassphraseMissingError)
	// 	if !ok {
	// 		return err
	// 	}
	//
	// 	fmt.Printf("Enter passphrase for %s: ", "/home/nixpig/.ssh/id_rsa_test2")
	// 	passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
	// 	fmt.Print("\n")
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	privateKey, err = GetPrivateKeyWithPassphrase("/home/nixpig/.ssh/id_rsa_test2", string(passphrase))
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	// decrypted, err := Decrypt(string(output), privateKey)
	// if err != nil {
	// 	return err
	// }
	//
	// output = []byte(decrypted)

	if _, err := w.Write(output); err != nil {
		return err
	}

	return nil
}

func NewSSHClient(
	host string,
	port int,
	username string,
	authMethod gossh.AuthMethod,
) (*SSHClient, error) {
	sshConfig := &gossh.ClientConfig{
		User: username,
		Auth: []gossh.AuthMethod{authMethod},

		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			khPath := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")

			kh, err := knownhosts.New(khPath)
			if err != nil {
				return fmt.Errorf("failed to open knownhosts file: %w", err)
			}

			err = kh(fmt.Sprintf("%s:%d", host, port), remote, key)

			if knownhosts.IsHostKeyChanged(err) {
				return fmt.Errorf("remote host identification has changed which may indicate a MITM attack: %w", err)
			}

			if knownhosts.IsHostUnknown(err) {
				khHandle, err := os.OpenFile(khPath, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					return fmt.Errorf("failed to open known hosts file for writing: %w", err)
				}

				defer khHandle.Close()

				if err := knownhosts.WriteKnownHost(khHandle, hostname, remote, key); err != nil {
					return fmt.Errorf("failed to write to known hosts: %w", err)
				}

				fmt.Printf("added host %s to known hosts\n", hostname)
			}

			return nil
		}),
	}

	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, err
	}

	return &SSHClient{
		conn,
	}, nil
}

func NewSSHAgentClient(sshAuthSock string) (agent.ExtendedAgent, error) {
	sshAgent, err := net.Dial("unix", sshAuthSock)
	if err != nil {
		return nil, err
	}

	sshAgentClient := agent.NewClient(sshAgent)

	return sshAgentClient, nil
}

func AgentAuthMethod(signersFunc func() ([]gossh.Signer, error)) (gossh.AuthMethod, error) {

	// TODO: find the agent key which matches the provided identity
	// TODO: create a signer from the key and use that for the publickeyscallback
	// DO NOT JUST PASS IN EVERYTHING IN THE AGENT!!!

	authMethod := gossh.PublicKeysCallback(signersFunc)

	return authMethod, nil
}

func IdentityAuthMethod(identity string) (gossh.AuthMethod, error) {
	var signer gossh.Signer

	keyContents, err := os.ReadFile(identity)
	if err != nil {
		return nil, err
	}

	signer, err = gossh.ParsePrivateKey(keyContents)
	if err != nil {
		_, ok := err.(*gossh.PassphraseMissingError)
		if !ok {
			return nil, err
		}

		fmt.Printf("Enter passphrase for %s: ", identity)
		passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Print("\n")
		if err != nil {
			return nil, err
		}

		signer, err = gossh.ParsePrivateKeyWithPassphrase(keyContents, passphrase)
		if err != nil {
			return nil, err
		}
	}

	authMethod := gossh.PublicKeys(signer)

	return authMethod, nil
}
