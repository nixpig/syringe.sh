package ssh

import (
	"crypto/rsa"
	"errors"
	"os"

	gossh "golang.org/x/crypto/ssh"
)

func GetPrivateKey(path string) (*rsa.PrivateKey, error) {
	fc, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	key, err := gossh.ParseRawPrivateKey(fc)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("unable to cast to rsa private key")
	}

	return rsaPrivateKey, nil
}
