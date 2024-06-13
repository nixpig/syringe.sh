package cmd

import (
	"context"
	"crypto/sha1"
	"database/sql"
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

func Execute(
	sess ssh.Session,
	appService services.AppService,
) error {
	rootCmd := &cobra.Command{
		Use:                "syringe",
		Short:              "A terminal-based utility to securely manage environment variables across projects and environments.",
		Long:               "A terminal-based utility to securely manage environment variables across projects and environments.",
		PersistentPreRunE:  initRootContext(sess),
		PersistentPostRunE: closeRootContext(sess),
	}

	rootCmd.AddCommand(NewRegisterCommand(sess, appService))
	rootCmd.AddCommand(projectCommand(sess))
	rootCmd.AddCommand(environmentCommand(sess))
	rootCmd.AddCommand(secretCommand())

	rootCmd.SetArgs(sess.Command())
	rootCmd.SetIn(sess)
	rootCmd.SetOut(sess)
	rootCmd.SetErr(sess.Stderr())
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

func initRootContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
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

		ctx = context.WithValue(ctx, "DB_CONN", db)

		envStore := stores.NewSqliteEnvStore(db)
		envService := services.NewSecretServiceImpl(envStore, validator.New(validator.WithRequiredStructEnabled()))

		ctx = context.WithValue(ctx, "ENV_SERVICE", envService)

		cmd.SetContext(ctx)

		return nil
	}
}

func closeRootContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		db := cmd.Context().Value("DB_CONN").(*sql.DB)

		if err := db.Close(); err != nil {
			fmt.Println("error closing db connection")
			return err
		}

		return nil
	}
}