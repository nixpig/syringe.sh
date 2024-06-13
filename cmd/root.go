package cmd

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/pkg/turso"
	"github.com/spf13/cobra"
	gossh "golang.org/x/crypto/ssh"
)

type contextKey string

const (
	DB_CTX   = contextKey("DB_CTX")
	SESS_CTX = contextKey("SESS_CTX")
)

func Execute(
	sess ssh.Session,
	appService services.AppService,
) error {
	rootCmd := &cobra.Command{
		Use:   "syringe",
		Short: "A terminal-based utility to securely manage environment variables across projects and environments.",
		Long:  "A terminal-based utility to securely manage environment variables across projects and environments.",
		// PersistentPreRunE:  initRootContext(sess),
		// PersistentPostRunE: closeRootContext(sess),
	}

	rootCmd.AddCommand(userCommand(sess, appService))

	rootCmd.AddCommand(projectCommand())
	rootCmd.AddCommand(environmentCommand())
	rootCmd.AddCommand(secretCommand())

	rootCmd.SetArgs(sess.Command())
	rootCmd.SetIn(sess)
	rootCmd.SetOut(sess)
	rootCmd.SetErr(sess.Stderr())
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	ctx := context.Background()

	ctx = context.WithValue(ctx, SESS_CTX, sess)

	fmt.Println("creating new turso api")
	// add user database to cobra cmd context
	api := turso.New(
		os.Getenv("DATABASE_ORG"),
		os.Getenv("API_TOKEN"),
		http.Client{},
	)

	marshalledKey := gossh.MarshalAuthorizedKey(sess.PublicKey())

	hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))

	fmt.Println("creating new token")
	token, err := api.CreateToken(hashedKey, "30s")
	if err != nil {
		fmt.Println("failed to create token:", err)
	}

	fmt.Println("creating new user-specific db connection")
	db, err := database.Connection(
		"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
		string(token.Jwt),
	)
	if err != nil {
		fmt.Println("error creating database connection:\n", err)
		return nil
	}

	fmt.Println("adding db to context")
	ctx = context.WithValue(ctx, DB_CTX, db)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

// func initRootContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
// 	return func(cmd *cobra.Command, args []string) error {
// 		ctx := cmd.Context()
//
// 		cmd.SetContext(ctx)
//
// 		return nil
// 	}
// }
//
// func closeRootContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
// 	return func(cmd *cobra.Command, args []string) error {
// 		db := cmd.Context().Value(DB_CTX).(*sql.DB)
//
// 		if err := db.Close(); err != nil {
// 			fmt.Println("error closing db connection")
// 			return err
// 		}
//
// 		return nil
// 	}
// }
