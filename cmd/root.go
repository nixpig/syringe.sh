package cmd

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/pkg/turso"
	"github.com/spf13/cobra"
	gossh "golang.org/x/crypto/ssh"
)

type contextKey string

const (
	dbCtxKey   = contextKey("DB_CTX")
	sessCtxKey = contextKey("SESS_CTX")
)

func Execute(
	publicKey ssh.PublicKey,
	args []string,
	cmdIn io.Reader,
	cmdOut io.Writer,
	cmdErr io.ReadWriter,
) error {
	rootCmd := &cobra.Command{
		Use:   "syringe",
		Short: "Distributed environment variable management over SSH.",
		Long:  "Distributed environment variable management over SSH.",
	}

	rootCmd.AddCommand(userCommand())
	rootCmd.AddCommand(projectCommand())
	rootCmd.AddCommand(environmentCommand())
	rootCmd.AddCommand(secretCommand())

	rootCmd.SetArgs(args)
	rootCmd.SetIn(cmdIn)
	rootCmd.SetOut(cmdOut)
	rootCmd.SetErr(cmdErr)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	walk(rootCmd, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, "Help for the "+c.Name()+" command")
	})

	ctx := context.Background()

	db, err := NewUserDBConnection(publicKey)
	if err != nil {
		return err
	}

	defer db.Close()

	ctx = context.WithValue(ctx, dbCtxKey, db)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

// TODO: really don't like this!!
func NewUserDBConnection(publicKey ssh.PublicKey) (*sql.DB, error) {
	api := turso.New(
		os.Getenv("DATABASE_ORG"),
		os.Getenv("API_TOKEN"),
		http.Client{},
	)

	marshalledKey := gossh.MarshalAuthorizedKey(publicKey)

	hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))
	expiration := "30s"

	token, err := api.CreateToken(hashedKey, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to create token:\n%s", err)
	}

	fmt.Println("creating new user-specific db connection")
	db, err := database.Connection(
		"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
		string(token.Jwt),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection:\n%s", err)
	}

	return db, nil
}

func walk(c *cobra.Command, f func(*cobra.Command)) {
	f(c)
	for _, c := range c.Commands() {
		walk(c, f)
	}
}
