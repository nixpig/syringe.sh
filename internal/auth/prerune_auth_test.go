package auth_test

import (
	"context"
	"testing"

	"github.com/nixpig/syringe.sh/internal/auth"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestAuthPreRunE(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test auth prerun authenticated":     testAuthPreRunAuthenticated,
		"test auth prerun not authenticated": testAuthPreRunNotAuthenticated,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}

}

func testAuthPreRunAuthenticated(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.Authenticated, true)

	cmd.SetContext(ctx)

	err := auth.PreRunEAuth(cmd, args)
	require.NoError(t, err)
}

func testAuthPreRunNotAuthenticated(t *testing.T) {
	cmd := &cobra.Command{}
	args := []string{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.Authenticated, false)

	cmd.SetContext(ctx)

	err := auth.PreRunEAuth(cmd, args)
	require.EqualError(t, err, "not authenticated")
}
