package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

func SSHClient(
	host string,
	port int,
	identity string,
	command string,
) ([]byte, error) {
	currentUser, err := user.Current()
	if err != nil || currentUser.Username == "" {
		return nil, err
	}

	sshConfig := &gossh.ClientConfig{
		User: currentUser.Username,
	}

	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")

	// if sshAuthSock == "" {
	if identity == "" {
		var signer gossh.Signer

		keyFile, err := os.Open(identity)
		if err != nil {
			return nil, err
		}

		keyContents, err := io.ReadAll(keyFile)
		if err != nil {
			return nil, err
		}

		signer, err = gossh.ParsePrivateKey(keyContents)
		if err != nil {
			_, ok := err.(*gossh.PassphraseMissingError)
			if !ok {
				return nil, err
			}

			fmt.Printf("Enter passphrase for key '%s': ", identity)
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

		sshConfig.Auth = []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		}

		sshConfig.HostKeyCallback = gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			fmt.Println("calling back...")
			khf := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")

			fmt.Println("creating knownhosts for: ", khf)
			kh, err := knownhosts.New(khf)
			if err != nil {
				return errors.New("couldn't open known_hosts file")
			}

			var kerr *knownhosts.KeyError

			fmt.Println("checking against knownhosts")
			fmt.Println("host: ", host)
			fmt.Println("remote: ", remote.String())
			fmt.Println("key: ", key)
			if err := kh(fmt.Sprintf("%s:%d", host, port), remote, key); err != nil {
				fmt.Println("checking if matches want")
				if errors.As(err, &kerr) && len(kerr.Want) > 0 {
					return fmt.Errorf("host %s is not a key of %s: %w", key, host, err)
				}

				fmt.Println("checking if present")
				if errors.As(err, &kerr) && len(kerr.Want) == 0 {
					fmt.Println("host key not present; adding to known hosts")

					khfh, err := os.OpenFile(khf, os.O_APPEND|os.O_WRONLY, 0600)
					if err != nil {
						return errors.New("unable to open known hosts file for writing")
					}

					defer khfh.Close()

					knownhost := knownhosts.Normalize(remote.String())

					fmt.Println("writing new known hosts")
					if _, err := khfh.WriteString(knownhosts.Line([]string{knownhost}, key)); err != nil {
						return errors.New("failed to write to knownhosts")
					}
				}
			}

			return nil
		})

	} else {
		sshAgent, err := net.Dial("unix", sshAuthSock)
		if err != nil {
			return nil, err
		}

		sshConfig.Auth = []gossh.AuthMethod{
			gossh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
		}

		sshConfig.HostKeyCallback = gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			// TODO: verify keys when used from agent
			return nil
		})
	}

	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return nil, err
	}

	return output, nil
}
