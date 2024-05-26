package cmd

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	internal "github.com/nixpig/syringe.sh/internal/variables"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MatchAll(cobra.ExactArgs(1)),
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

		variableKey := args[0]

		store := internal.NewVariableStoreSqlite(DB)
		handler := internal.NewVariableCliHandler(store, validator.New())

		variable, err := handler.Get(projectName, environmentName, variableKey)
		if err != nil {
			fmt.Println("error getting variable: ", err)
			os.Exit(1)
		}

		fmt.Println(variable)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("project", "p", "", "Project name")
	getCmd.Flags().StringP("environment", "e", "", "Environment name")
}
