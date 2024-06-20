package secret

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg"
	"github.com/spf13/cobra"
)

const (
	dbCtxKey     = pkg.DBCtxKey
	secretCtxKey = pkg.SecretCtxKey
)

func SecretCommand() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:               "secret",
		Aliases:           []string{"s"},
		Short:             "Manage secrets",
		PersistentPreRunE: initSecretContext,
	}

	secretCmd.AddCommand(secretSetCommand())
	secretCmd.AddCommand(secretGetCommand())
	secretCmd.AddCommand(secretListCommand())

	return secretCmd
}

func secretSetCommand() *cobra.Command {
	secretSetCmd := &cobra.Command{
		Use:     "set [flags] SECRET_KEY SECRET_VALUE",
		Aliases: []string{"s"},
		Short:   "Set a secret",
		Example: "syringe secret set -p my_cool_project -e local AWS_ACCESS_KEY_ID AKIAIOSFODNN7EXAMPLE",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    secretSetRunE,
	}

	secretSetCmd.Flags().StringP("project", "p", "", "Project to use")
	secretSetCmd.MarkFlagRequired("project")

	secretSetCmd.Flags().StringP("environment", "e", "", "Environment to use")
	secretSetCmd.MarkFlagRequired("environment")

	secretSetCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return secretSetCmd
}

func secretSetRunE(cmd *cobra.Command, args []string) error {
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

	secretService := cmd.Context().Value(secretCtxKey).(services.SecretService)

	if err := secretService.Set(services.SetSecretRequest{
		Project:     project,
		Environment: environment,
		Key:         key,
		Value:       value,
	}); err != nil {
		return err
	}

	return nil
}

func secretGetCommand() *cobra.Command {
	secretGetCmd := &cobra.Command{
		Use:     "get [flags] SECRET_KEY",
		Aliases: []string{"g"},
		Short:   "Get a secret",
		Example: "syringe get -p my_cool_project -e local AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    secretGetRunE,
	}

	secretGetCmd.Flags().StringP("project", "p", "", "Project")
	secretGetCmd.MarkFlagRequired("project")

	secretGetCmd.Flags().StringP("environment", "e", "", "Environment")
	secretGetCmd.MarkFlagRequired("environment")

	secretGetCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return secretGetCmd
}

func secretGetRunE(cmd *cobra.Command, args []string) error {
	key := args[0]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService := cmd.Context().Value(secretCtxKey).(services.SecretService)

	secret, err := secretService.Get(services.GetSecretRequest{
		Project:     project,
		Environment: environment,
		Key:         key,
	})
	if err != nil {
		return err
	}

	cmd.Print(secret)

	return nil
}

func secretListCommand() *cobra.Command {
	secretListCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Example: "syringe secret list -p my_cool_project -e staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    secretListRunE,
	}

	secretListCmd.Flags().StringP("project", "p", "", "Project name")
	secretListCmd.MarkFlagRequired("project")

	secretListCmd.Flags().StringP("environment", "e", "", "Environment name")
	secretListCmd.MarkFlagRequired("environment")

	return secretListCmd
}

func secretListRunE(cmd *cobra.Command, args []string) error {
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	secretService, ok := cmd.Context().Value(secretCtxKey).(services.SecretService)
	if !ok {
		return fmt.Errorf("unable to load secret service from context")
	}

	secrets, err := secretService.List(services.ListSecretsRequest{
		Project:     project,
		Environment: environment,
	})
	if err != nil {
		return err
	}

	for _, s := range secrets.Secrets {
		cmd.Println(s.ID, s.Key, s.Value)
	}

	return nil
}

func initSecretContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(dbCtxKey).(*sql.DB)
	if !ok {
		return fmt.Errorf("unable to get database from context")
	}

	secretStore := stores.NewSqliteSecretStore(db)
	secretService := services.NewSecretServiceImpl(secretStore, validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, secretCtxKey, secretService)

	cmd.SetContext(ctx)

	return nil
}
