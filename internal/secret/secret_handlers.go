package secret

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
)

func SetCmdHandler(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService, ok := cmd.Context().Value(ctxkeys.SecretService).(SecretService)
	if !ok {
		return fmt.Errorf("unable to get secret service from context")
	}

	if err := secretService.Set(SetSecretRequest{
		Project:     project,
		Environment: environment,
		Key:         key,
		Value:       value,
	}); err != nil {
		return err
	}

	cmd.Print("")
	return nil
}

func GetCmdHandler(cmd *cobra.Command, args []string) error {
	key := args[0]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService := cmd.Context().Value(ctxkeys.SecretService).(SecretService)

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

func ListCmdHandler(cmd *cobra.Command, args []string) error {
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService, ok := cmd.Context().Value(ctxkeys.SecretService).(SecretService)
	if !ok {
		return fmt.Errorf("unable to load secret service from context")
	}

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

func RemoveCmdHandler(cmd *cobra.Command, args []string) error {
	key := args[0]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService, ok := cmd.Context().Value(ctxkeys.SecretService).(SecretService)
	if !ok {
		return fmt.Errorf("unable to get secret service from context")
	}

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
