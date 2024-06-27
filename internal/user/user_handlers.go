package user

import (
	"fmt"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func RegisterCmdHandler(cmd *cobra.Command, args []string) error {
	username, ok := cmd.Context().Value(ctxkeys.Username).(string)
	if !ok {
		return fmt.Errorf("unable to get username from context")
	}

	fmt.Println("username: ", username)

	publicKey, ok := cmd.Context().Value(ctxkeys.PublicKey).(ssh.PublicKey)
	if !ok {
		return fmt.Errorf("unable to get public key from context")
	}

	fmt.Println("publicKey: ", publicKey)

	userService, ok := cmd.Context().Value(ctxkeys.UserService).(UserService)
	if !ok {
		return fmt.Errorf("unable to get user service from context")
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
