package ssh

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"os"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func GetPrivateKey(path string, out io.Writer) (*rsa.PrivateKey, error) {
	var err error

	fc, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var key interface{}

	key, err = gossh.ParseRawPrivateKey(fc)
	if err != nil {
		if _, ok := err.(*gossh.PassphraseMissingError); !ok {
			return nil, err
		}

		out.Write([]byte(fmt.Sprintf("Enter passphrase for %s: ", path)))

		passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}

		key, err = gossh.ParseRawPrivateKeyWithPassphrase(fc, []byte(passphrase))
		if err != nil {
			return nil, err
		}
	}

	rsaPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed to cast to rsa private key")
	}

	return rsaPrivateKey, nil
}
