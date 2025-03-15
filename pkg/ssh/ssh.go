package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

type PasswordReader func(int) ([]byte, error)

type Cryptor func(string) (string, error)

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

func GetPrivateKey(path string, out io.Writer, pr PasswordReader) (*rsa.PrivateKey, error) {
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

		passphrase, err := pr(int(os.Stdin.Fd()))
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}

		out.Write([]byte("\n"))

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

func GetSigner(path string, out io.Writer, pr PasswordReader) (gossh.Signer, error) {
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

		passphrase, err := pr(int(os.Stdin.Fd()))
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

func AuthMethod(identity string, out io.Writer) (gossh.AuthMethod, error) {
	var authMethod gossh.AuthMethod

	publicKey, err := GetPublicKey(fmt.Sprintf("%s.pub", identity))
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	sshAgentClient, err := NewSSHAgentClient(sshAuthSock)
	if err != nil {
		out.Write([]byte("unable to connect to agent, falling back to identity"))

		signer, err := GetSigner(identity, out, term.ReadPassword)
		if err != nil {
			return nil, err
		}

		authMethod = gossh.PublicKeys(signer)

	} else {
		agentKeys, err := sshAgentClient.List()
		if err != nil {
			return nil, fmt.Errorf("failed to get identities from ssh agent: %w", err)
		}

		// if the agent doesn't already contain the identity, then add it
		if i := slices.IndexFunc(agentKeys, func(agentKey *agent.Key) bool {
			return string(agentKey.Marshal()) == string(publicKey.Marshal())
		}); i == -1 {
			privateKey, err := GetPrivateKey(identity, out, term.ReadPassword)
			if err != nil {
				return nil, fmt.Errorf("failed to read private key: %w", err)
			}

			if err := sshAgentClient.Add(agent.AddedKey{PrivateKey: privateKey}); err != nil {
				return nil, fmt.Errorf("failed to add key to agent: %w", err)
			}
		}

		sshAgentClientSigners, err := sshAgentClient.Signers()
		if err != nil {
			return nil, fmt.Errorf("failed to get signers from ssh client: %w", err)
		}

		authMethod = gossh.PublicKeysCallback(
			// use only signer for the specified identity key
			NewSignersFunc(publicKey, sshAgentClientSigners),
		)
	}

	return authMethod, nil
}
