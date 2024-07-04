package auth

import (
	"errors"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
)

func PreRunE(cmd *cobra.Command, args []string) error {
	authenticated, ok := cmd.Context().Value(ctxkeys.Authenticated).(bool)
	if !ok || !authenticated {
		return errors.New("not authenticated")
	}

	return nil
}
