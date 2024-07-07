package ssh

import (
	"os"

	gossh "golang.org/x/crypto/ssh"
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
