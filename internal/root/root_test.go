package root_test

import (
	"context"
	"testing"

	"github.com/nixpig/syringe.sh/internal/root"
	"github.com/stretchr/testify/require"
)

// persistence of identity flag
// initialisation of config in prerun
// check passed in context is set

func TestCmdRoot(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test root command happy path": testRootCmdHappyPath,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, fn)
	}
}

func testRootCmdHappyPath(t *testing.T) {
	ctx := context.Background()
	cmdRoot := root.New(ctx)

	cmdRoot.SetArgs([]string{
		"-i",
		"identity_file_path",
	})

	err := cmdRoot.Execute()
	require.NoError(t, err)

	identity, err := cmdRoot.PersistentFlags().GetString("identity")
	require.NoError(t, err)
	require.Equal(t, "identity_file_path", identity)

	// todo: fill out tests...
}
