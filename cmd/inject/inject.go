package inject

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/cmd/secret"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

func InjectCommand() *cobra.Command {
	injectCmd := &cobra.Command{
		Use:               "inject",
		Aliases:           []string{"i"},
		Short:             "Inject secrets",
		Example:           "syringe inject -p my_cool_project -e dev ./startserver",
		PersistentPreRunE: initInjectContext,
		RunE:              injectRunE,
	}

	injectCmd.Flags().StringP("project", "p", "", "Project name")
	injectCmd.MarkFlagRequired("project")

	injectCmd.Flags().StringP("environment", "e", "", "Environment name")
	injectCmd.MarkFlagRequired("environment")

	return injectCmd
}

func injectRunE(cmd *cobra.Command, args []string) error {
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

func initInjectContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.DB).(*sql.DB)
	if !ok {
		return fmt.Errorf("failed to get database from context")
	}

	secretStore := secret.NewSqliteSecretStore(db)
	secretService := secret.NewSecretServiceImpl(
		secretStore,
		validation.NewValidator(),
	)

	ctx = context.WithValue(ctx, ctxkeys.SecretService, secretService)

	cmd.SetContext(ctx)

	return nil
}
