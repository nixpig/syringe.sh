package ssh_test

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"os"
	"strings"
	"testing"

	tt "github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func TestSSHUtils(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test get private key no password happy path":               testGetPrivateKeyNoPasswordHappyPath,
		"test get private key with password happy path":             testGetPrivateKeyWithPasswordHappyPath,
		"test get private key with empty passphrase error":          testGetPrivateKeyEmptyPassphraseError,
		"test get private key with incorrect passphrase error":      testGetPrivateKeyIncorrectPassphraseError,
		"test get private key with password read error":             testGetPrivateKeyPasswordReadError,
		"test get private key with invalid filepath error":          testGetPrivateKeyWithInvalidFilepathError,
		"test get private key with invalid contents error":          testGetPrivateKeyWithInvalidContentsError,
		"test get private key with password invalid contents error": testGetPrivateKeyWithPasswordInvalidContentsError,

		"test get signer happy path":                           testGetSignerHappyPath,
		"test get signer no password happy path":               testGetSignerNoPasswordHappyPath,
		"test get signer with password happy path":             testGetSignerWithPasswordHappyPath,
		"test get signer with empty passphrase error":          testGetSignerEmptyPassphraseError,
		"test get signer with incorrect passphrase error":      testGetSignerIncorrectPassphraseError,
		"test get signer with password read error":             testGetSignerPasswordReadError,
		"test get signer with invalid filepath error":          testGetSignerWithInvalidFilepathError,
		"test get signer with invalid contents error":          testGetSignerWithInvalidContentsError,
		"test get signer with password invalid contents error": testGetSignerWithPasswordInvalidContentsError,

		"test get public key happy path":                  testGetPublicKeyHappyPath,
		"test get public key with invalid filepath error": testGetPublicKeyWithInvalidFilepathError,
		"test get public key with invalid contents error": testGetPublicKeyWithInvalidContentsError,

		"test new signers func happy path":       testNewSignersFuncHappyPath,
		"test new signers func no signers error": testNewSignersFuncNoSignersError,

		"test auth method from identity happy path": testAuthMethodFromIdentityHappyPath,
		"test auth method get public key error":     testAuthMethodGetPublicKeyError,
		"test auth method get signer error":         testAuthMethodGetSignerError,

		"test auth method from agent happy path":            testAuthMethodFromAgentHappyPath,
		"test auth method from agent get private key error": testAuthMethodFromAgentGetPrivateKeyError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, fn)
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
	require.Empty(t, w.String())
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
	require.Empty(t, w.String())
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

func testGetPublicKeyWithInvalidContentsError(t *testing.T) {
	key, err := ssh.GetPublicKey("../../test/crypt_test_invalid.pub")

	require.Empty(t, key)
	require.EqualError(t, err, "ssh: no key found")
}

func testGetPublicKeyHappyPath(t *testing.T) {
	key, err := ssh.GetPublicKey("../../test/crypt_test_rsa.pub")

	require.NoError(t, err)
	require.Implements(t, (*gossh.PublicKey)(nil), key)
}

func testGetSignerHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	signer, err := ssh.GetSigner("../../test/crypt_test_rsa", w, mockTerm.ReadPassword)
	require.NoError(t, err)
	require.Empty(t, w.String())

	require.Implements(t, (*gossh.Signer)(nil), signer)
}

func testNewSignersFuncHappyPath(t *testing.T) {
	publicKey, privateKey1, err := test.GenerateKeyPair()
	require.NoError(t, err)
	signer1, err := gossh.NewSignerFromKey(privateKey1)
	require.NoError(t, err)

	_, privateKey2, err := test.GenerateKeyPair()
	require.NoError(t, err)
	signer2, err := gossh.NewSignerFromKey(privateKey2)
	require.NoError(t, err)

	_, privateKey3, err := test.GenerateKeyPair()
	require.NoError(t, err)
	signer3, err := gossh.NewSignerFromKey(privateKey3)
	require.NoError(t, err)

	signers := []gossh.Signer{
		signer1,
		signer2,
		signer3,
	}

	signersFunc := ssh.NewSignersFunc(publicKey, signers)

	validSigners, err := signersFunc()

	require.NoError(t, err)

	require.Len(t, validSigners, 1)
	require.Equal(t, signer1, validSigners[0])
}

func testNewSignersFuncNoSignersError(t *testing.T) {
	_, privateKey1, err := test.GenerateKeyPair()
	require.NoError(t, err)
	signer1, err := gossh.NewSignerFromKey(privateKey1)
	require.NoError(t, err)

	publicKey, privateKey2, err := test.GenerateKeyPair()
	require.NoError(t, err)
	_, err = gossh.NewSignerFromKey(privateKey2)
	require.NoError(t, err)

	_, privateKey3, err := test.GenerateKeyPair()
	require.NoError(t, err)
	signer3, err := gossh.NewSignerFromKey(privateKey3)
	require.NoError(t, err)

	signers := []gossh.Signer{
		signer1,
		signer3,
	}

	signersFunc := ssh.NewSignersFunc(publicKey, signers)

	validSigners, err := signersFunc()

	require.EqualError(t, err, "no valid signers in agent")
	require.Empty(t, validSigners)
}

// -----------------------------------

func testGetSignerEmptyPassphraseError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte(""), nil)

	signer, err := ssh.GetSigner("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "bcrypt_pbkdf: empty password")
	require.Empty(t, signer)

	mockTermReadPassword.Unset()
}

func testGetSignerIncorrectPassphraseError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("incorrect_passphrase"), nil)

	signer, err := ssh.GetSigner("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "x509: decryption password incorrect")
	require.Empty(t, signer)

	mockTermReadPassword.Unset()
}

func testGetSignerWithInvalidFilepathError(t *testing.T) {
	w := bytes.NewBufferString("")
	signer, err := ssh.GetSigner("some/invalid/filepath", w, mockTerm.ReadPassword)

	require.Empty(t, signer)
	require.Empty(t, w.String())
	require.EqualError(t, err, "open some/invalid/filepath: no such file or directory")
}

func testGetSignerWithInvalidContentsError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	signer, err := ssh.GetSigner("../../test/crypt_test_invalid", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "ssh: no key found")
	require.Empty(t, signer)

	mockTermReadPassword.Unset()
}

func testGetSignerWithPasswordInvalidContentsError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	signer, err := ssh.GetSigner("../../test/crypt_test_pass_invalid", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "ssh: no key found")
	require.Empty(t, signer)

	mockTermReadPassword.Unset()
}

func testGetSignerPasswordReadError(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte{}, errors.New("failed to read password"))

	signer, err := ssh.GetSigner("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.EqualError(t, err, "failed to read password: failed to read password")
	require.Empty(t, signer)

	mockTermReadPassword.Unset()
}

func testGetSignerNoPasswordHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")
	signer, err := ssh.GetSigner("../../test/crypt_test_rsa", w, mockTerm.ReadPassword)

	require.NoError(t, err)
	require.Empty(t, w.String())
	require.Implements(t, (*gossh.Signer)(nil), signer)
}

func testGetSignerWithPasswordHappyPath(t *testing.T) {
	w := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.On("ReadPassword", mock.Anything).Return([]byte("test"), nil)

	signer, err := ssh.GetSigner("../../test/crypt_test_pass_rsa", w, mockTerm.ReadPassword)

	require.NoError(t, err)
	require.Implements(t, (*gossh.Signer)(nil), signer)

	mockTermReadPassword.Unset()
}

func testAuthMethodFromIdentityHappyPath(t *testing.T) {
	identity := "../../test/crypt_test_rsa"
	os.Setenv("SSH_AUTH_SOCK", "") // empty to fallback to identity

	out := bytes.NewBufferString("")

	authMethod, err := ssh.AuthMethod(identity, out)

	require.NoError(t, err)
	require.Implements(t, (*gossh.AuthMethod)(nil), authMethod)
}

func testAuthMethodGetPublicKeyError(t *testing.T) {
	// force error  from GetPublicKey with invalid file path
	identity := "not_found_file_path"

	out := bytes.NewBufferString("")

	authMethod, err := ssh.AuthMethod(identity, out)

	require.Error(t, err)
	require.Nil(t, authMethod)
}

func testAuthMethodGetSignerError(t *testing.T) {
	// force error from GetSigner with incorrect password
	identity := "../../test/crypt_test_pass_rsa"

	out := bytes.NewBufferString("")

	mockTermReadPassword := mockTerm.
		On("ReadPassword", mock.Anything).
		Return([]byte("something wrong in here"), nil)

	authMethod, err := ssh.AuthMethod(identity, out)

	require.Error(t, err)
	require.Nil(t, authMethod)

	mockTermReadPassword.Unset()
}

func testAuthMethodFromAgentHappyPath(t *testing.T) {
	keyPair := tt.GenerateRSAKeyPair(t, 4096)

	sshAgent := tt.SshAgentWithKeyPair(t, keyPair)

	defer sshAgent.Stop()

	os.Setenv("SSH_AUTH_SOCK", sshAgent.SocketFile())

	// different key than is already in the agent, so it goes through the 'add key to agent' branch
	authMethod, err := ssh.AuthMethod("../../test/crypt_test_rsa", bytes.NewBufferString(""))
	require.NoError(t, err)

	require.Implements(t, (*gossh.AuthMethod)(nil), authMethod)

	os.Setenv("SSH_AUTH_SOCK", "")
}

func testAuthMethodFromAgentGetPrivateKeyError(t *testing.T) {
	keyPair := tt.GenerateRSAKeyPair(t, 4096)

	sshAgent := tt.SshAgentWithKeyPair(t, keyPair)

	defer sshAgent.Stop()

	os.Setenv("SSH_AUTH_SOCK", sshAgent.SocketFile())

	// use incorrect passphrase to force error in 'get private key' flow
	mockTermReadPassword := mockTerm.
		On("ReadPassword", mock.Anything).
		Return([]byte("incorrect passphrase in here"), nil)

	// different key than is already in the agent, so it goes through the 'add key to agent' branch
	authMethod, err := ssh.AuthMethod("../../test/crypt_test_pass_rsa", bytes.NewBufferString(""))
	require.True(t, strings.HasPrefix(err.Error(), "failed to read private key"))
	require.Empty(t, authMethod)

	os.Setenv("SSH_AUTH_SOCK", "")
	mockTermReadPassword.Unset()
}
