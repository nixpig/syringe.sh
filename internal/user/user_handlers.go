package user

import (
	"fmt"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewHandlerUserRegister(userService UserService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		username, ok := cmd.Context().Value(ctxkeys.Username).(string)
		if !ok {
			return fmt.Errorf("unable to get username from context")
		}

		publicKey, ok := cmd.Context().Value(ctxkeys.PublicKey).(ssh.PublicKey)
		if !ok {
			return fmt.Errorf("unable to get public key from context")
		}

		user, err := userService.RegisterUser(RegisterUserRequest{
			Username:  username,
			Email:     "not_used_yet@example.org",
			PublicKey: publicKey,
		})
		if err != nil {
			return fmt.Errorf("unable to register user: %w", err)
		}

		cmd.Print(user)

		return nil
	}
}
