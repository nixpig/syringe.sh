package secret

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewSSHHandlerSecretSet(secretService SecretService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

		if err := secretService.Set(SetSecretRequest{
			Project:     project,
			Environment: environment,
			Key:         key,
			Value:       value,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Secret '%s' set.", key))

		return nil
	}
}

func NewSSHHandlerSecretGet(secretService SecretService) pkg.CobraHandler {
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

		cmd.Println(secret.Value)

		return nil
	}
}

func NewSSHHandlerSecretList(secretService SecretService) pkg.CobraHandler {
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

func NewSSHHandlerSecretRemove(secretService SecretService) pkg.CobraHandler {
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

		cmd.Println("")

		return nil
	}
}

func NewSSHHandlerSecretInject(secretService SecretService) pkg.CobraHandler {
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

		injectableSecrets := strings.Join(secretsList, " ")

		cmd.Print(injectableSecrets)

		return nil
	}
}
