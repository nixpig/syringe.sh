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

func GetPublicKey(path string) (gossh.PublicKey, error) {
	fc, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	publicKey, _, _, _, err := gossh.ParseAuthorizedKey(fc)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

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

func NewSignersFunc(publicKey gossh.PublicKey, agentSigners []gossh.Signer) func() ([]gossh.Signer, error) {
	return func() ([]gossh.Signer, error) {
		var signers []gossh.Signer

		for _, signer := range agentSigners {
			if string(publicKey.Marshal()) == string(signer.PublicKey().Marshal()) {
				signers = append(signers, signer)
			}
		}

		if len(signers) == 0 {
			return nil, errors.New("no valid signers in agent")
		}

		return signers, nil
	}
}
