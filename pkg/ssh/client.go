package client

import (
	"fmt"
	"io"
	"net"
	"os"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func SSHClient(
	host string,
	port int,
	identity string,
	command string,
) ([]byte, error) {
	keyFile, err := os.Open(identity)
	if err != nil {
		return nil, err
	}

	keyContents, err := io.ReadAll(keyFile)
	if err != nil {
		return nil, err
	}

	var signer gossh.Signer

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

	config := &gossh.ClientConfig{
		User: "nixpig",
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error {
			return nil
		}),
	}

	conn, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
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
