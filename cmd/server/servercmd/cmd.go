package servercmd

import (
	"context"
	"database/sql"
	"io"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
	// "github.com/spf13/cobra/doc"
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
		Use:   "syringe",
		Short: "🔐 Distributed database-per-user encrypted secrets management over SSH protocol.",
		Long:  "🔐 Distributed database-per-user encrypted secrets management over SSH protocol.",

		Example: `  Register user:
    syringe user register

  Add a project:
    syringe project add my_cool_project

  Add an environment:
    syringe environment add -p my_cool_project dev

  Add a secret:
    syringe secret set -p my_cool_project -e dev SECRET_KEY secret_value

  List secrets:
    syringe secret list -p my_cool_project -e dev

  Inject into command:
    syringe inject -p my_cool_project -e dev ./startserver

  For more examples, go to https://syringe.sh/examples`,

		// SilenceErrors: true,
	}

	additionalHelp := `
Please note: some commands are only available (and listed above) once registered and authenticated. Trying to use one of these while not authenticated will result in an 'unknown command' error.

For more help on how to use Syringe, go to https://syringe.sh/help

`

	rootCmd.SetHelpTemplate(rootCmd.HelpTemplate() + additionalHelp)

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

	ctx = context.WithValue(ctx, ctxkeys.DB, db)

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
