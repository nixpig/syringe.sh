package cmd

import (
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/spf13/cobra"
)

func secretCommand() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "Secret",
		Long:    "Secret",
		Example: "syringe secret",
	}

	secretCmd.AddCommand(setCommand())
	secretCmd.AddCommand(getCommand())

	return secretCmd
}

func setCommand() *cobra.Command {
	setCmd := &cobra.Command{
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

			envService := cmd.Context().Value("ENV_SERVICE").(services.UserService)

			if err := envService.SetSecret(services.SetSecretRequest{
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

	setCmd.Flags().StringP("project", "p", "", "Project to use")
	setCmd.Flags().StringP("environment", "e", "", "Environment to use")
	setCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return setCmd
}

func getCommand() *cobra.Command {
	getCmd := &cobra.Command{
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

			envService := cmd.Context().Value("ENV_SERVICE").(services.UserService)

			secret, err := envService.GetSecret(services.GetSecretRequest{
				Project:     project,
				Environment: environment,
				Key:         key,
			})
			if err != nil {
				return err
			}

			sess := cmd.Context().Value("SESS").(ssh.Session)

			fmt.Println("SECRET:\n", secret)
			wish.Println(sess, "SECRET", secret)

			return nil
		},
	}

	getCmd.Flags().StringP("project", "p", "", "Project")
	getCmd.Flags().StringP("environment", "e", "", "Environment")
	getCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return getCmd
}
