package cmd

import (
	"context"
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

func environmentCommand(sess ssh.Session) *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:                "environment",
		Aliases:            []string{"e"},
		Short:              "Environment",
		Long:               "Environment",
		Example:            "syringe environment",
		PersistentPreRunE:  initEnvironmentContext(sess),
		PersistentPostRunE: closeEnvironmentContext(sess),
	}

	environmentCmd.AddCommand(environmentAddCommand())

	return environmentCmd
}

func environmentAddCommand() *cobra.Command {
	environmentAddCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "add",
		Long:    "add",
		Example: "syringe environment add []",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {

			environment := args[0]

			project, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}

			environmentService := cmd.Context().Value("ENVIRONMENT_SERVICE").(services.EnvironmentService)

			if err := environmentService.AddEnvironment(services.AddEnvironmentRequest{
				Name:        environment,
				ProjectName: project,
			}); err != nil {
				return err
			}

			return nil
		},
	}

	environmentAddCmd.Flags().StringP("project", "p", "", "Project")

	return environmentAddCmd
}

func initEnvironmentContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		ctx := cmd.Context()
		// add ssh sess to cobra cmd context
		// REVIEW: is this even used??
		ctx = context.WithValue(ctx, "sess", sess)

		// add user database to cobra cmd context
		api := turso.New(
			os.Getenv("DATABASE_ORG"),
			os.Getenv("API_TOKEN"),
			http.Client{},
		)

		marshalledKey := gossh.MarshalAuthorizedKey(sess.PublicKey())

		hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))

		token, err := api.CreateToken(hashedKey, "30s")
		if err != nil {
			fmt.Println("failed to create token:", err)
		}

		db, err := database.Connection(
			"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
			string(token.Jwt),
		)
		if err != nil {
			fmt.Println("error creating database connection:\n", err)
			return nil
		}

		environmentStore := stores.NewSqliteEnvironmentStore(db)
		environmentService := services.NewEnvironmentServiceImpl(environmentStore, validator.New(validator.WithRequiredStructEnabled()))

		ctx = context.WithValue(ctx, "ENVIRONMENT_SERVICE", environmentService)

		cmd.SetContext(ctx)

		return nil
	}
}

func closeEnvironmentContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}
