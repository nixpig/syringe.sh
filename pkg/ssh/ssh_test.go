package ssh_test

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"io"
	"testing"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// secret_value
// eW1Ohpp+p9rNFZiRu03hE+CpEaNOEVnXWsLUHtb0s8SUsvIyUnQt531XGvK67jwU2SVJIt4sSZLTWnp8iLOePTTzQLVF9AcSWXPTpQSWq7XgS9B91GIK+JxPaLMQ4PAE4w8J9F5K/xuLIO763EzUcGfFBEqJlHxz+h9tRJIxpoNuxV9vjf7s+dNnFSfIKOeFA47e+sG1FhNZhWPJk5Xqgmxx/RjwqW+RcLIciUokG5mm9yROpe8I5JPJB4DZsFFZPnPaKmdUQqUnfSfQRHYwOezVzU3I45KGvYuTjO9Atnk6OjjWHj4x6jlA3yvrU1RKc5ofv+YZPPjZa+2TqkmEjJLa1Dto5xjrmwl/vti1j3S8Hp8B6hnniJemzzcSuhkU7vTUivxPjM+mC6IN93sDzn5y8eUO4Gpuz/gkv4O1FJu0ZpDd5S+KMhQPWku7sLv+cv3a2PWAKY+zhMJq9l0MXzLu0W8Aj7OQpboMs+SC3AuP6PWf5RxcSd1vHD+PtYla

func TestSSH(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test encrypt/decrypt happy path": testEncryptDecryptHappyPath,

		// TODO: test with correct password and valid contents!!!

		"test get private key no password happy path":               testGetPrivateKeyNoPasswordHappyPath,
		"test get private key with password happy path":             testGetPrivateKeyWithPasswordHappyPath,
		"test get private key with empty passphrase error":          testGetPrivateKeyEmptyPassphraseError,
		"test get private key with incorrect passphrase error":      testGetPrivateKeyIncorrectPassphraseError,
		"test get private key with password read error":             testGetPrivateKeyPasswordReadError,
		"test get private key with invalid filepath error":          testGetPrivateKeyWithInvalidFilepathError,
		"test get private key with invalid contents error":          testGetPrivateKeyWithInvalidContentsError,
		"test get private key with password invalid contents error": testGetPrivateKeyWithPasswordInvalidContentsError,

		"test get public key with invalid filepath error": testGetPublicKeyWithInvalidFilepathError,
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

func testGetPrivateKeyEmptyPassphraseError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte(""), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "bcrypt_pbkdf: empty password")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyIncorrectPassphraseError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("incorrect_passphrase"), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "x509: decryption password incorrect")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyWithInvalidFilepathError(t *testing.T) {
	w := bytes.NewBufferString("")
	key, err := ssh.GetPrivateKey("some/invalid/filepath", w, mockTerm.ReadPassword)

	require.Empty(t, key)
	require.EqualError(t, err, "open some/invalid/filepath: no such file or directory")
}

func testGetPublicKeyWithInvalidFilepathError(t *testing.T) {
	key, err := ssh.GetPublicKey("some/invalid/filepath")

	require.Empty(t, key)
	require.EqualError(t, err, "open some/invalid/filepath: no such file or directory")
}

func testGetPrivateKeyWithInvalidContentsError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_invalid", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "ssh: no key found")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyWithPasswordInvalidContentsError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_invalid", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "ssh: no key found")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyPasswordReadError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte{}, errors.New("failed to read password"))

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "failed to read password: failed to read password")
	require.Empty(t, key)

	mockTermReadPassword.Unset()
}

func testGetPrivateKeyNoPasswordHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	key, err := ssh.GetPrivateKey("../../test/crypt_test_rsa", w, mockTerm.ReadPassword)

	require.NoError(t, err)
	require.IsType(t, &rsa.PrivateKey{}, key)
}

func testGetPrivateKeyWithPasswordHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	key, err := ssh.GetPrivateKey("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.NoError(t, err)
	require.IsType(t, &rsa.PrivateKey{}, key)

	mockTermReadPassword.Unset()
}
