package secret

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	myssh "github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewHandlerSecretSet(secretService SecretService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

		c, ok := cmd.Context().Value(ctxkeys.PublicKey).(ssh.PublicKey)
		if !ok {
			return errors.New("unable to get public key from context")
		}

		encryptedSecret, err := myssh.Encrypt(value, c)
		if err != nil {
			return err
		}

		if err := secretService.Set(SetSecretRequest{
			Project:     project,
			Environment: environment,
			Key:         key,
			Value:       encryptedSecret,
		}); err != nil {
			return err
		}

		cmd.Print("")
		return nil
	}
}

func NewHandlerSecretGet(secretService SecretService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		key := args[0]

		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

		secret, err := secretService.Get(GetSecretRequest{
			Project:     project,
			Environment: environment,
			Key:         key,
		})
		if err != nil {
			return err
		}

		cmd.Print(secret.Value)

		return nil
	}
}

func NewHandlerSecretList(secretService SecretService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

		secrets, err := secretService.List(ListSecretsRequest{
			Project:     project,
			Environment: environment,
		})
		if err != nil {
			return err
		}

		secretsList := make([]string, len(secrets.Secrets))
		for i, s := range secrets.Secrets {
			secretsList[i] = fmt.Sprintf("%s=%s", s.Key, s.Value)
		}

		cmd.Print(strings.Join(secretsList, "\n"))

		return nil
	}
}

func NewHandlerSecretRemove(secretService SecretService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		key := args[0]

		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

		if err := secretService.Remove(RemoveSecretRequest{
			Project:     project,
			Environment: environment,
			Key:         key,
		}); err != nil {
			return err
		}

		cmd.Print("")

		return nil
	}
}
