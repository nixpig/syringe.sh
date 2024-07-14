package ssh_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	glssh "github.com/gliderlabs/ssh"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/test"
	"github.com/skeema/knownhosts"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func TestSSHClient(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test ssh client run command existing known host":      testSSHClientRunCommandExistingKnownHost,
		"test ssh client run command new known host":           testSSHClientRunCommandNewKnownHost,
		"test ssh client run command invalid known hosts file": testSSHClientRunCommandInvalidKnownHostsFile,
		"test ssh client run command known hosts key mismatch": testSSHClientRunCommandKnownHostsKeyMismatch,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, fn)
	}
}

func testSSHClientRunCommandExistingKnownHost(t *testing.T) {
	sshServer := &glssh.Server{
		Addr: "localhost:23234",
		Handler: func(s glssh.Session) {
			s.Write([]byte(fmt.Sprintf("handled %s", s.RawCommand())))
		},
		PublicKeyHandler: func(ctx glssh.Context, publicKey glssh.PublicKey) bool {
			return true
		},
	}

	go func() {
		sshServer.ListenAndServe()
	}()

	t.Cleanup(func() {
		sshServer.Close()
	})

	publicKey, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	sshServer.AddHostKey(signer)

	kh, err := os.CreateTemp("", "known_hosts")
	require.NoError(t, err)

	knownHost := knownhosts.Line([]string{"localhost:23234", "127.0.0.1:23234"}, publicKey)

	fmt.Println(knownHost)

	kh.WriteString(knownHost)

	client, err := ssh.NewSSHClient(
		"localhost",
		23234,
		"nixpig",
		gossh.PublicKeys(signer),
		kh.Name(),
	)
	require.NoError(t, err)

	out := bytes.NewBufferString("")

	err = client.Run("cmd --option val subcmd", out)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)

	require.Equal(t, "handled cmd --option val subcmd", out.String())

}

func testSSHClientRunCommandNewKnownHost(t *testing.T) {
	sshServer := &glssh.Server{
		Addr: "localhost:23234",
		Handler: func(s glssh.Session) {
			s.Write([]byte(fmt.Sprintf("handled %s", s.RawCommand())))
		},
		PublicKeyHandler: func(ctx glssh.Context, publicKey glssh.PublicKey) bool {
			return true
		},
	}

	go func() {
		sshServer.ListenAndServe()
	}()

	t.Cleanup(func() {
		sshServer.Close()
	})

	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	kh, err := os.CreateTemp("", "known_hosts")
	require.NoError(t, err)

	client, err := ssh.NewSSHClient(
		"localhost",
		23234,
		"nixpig",
		gossh.PublicKeys(signer),
		kh.Name(),
	)
	require.NoError(t, err)

	out := bytes.NewBufferString("")

	err = client.Run("cmd --option val subcmd", out)
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)

	require.Equal(t, "handled cmd --option val subcmd", out.String())
}

func testSSHClientRunCommandInvalidKnownHostsFile(t *testing.T) {
	sshServer := &glssh.Server{
		Addr: "localhost:23234",
		Handler: func(s glssh.Session) {
			s.Write([]byte(fmt.Sprintf("handled %s", s.RawCommand())))
		},
		PublicKeyHandler: func(ctx glssh.Context, publicKey glssh.PublicKey) bool {
			return true
		},
	}

	go func() {
		sshServer.ListenAndServe()
	}()

	t.Cleanup(func() {
		sshServer.Close()
	})

	_, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	client, err := ssh.NewSSHClient(
		"localhost",
		23234,
		"nixpig",
		gossh.PublicKeys(signer),
		"/invalid/known_hosts/path",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open knownhosts file")
	require.Nil(t, client)
}

func testSSHClientRunCommandKnownHostsKeyMismatch(t *testing.T) {
	sshServer := &glssh.Server{
		Addr: "localhost:23234",
		Handler: func(s glssh.Session) {
			s.Write([]byte(fmt.Sprintf("handled %s", s.RawCommand())))
		},
		PublicKeyHandler: func(ctx glssh.Context, publicKey glssh.PublicKey) bool {
			return true
		},
	}

	go func() {
		sshServer.ListenAndServe()
	}()

	t.Cleanup(func() {
		sshServer.Close()
	})

	publicKey, privateKey, err := test.GenerateKeyPair()
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	kh, err := os.CreateTemp("", "known_hosts")
	require.NoError(t, err)
	kh.WriteString(fmt.Sprintf(
		"[localhost]:23234 %s",
		string(gossh.MarshalAuthorizedKey(publicKey)),
	))

	client, err := ssh.NewSSHClient(
		"localhost",
		23234,
		"nixpig",
		gossh.PublicKeys(signer),
		kh.Name(),
	)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "knownhosts: key mismatch")
}
