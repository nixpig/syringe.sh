package ssh_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// secret_value
// eW1Ohpp+p9rNFZiRu03hE+CpEaNOEVnXWsLUHtb0s8SUsvIyUnQt531XGvK67jwU2SVJIt4sSZLTWnp8iLOePTTzQLVF9AcSWXPTpQSWq7XgS9B91GIK+JxPaLMQ4PAE4w8J9F5K/xuLIO763EzUcGfFBEqJlHxz+h9tRJIxpoNuxV9vjf7s+dNnFSfIKOeFA47e+sG1FhNZhWPJk5Xqgmxx/RjwqW+RcLIciUokG5mm9yROpe8I5JPJB4DZsFFZPnPaKmdUQqUnfSfQRHYwOezVzU3I45KGvYuTjO9Atnk6OjjWHj4x6jlA3yvrU1RKc5ofv+YZPPjZa+2TqkmEjJLa1Dto5xjrmwl/vti1j3S8Hp8B6hnniJemzzcSuhkU7vTUivxPjM+mC6IN93sDzn5y8eUO4Gpuz/gkv4O1FJu0ZpDd5S+KMhQPWku7sLv+cv3a2PWAKY+zhMJq9l0MXzLu0W8Aj7OQpboMs+SC3AuP6PWf5RxcSd1vHD+PtYla

func TestSSH(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test encrypt/decrypt happy path":                      testEncryptDecryptHappyPath,
		"test encrypt/decrypt with passphrase happy path":      testEncryptDecryptWithPassphraseHappyPath,
		"test get private key with empty passphrase error":     testGetPrivateKeyEmptyPassphraseError,
		"test get private key with incorrect passphrase error": testGetPrivateKeyIncorrectPassphraseError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}

}

type MockTerm struct {
	mock.Mock
}

func (mt *MockTerm) ReadPassword(fd int) ([]byte, error) {
	args := mt.Called(fd)

	return args.Get(0).([]byte), args.Error(1)
}

var mockTerm = new(MockTerm)

func testEncryptDecryptHappyPath(t *testing.T) {
	publicKey, err := ssh.GetPublicKey("../../test/crypt_test_rsa.pub")
	require.NoError(t, err)

	encryptedSecret, err := ssh.Encrypt(
		"secret_value",
		publicKey,
	)

	require.NoError(t, err)

	w := bytes.NewBufferString("")

	privateKey, err := ssh.GetPrivateKey("../../test/crypt_test_rsa", w, mockTerm.ReadPassword)
	require.NoError(t, err)

	decryptedSecret, err := ssh.Decrypt(
		encryptedSecret,
		privateKey,
	)

	b, err := io.ReadAll(w)
	require.NoError(t, err)
	require.Empty(t, b)

	require.Equal(
		t,
		"secret_value",
		decryptedSecret,
	)
}

func testEncryptDecryptWithPassphraseHappyPath(t *testing.T) {
	publicKey, err := ssh.GetPublicKey("../../test/crypt_test_pass_rsa.pub")
	require.NoError(t, err)

	encryptedSecret, err := ssh.Encrypt(
		"secret_value",
		publicKey,
	)

	require.NoError(t, err)

	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	privateKey, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)
	require.NoError(t, err)

	decryptedSecret, err := ssh.Decrypt(
		encryptedSecret,
		privateKey,
	)

	_, err = io.ReadAll(w)
	require.NoError(t, err)

	require.Equal(
		t,
		"secret_value",
		decryptedSecret,
	)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyEmptyPassphraseError(t *testing.T) {
	publicKey, err := ssh.GetPublicKey("../../test/crypt_test_pass_rsa.pub")
	require.NoError(t, err)

	_, err = ssh.Encrypt(
		"secret_value",
		publicKey,
	)

	require.NoError(t, err)

	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte(""), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "bcrypt_pbkdf: empty password")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyIncorrectPassphraseError(t *testing.T) {
	publicKey, err := ssh.GetPublicKey("../../test/crypt_test_pass_rsa.pub")
	require.NoError(t, err)

	_, err = ssh.Encrypt(
		"secret_value",
		publicKey,
	)

	require.NoError(t, err)

	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("incorrect_passphrase"), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "x509: decryption password incorrect")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}
