package cli

import (
	"fmt"

	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
)

func PreRunESecretSet(cmd *cobra.Command, args []string) error {
	identity, err := cmd.Flags().GetString("identity")
	if err != nil {
		return err
	}

	publicKey, err := ssh.GetPublicKey(fmt.Sprintf("%s.pub", identity))
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	encryptedSecret, err := ssh.Encrypt(args[1], publicKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	args[1] = encryptedSecret
	return nil
}
