package ssh_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/test"
	"github.com/stretchr/testify/require"
)

func TestCrypt(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test encrypt/decrypt happy path": testEncryptDecryptHappyPath,
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
