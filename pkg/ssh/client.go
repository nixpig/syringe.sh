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

var keyError *knownhosts.KeyError
var revokedError *knownhosts.RevokedError

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

	sshConfig.HostKeyCallback = gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		khf := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")

		kh, err := knownhosts.New(khf)
		if err != nil {
			return fmt.Errorf("unable to open knownhosts file: %w", err)
		}

		kerr := kh(fmt.Sprintf("%s:%d", host, port), remote, key)

		if errors.As(kerr, &revokedError) {
			return fmt.Errorf("key revoked: %w", revokedError)
		}

		if errors.As(kerr, &keyError) && len(keyError.Want) > 0 {
			return fmt.Errorf("host %s is not a key of %s: %w", key, host, keyError)
		}

		if errors.As(err, &keyError) && len(keyError.Want) == 0 {
			khfh, err := os.OpenFile(khf, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("unable to open known hosts file for writing: %w", err)
			}

			defer khfh.Close()

			knownhost := knownhosts.Normalize(remote.String())

			if _, err := khfh.WriteString(knownhosts.Line([]string{knownhost}, key)); err != nil {
				return fmt.Errorf("failed to write to knownhosts: %w", err)
			}

			return nil
		}

		fmt.Println("kerr: ", kerr)

		fmt.Println("pub key already exists!!")
		return nil
	})

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

	} else {
		sshAgent, err := net.Dial("unix", sshAuthSock)
		if err != nil {
			return nil, err
		}

		sshConfig.Auth = []gossh.AuthMethod{
			gossh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
		}

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
