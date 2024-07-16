package ssh_test

import (
	"os"
	"testing"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/stretchr/testify/require"
)

func TestSSHConfig(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test add identity to ssh config new host":                   testAddIdentityToSSHConfigNewHost,
		"test add identity to ssh config existing host":              testAddIdentityToSSHConfigExistingHost,
		"test add identity to ssh config existing identity":          testAddIdentityToSSHConfigExistingIdentity,
		"test add identity to ssh config don't match on home prefix": testAddIdentityToSSHConfigHomePrefix,
		"test add identity to ssh config match on home prefix":       testAddIdentityToSSHConfigHomePrefixMatch,
		"test add identity to invalid ssh config file error":         testAddIdentityToInvalidSSHConfigFileError,
		"test add empty identity to ssh config error":                testAddEmptyIdentityToSSHConfigError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, fn)
	}
}

func testAddIdentityToSSHConfigNewHost(t *testing.T) {
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	id := "../../test/crypt_test_rsa"
	hostname := "localhost"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.NoError(t, err)

	// read contents of file and check
	w, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	require.Equal(
		t,
		"Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ../../test/crypt_test_rsa\n",
		string(w),
	)
}

func testAddIdentityToSSHConfigExistingHost(t *testing.T) {
	hostname := "localhost"
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	f.WriteString("Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\n")

	id := "../../test/crypt_test_rsa"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.NoError(t, err)

	w, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	require.Equal(
		t,
		"Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ../../test/crypt_test_rsa\n",
		string(w),
	)
}

func testAddIdentityToSSHConfigExistingIdentity(t *testing.T) {
	hostname := "localhost"
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	f.WriteString("Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ../../test/crypt_test_rsa\n")

	id := "../../test/crypt_test_rsa"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.NoError(t, err)

	w, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	require.Equal(
		t,
		"Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ../../test/crypt_test_rsa\n",
		string(w),
	)
}

func testAddIdentityToSSHConfigHomePrefix(t *testing.T) {
	hostname := "localhost"
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	f.WriteString("Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ~/test/crypt_test_rsa\n")

	id := "../../test/crypt_test_rsa"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.NoError(t, err)

	w, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	require.Equal(
		t,
		"Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ~/test/crypt_test_rsa\nIdentityFile ../../test/crypt_test_rsa\n",
		string(w),
	)
}

func testAddIdentityToSSHConfigHomePrefixMatch(t *testing.T) {
	hostname := "localhost"
	os.Setenv("HOME", "/home/test")
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	f.WriteString("Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ~/crypt_test_rsa\n")

	id := "/home/test/crypt_test_rsa"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.NoError(t, err)

	w, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	require.Equal(
		t,
		"Host localhost\nAddKeysToAgent yes\nIgnoreUnknown UseKeychain\nUseKeychain yes\nIdentityFile ~/crypt_test_rsa\n",
		string(w),
	)
}

func testAddIdentityToInvalidSSHConfigFileError(t *testing.T) {
	hostname := "localhost"
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	// remove so read error
	f.Close()
	os.Remove(f.Name())

	id := "../../test/crypt_test_rsa"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.Error(t, err)
	require.Contains(t, err.Error(), "file already closed")
}

func testAddEmptyIdentityToSSHConfigError(t *testing.T) {
	f, err := os.CreateTemp("", "tmp_ssh_config")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	hostname := "" // empty
	id := "../../test/crypt_test_rsa"

	err = ssh.AddIdentityToSSHConfig(id, hostname, f)
	require.EqualError(t, err, "ssh_config: empty pattern")
}
