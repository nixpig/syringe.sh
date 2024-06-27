package inject

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
)

func InjectCmdHandler(cmd *cobra.Command, args []string) error {
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService, ok := cmd.Context().Value(ctxkeys.SecretService).(secret.SecretService)
	if !ok {
		return fmt.Errorf("unable to get secret service from context")
	}

	secrets, err := secretService.List(secret.ListSecretsRequest{
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

	cmd.Println(injectableSecrets)

	return nil
}
