package ssh_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	glssh "github.com/gliderlabs/ssh"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/test"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func TestSSHClient(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test ssh client run command happy path": testSSHClientRunCommandHappyPath,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, fn)
	}
}

func testSSHClientRunCommandHappyPath(t *testing.T) {
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

	require.Equal(t, "handled cmd --option val subcmd", out.String())
}
