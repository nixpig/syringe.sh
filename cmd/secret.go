package cmd

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/spf13/cobra"
)

const (
	secretCtxKey = contextKey("SECRET_CTX")
)

func secretCommand() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:               "secret",
		Aliases:           []string{"s"},
		Short:             "Manage secrets",
		PersistentPreRunE: initSecretContext,
	}

	secretCmd.AddCommand(secretSetCommand())
	secretCmd.AddCommand(secretGetCommand())

	return secretCmd
}

func secretSetCommand() *cobra.Command {
	secretSetCmd := &cobra.Command{
		Use:     "set [flags] SECRET_KEY SECRET_VALUE",
		Aliases: []string{"s"},
		Short:   "Set a secret",
		Example: "syringe secret set -p my_cool_project -e local AWS_ACCESS_KEY_ID AKIAIOSFODNN7EXAMPLE",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if err := secretService.SetSecret(services.SetSecretRequest{
				Project:     project,
				Environment: environment,
				Key:         key,
				Value:       value,
			}); err != nil {
				return err
			}

			return nil
		},
	}

	secretSetCmd.Flags().StringP("project", "p", "", "Project to use")
	secretSetCmd.Flags().StringP("environment", "e", "", "Environment to use")
	secretSetCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	secretSetCmd.MarkFlagRequired("project")
	secretSetCmd.MarkFlagRequired("environment")

	return secretSetCmd
}

func secretGetCommand() *cobra.Command {
	secretGetCmd := &cobra.Command{
		Use:     "get [flags] SECRET_KEY",
		Aliases: []string{"g"},
		Short:   "Get a secret",
		Example: "syringe get -p my_cool_project -e local AWS_ACCESS_KEY_ID",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			secret, err := secretService.GetSecret(services.GetSecretRequest{
				Project:     project,
				Environment: environment,
				Key:         key,
			})
			if err != nil {
				return err
			}

			cmd.Print("cobra secret: ", secret)

			return nil
		},
	}

	secretGetCmd.Flags().StringP("project", "p", "", "Project")
	secretGetCmd.Flags().StringP("environment", "e", "", "Environment")
	secretGetCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	secretGetCmd.MarkFlagRequired("project")
	secretGetCmd.MarkFlagRequired("environment")

	return secretGetCmd
}

func initSecretContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db := ctx.Value(dbCtxKey).(*sql.DB)

	secretStore := stores.NewSqliteSecretStore(db)
	secretService := services.NewSecretServiceImpl(secretStore, validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, secretCtxKey, secretService)

	cmd.SetContext(ctx)

	return nil
}
