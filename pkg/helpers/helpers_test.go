package helpers_test

import (
	"testing"

	"github.com/nixpig/syringe.sh/pkg/helpers"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestWalkCmd(t *testing.T) {
	cmdRoot := &cobra.Command{Use: "root"}
	cmdL1 := &cobra.Command{Use: "l1"}
	cmdL2 := &cobra.Command{Use: "l2"}

	cmdRoot.AddCommand(cmdL1)
	cmdL1.AddCommand(cmdL2)

	helpers.WalkCmd(cmdRoot, func(cmd *cobra.Command) {
		cmd.Use = cmd.Use + " walked"
	})

	require.Equal(t, "root walked", cmdRoot.Use)
	require.Equal(t, "l1 walked", cmdL1.Use)
	require.Equal(t, "l2 walked", cmdL2.Use)
}
