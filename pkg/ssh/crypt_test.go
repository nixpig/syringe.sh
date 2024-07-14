package ssh_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/test"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func TestCrypt(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test encrypt/decrypt happy path":        testEncryptDecryptHappyPath,
		"test decrypt invalid cypher text error": testDecryptInvalidCypherTextError,
		"test encrypt invalid public key error":  testEncryptInvalidPublicKeyError,
		"test decrypt invalid private key error": testDecryptInvalidPrivateKeyError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}

}

func testEncryptDecryptHappyPath(t *testing.T) {
	publicKey, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	encryptedSecret, err := ssh.Encrypt(
		"secret_value",
		publicKey,
	)
	require.NoError(t, err)

	w := bytes.NewBufferString("")

	decryptedSecret, err := ssh.Decrypt(
		encryptedSecret,
		privateKey,
	)
	require.NoError(t, err)

	b, err := io.ReadAll(w)
	require.NoError(t, err)
	require.Empty(t, b)

	require.Equal(
		t,
		"secret_value",
		decryptedSecret,
	)
}

func testDecryptInvalidCypherTextError(t *testing.T) {
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	decryptedSecret, err := ssh.Decrypt(
		"invalid base64 cypher text",
		privateKey,
	)
	require.Error(t, err)
	require.Empty(t, decryptedSecret)
	require.Contains(t, err.Error(), "failed to decode cypher text: illegal base64 data")
}

func testEncryptInvalidPublicKeyError(t *testing.T) {
	// create an 'empty' public key
	publicKey := emptyPublicKey{}

	encryptedSecret, err := ssh.Encrypt(
		"secret_value",
		publicKey,
	)
	require.EqualError(t, err, "ssh: no key found")
	require.Empty(t, encryptedSecret)
}

func testDecryptInvalidPrivateKeyError(t *testing.T) {
	// generate non-matching public and private keys
	publicKey, _, err := test.GenerateKeyPair()
	require.NoError(t, err)
	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	encryptedSecret, err := ssh.Encrypt(
		"secret_value",
		publicKey,
	)
	require.NoError(t, err)

	decryptedSecret, err := ssh.Decrypt(
		encryptedSecret,
		privateKey,
	)
	require.Error(t, err)
	require.Empty(t, decryptedSecret)
	require.Contains(t, err.Error(), "failed to decrypt cypher text")
}

type emptyPublicKey struct{}

func (k emptyPublicKey) Type() string                                   { return "" }
func (k emptyPublicKey) Marshal() []byte                                { return []byte{} }
func (k emptyPublicKey) Verify(data []byte, sig *gossh.Signature) error { return nil }
