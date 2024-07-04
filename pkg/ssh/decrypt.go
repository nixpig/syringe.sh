package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

func Decrypt(cypherText string, privateKey *rsa.PrivateKey) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cypherText)
	if err != nil {
		return "", err
	}

	decrypted, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		data,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
