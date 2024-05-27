package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var injectCmd = &cobra.Command{
	Use:     "inject [flags] COMMAND",
	Aliases: []string{"i"},
	Short:   "Inject environment variables into command execution",
	Long:    `Inject environment variables for the current or specified project and environment into a command.`,
	Example: `  syringe inject server
  syringe inject -p dunce -e dev server
  syringe i -v DB_USERNAME -v DB_PASSWORD server`,
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

		variables, err := cmd.Flags().GetStringSlice("variable")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read 'variable' flag(s).\n%s", err)
			os.Exit(1)
		}

		fmt.Println(projectName)
		fmt.Println(environmentName)
		fmt.Println(variables)
	},
}

func init() {
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringP("project", "p", "", "Project name")
	injectCmd.Flags().StringP("environment", "e", "", "Environment name")

	var variables []string
	injectCmd.Flags().StringSliceVarP(&variables, "variable", "v", []string{}, "Variable keys")
}
