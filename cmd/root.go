package cmd

import (
	"context"
	"database/sql"
	"io"

	"github.com/nixpig/syringe.sh/server/pkg"
	"github.com/spf13/cobra"
	// "github.com/spf13/cobra/doc"
)

const (
	dbCtxKey = pkg.DBCtxKey
)

func Execute(
	commands []*cobra.Command,
	args []string,
	cmdIn io.Reader,
	cmdOut io.Writer,
	cmdErr io.ReadWriter,
	db *sql.DB,
) error {
	rootCmd := &cobra.Command{
		Use:           "syringe",
		Short:         "Distributed environment variable management over SSH.",
		Long:          "Distributed environment variable management over SSH.",
		SilenceErrors: true,
	}

	for _, command := range commands {
		rootCmd.AddCommand(command)
	}

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

	// f, err := os.Create("./docs.md")
	// if err != nil {
	// 	return err
	// }
	//
	// defer f.Close()
	//
	// walk(rootCmd, func(c *cobra.Command) {
	// 	if err := doc.GenMarkdown(c, f); err != nil {
	// 		panic("at the disco!")
	// 	}
	// })

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return pkg.FormattedError{Err: err}
	}

	return nil
}

func walk(c *cobra.Command, f func(*cobra.Command)) {
	f(c)
	for _, c := range c.Commands() {
		walk(c, f)
	}
}
