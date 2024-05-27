package cmd

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	internal "github.com/nixpig/syringe.sh/internal/variables"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove [flags] VARIABLE_KEY",
	Aliases: []string{"r"},
	Short:   "Remove an environment variable",
	Long:    `Remove an environment variable against the current or specified project and environment.`,
	Example: `  syringe remove DB_PASSWORD
  syringe remove --env dev DB_PASSWORD
  syringe r -p dunce -e dev DB_PASSWORD`,
	Args: cobra.MatchAll(cobra.ExactArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		projectName, err := cmd.Flags().GetString("project")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read 'project' flag.\n%s", err)
			os.Exit(1)
		}

		environmentName, err := cmd.Flags().GetString("environment")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read 'environment' flag.\n%s", err)
			os.Exit(1)
		}

		variableKey := args[0]

		store := internal.NewVariableStoreSqlite(DB)
		internal.NewVariableCliHandler(store, validator.New())

		err = store.Delete(projectName, environmentName, variableKey)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to delete variable: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	removeCmd.Flags().StringP("project", "p", "", "Project name")
	removeCmd.Flags().StringP("environment", "e", "", "Environment name")
}
