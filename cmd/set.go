package cmd

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	internal "github.com/nixpig/syringe.sh/internal/variables"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set an environment variable.",
	Long: `Set an environment variable against the current project and environment.

Examples:
  syringe set DB_PASSWORD p4ssw0rd
  syringe set -p dunce -e dev DB_PASSWORD p4ssw0rd
	`,
	Args: cobra.MatchAll(cobra.ExactArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		projectName, err := cmd.Flags().GetString("project")
		if err != nil {
			fmt.Println("no project provided")
			os.Exit(1)
		}

		environmentName, err := cmd.Flags().GetString("environment")
		if err != nil {
			fmt.Println("no environment provided")
			os.Exit(1)
		}

		secret, err := cmd.Flags().GetBool("secret")
		if err != nil {
			fmt.Println("unable to get secret value")
			os.Exit(1)
		}

		variableKey := args[0]
		variableValue := args[1]

		store := internal.NewVariableStoreSqlite(DB)
		handler := internal.NewVariableCliHandler(store, validator.New())

		err = handler.Set(
			projectName,
			environmentName,
			variableKey,
			variableValue,
			secret,
		)
		if err != nil {
			fmt.Println("error setting variable: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().StringP("project", "p", "", "Project name")
	setCmd.Flags().StringP("environment", "e", "", "Environment name")
	setCmd.Flags().BoolP("secret", "s", false, "Variable is secret")
}
