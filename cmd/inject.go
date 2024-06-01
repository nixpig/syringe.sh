package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/go-playground/validator/v10"
	internal "github.com/nixpig/syringe.sh/internal/variables"
	"github.com/spf13/cobra"
)

var injectCmd = &cobra.Command{
	Use:     "inject [flags] COMMAND",
	Aliases: []string{"i"},
	Short:   "Inject environment variables into command execution",
	Long:    `Inject environment variables for the current or specified project and environment into a command.`,
	Example: `  syringe inject server`,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName, err := cmd.Flags().GetString("project")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read 'project' flag.\n%s", err)
			os.Exit(1)
		}

		environmentName, err := cmd.Flags().GetString("environment")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read 'environment' flag.\n%s", err)
		}

		fmt.Println("projectName: ", projectName)
		fmt.Println("environmentName: ", environmentName)

		subcommand := args[0]
		subargs := []string{}

		if len(args) > 1 {
			subargs = args[1:]
		}

		command := exec.Command(subcommand)
		command.Args = slices.Concat(command.Args, subargs)
		command.Stderr = os.Stderr
		command.Stdout = os.Stdout
		command.Stdin = os.Stdin
		env := command.Environ()

		store := internal.NewVariableStoreSqlite(DB)
		handler := internal.NewVariableCliHandler(store, validator.New())

		variables, err := handler.GetAll(projectName, environmentName)

		command.Env = slices.Concat(env, variables)

		fmt.Println("injecting: ", variables)
		command.Run()
	},
}

func init() {
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringP("project", "p", "", "Project name")
	injectCmd.Flags().StringP("environment", "e", "", "Environment name")
}
