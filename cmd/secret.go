package cmd

import (
	"context"
	"database/sql"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/spf13/cobra"
)

func secretCommand() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:               "secret",
		Aliases:           []string{"s"},
		Short:             "Secret",
		Long:              "Secret",
		Example:           "syringe secret",
		PersistentPreRunE: initSecretContext,
	}

	secretCmd.AddCommand(secretSetCommand())
	secretCmd.AddCommand(secretGetCommand())

	return secretCmd
}

func secretSetCommand() *cobra.Command {
	secretSetCmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"s"},
		Short:   "set",
		Long:    "set",
		Example: "syringe secret set []",
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

			secretService := cmd.Context().Value("SECRET_SERVICE").(services.SecretService)

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
		Use:     "get",
		Aliases: []string{"g"},
		Short:   "get",
		Long:    "get",
		Example: "syringe get []",
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

			secretService := cmd.Context().Value("SECRET_SERVICE").(services.SecretService)

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

	db := ctx.Value(DB_CTX).(*sql.DB)

	secretStore := stores.NewSqliteSecretStore(db)
	secretService := services.NewSecretServiceImpl(secretStore, validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, "SECRET_SERVICE", secretService)

	cmd.SetContext(ctx)

	return nil
}
