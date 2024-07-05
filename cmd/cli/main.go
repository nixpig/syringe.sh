package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/internal/cli"
	"github.com/nixpig/syringe.sh/internal/environment"
	"github.com/nixpig/syringe.sh/internal/inject"
	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/internal/root"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/internal/user"
	"github.com/nixpig/syringe.sh/pkg/helpers"
	"github.com/spf13/cobra"
)

const (
	host = "localhost"
	port = 23234
)

func main() {
	cmdRoot := root.New(context.Background())

	handlerCLI := cli.NewHandlerCLI(host, port, cmdRoot.OutOrStdout())
	handlerInjectCLI := cli.NewHandlerInjectCLI(host, port, cmdRoot.OutOrStdout())

	cmdRoot.PersistentFlags().StringP("identity", "i", "", "Path to SSH key (optional).\nIf not provided, SSH agent is used and syringe.sh host must be configured in SSH config.")

	// -- project
	cmdProject := project.NewCmdProject()
	cmdProject.AddCommand(project.NewCmdProjectList(handlerCLI))
	cmdProject.AddCommand(project.NewCmdProjectAdd(handlerCLI))
	cmdProject.AddCommand(project.NewCmdProjectRename(handlerCLI))
	cmdProject.AddCommand(project.NewCmdProjectRemove(handlerCLI))
	cmdRoot.AddCommand(cmdProject)

	// -- environment
	cmdEnvironment := environment.NewCmdEnvironment()
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentList(handlerCLI))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentAdd(handlerCLI))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentRename(handlerCLI))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentRemove(handlerCLI))
	cmdRoot.AddCommand(cmdEnvironment)

	// -- secret
	cmdSecret := secret.NewCmdSecret()

	cmdSecretSet := secret.NewCmdSecretSet(handlerCLI)
	cmdSecretSet.PreRunE = func(cmd *cobra.Command, args []string) error {
		// encrypt in here
		return nil
	}
	cmdSecret.AddCommand(cmdSecretSet)

	cmdSecretGet := secret.NewCmdSecretGet(handlerCLI)
	cmdSecretGet.PreRunE = func(cmd *cobra.Command, args []string) error {
		// decrypt in here
		return nil
	}
	cmdSecret.AddCommand(cmdSecretGet)

	cmdSecret.AddCommand(secret.NewCmdSecretList(handlerCLI))
	cmdSecret.AddCommand(secret.NewCmdSecretRemove(handlerCLI))
	cmdRoot.AddCommand(cmdSecret)

	// -- user
	cmdUser := user.NewCmdUser()
	cmdUser.AddCommand(user.NewCmdUserRegister(handlerCLI))
	cmdRoot.AddCommand(cmdUser)

	cmdInject := inject.NewCmdInject(handlerInjectCLI)
	cmdRoot.AddCommand(cmdInject)

	// -- update help and version for all subcommands
	helpers.WalkCmd(cmdRoot, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
		c.Flags().BoolP("version", "v", false, "Print version information")
	})

	if err := cmdRoot.Execute(); err != nil {
		os.Exit(1)
	}
}
