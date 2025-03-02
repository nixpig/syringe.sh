package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	gossh "golang.org/x/crypto/ssh"
)

type PasswordReader func(int) ([]byte, error)

type Cryptor func(string) (string, error)

func GetPublicKey(path string) (gossh.PublicKey, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key from path (%s): %w", path, err)
	}

	publicKey, _, _, _, err := gossh.ParseAuthorizedKey(f)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return publicKey, nil
}

func GetPrivateKey(path string, r PasswordReader) (*rsa.PrivateKey, error) {
	var err error

	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key from path (%s): %w", path, err)
	}

	var key interface{}

	key, err = gossh.ParseRawPrivateKey(f)
	if err != nil {
		if _, ok := err.(*gossh.PassphraseMissingError); !ok {
			return nil, fmt.Errorf("parse private key without passphrase: %w", err)
		}

		_, err := os.Stdout.Write([]byte(fmt.Sprintf("Enter passphrase for %s: ", path)))
		if err != nil {
			return nil, fmt.Errorf("prompt for passphrase: %w", err)
		}

		passphrase, err := r(int(os.Stdin.Fd()))
		if err != nil {
			return nil, fmt.Errorf("read in passphrase: %w", err)
		}

		os.Stdout.Write([]byte("\n"))

		key, err = gossh.ParseRawPrivateKeyWithPassphrase(f, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("parse private key with passphrase: %w", err)
		}
	}

	rsaPrivateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast key (%s) to rsa private key", path)
	}

	return rsaPrivateKey, nil
}

func NewEncryptor(publicKey gossh.PublicKey) Cryptor {
	return func(s string) (string, error) {
		authorisedKey, _, _, _, err := gossh.ParseAuthorizedKey([]byte(
			gossh.MarshalAuthorizedKey(publicKey),
		))
		if err != nil {
			return "", fmt.Errorf("parse authorised key: %w", err)
		}

		cryptoKey, ok := authorisedKey.(gossh.CryptoPublicKey)
		if !ok {
			return "", fmt.Errorf("failed to cast authorised key to crypto key")
		}

		rsaPublicKey, ok := cryptoKey.CryptoPublicKey().(*rsa.PublicKey)
		if !ok {
			return "", fmt.Errorf("failed to cast crypto key to rsa public key")
		}

		encryptedValue, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			rsaPublicKey,
			[]byte(s),
			nil,
		)
		if err != nil {
			return "", fmt.Errorf("encrypt provided value: %w", err)
		}

		return base64.StdEncoding.EncodeToString(encryptedValue), nil
	}
}

func NewDecryptor(privateKey *rsa.PrivateKey) Cryptor {
	return func(s string) (string, error) {
		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", fmt.Errorf("decode cypher text: %w", err)
		}

		decryptedValue, err := rsa.DecryptOAEP(
			sha256.New(),
			rand.Reader,
			privateKey,
			data,
			nil,
		)
		if err != nil {
			return "", fmt.Errorf("decrypt cypher text: %w", err)
		}

		return string(decryptedValue), nil
	}
}
