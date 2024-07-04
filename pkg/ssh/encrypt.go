package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/charmbracelet/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func Encrypt(secret string, publicKey ssh.PublicKey) (string, error) {
	parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(gossh.MarshalAuthorizedKey(publicKey)))
	if err != nil {
		return "", err
	}

	parsedCryptoKey, ok := parsed.(gossh.CryptoPublicKey)
	if !ok {
		return "", errors.New("unable to parse parsed to ssh.CryptoPublicKey")
	}

	pubCrypto := parsedCryptoKey.CryptoPublicKey()

	pub, ok := pubCrypto.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("unable to parse pubCrypto to *rsa.PublicKey")
	}

	encryptedSecret, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		[]byte(secret),
		nil,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedSecret), nil
}
