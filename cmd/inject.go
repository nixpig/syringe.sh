package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/spf13/cobra"
)

var injectCmd = &cobra.Command{
	Use:                "inject [flags] COMMAND",
	Aliases:            []string{"i"},
	Short:              "Inject environment variables into command execution",
	Long:               `Inject environment variables for the current or specified project and environment into a command.`,
	DisableFlagParsing: true,
	Example:            `  syringe inject server`,
	Args:               cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		variables := []string{} // GET VARIABLES FOR LINKED PROJECT/ENV

		command.Env = slices.Concat(env, variables)

		fmt.Println("env: ", command.Env)

		command.Run()
	},
}

func init() {
	rootCmd.AddCommand(injectCmd)
}
