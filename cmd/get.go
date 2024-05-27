package cmd

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	internal "github.com/nixpig/syringe.sh/internal/variables"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get [flags] VARIABLE_KEY",
	Aliases: []string{"g"},
	Short:   "Get an environment variable",
	Long:    `Get an environment variable for the current or specified project and environment.`,
	Example: `  syringe get DB_PASSWORD
  syringe get --env dev DB_PASSWORD
  syringe g -p dunce -e dev DB_PASSWORD`,
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
		handler := internal.NewVariableCliHandler(store, validator.New())

		variable, err := handler.Get(projectName, environmentName, variableKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting variable.%s\n", err)
			os.Exit(1)
		}

		fmt.Fprint(os.Stdout, variable)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("project", "p", "", "Project name")
	getCmd.Flags().StringP("environment", "e", "", "Environment name")
}
