package ssh

import (
	"fmt"
	"io"
	"os"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func GetSigner(path string, out io.Writer) (gossh.Signer, error) {
	var err error

	fc, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var signer gossh.Signer

	signer, err = gossh.ParsePrivateKey(fc)
	if err != nil {
		if _, ok := err.(*gossh.PassphraseMissingError); !ok {
			return nil, err
		}

		out.Write([]byte(fmt.Sprintf("Enter passphrase for %s: ", path)))

		passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}

		signer, err = gossh.ParsePrivateKeyWithPassphrase(fc, passphrase)
		if err != nil {
			return nil, err
		}
	}

	return signer, nil
}
