package cmd

import (
	"context"
	"database/sql"
	"io"

	"github.com/charmbracelet/ssh"
	"github.com/spf13/cobra"
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
	db *sql.DB,
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

	ctx = context.WithValue(ctx, dbCtxKey, db)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

func walk(c *cobra.Command, f func(*cobra.Command)) {
	f(c)
	for _, c := range c.Commands() {
		walk(c, f)
	}
}
