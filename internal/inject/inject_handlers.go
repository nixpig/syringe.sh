package inject

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewHandlerInject(secretService secret.SecretService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		project, _ := cmd.Flags().GetString("project")
		environment, _ := cmd.Flags().GetString("environment")

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
}
