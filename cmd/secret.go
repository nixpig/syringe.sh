package cmd

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg/turso"
	"github.com/spf13/cobra"
	gossh "golang.org/x/crypto/ssh"
)

func NewSecretCommand(
	sess ssh.Session,
) *cobra.Command {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"s"},
		Short:   "Secret",
		Long:    "Secret",
		Example: "syringe secret",
	}

	api := turso.New(
		os.Getenv("DATABASE_ORG"),
		os.Getenv("API_TOKEN"),
		http.Client{},
	)

	fmt.Println(
		os.Getenv("DATABASE_ORG"),
		os.Getenv("API_TOKEN"),
	)

	marshalledKey := gossh.MarshalAuthorizedKey(sess.PublicKey())

	hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))

	token, err := api.CreateToken(hashedKey, "30s")
	if err != nil {
		fmt.Println("failed to create token")
	}

	db, err := database.Connection(
		"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
		string(token.Jwt),
	)
	if err != nil {
		fmt.Println("error creating database connection:\n", err)
		return nil
	}

	fmt.Println("db stats: ", db.Stats())
	envStore := stores.NewSqliteEnvStore(db)
	envService := services.NewEnvServiceImpl(envStore, validator.New(validator.WithRequiredStructEnabled()))

	secretCmd.AddCommand(NewSecretSetCmd(envService))

	return secretCmd
}

func NewSecretSetCmd(envService services.EnvService) *cobra.Command {
	fmt.Println("create new command")
	var secretSetCmd = &cobra.Command{
		Use:     "set",
		Aliases: []string{"s"},
		Short:   "set",
		Long:    "set",
		Example: "syringe secret set []",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			project, err := cmd.Flags().GetString("project")
			if err != nil {
				fmt.Println("unable to get secret set PROJECT flag:\n", err)
				return
			}

			environment, err := cmd.Flags().GetString("environment")
			if err != nil {
				fmt.Println("unable to get secret set ENVIRONMENT flag:\n", err)
				return
			}

			fmt.Println("in the SET command")
			if err := envService.SetSecret(services.SetSecretRequest{
				Project:     project,
				Environment: environment,
				Key:         key,
				Value:       value,
			}); err != nil {
				fmt.Println("failed to set secret:\n", err)
				return
			}
		},
	}

	secretSetCmd.Flags().StringP("project", "p", "", "Project to use")
	secretSetCmd.Flags().StringP("environment", "e", "", "Environment to use")
	secretSetCmd.Flags().BoolP("secret", "s", false, "Is secret?")

	return secretSetCmd
}
