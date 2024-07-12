package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/internal/cli"
	"github.com/nixpig/syringe.sh/internal/environment"
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
	os.Setenv("APP_HOST", "localhost")
	cmdRoot := root.New(context.Background())

	handlerCLI := cli.NewHandlerCLI(host, port, cmdRoot.OutOrStdout())
	handlerInjectCLI := cli.NewHandlerInjectCLI(host, port, cmdRoot.OutOrStdout())

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
	cmdSecretSet.PreRunE = cli.PreRunEEncrypt
	cmdSecret.AddCommand(cmdSecretSet)

	cmdSecretList := secret.NewCmdSecretList(handlerCLI)
	cmdSecret.AddCommand(cmdSecretList)

	cmdSecretInject := secret.NewCmdSecretInject(handlerInjectCLI)
	cmdSecret.AddCommand(cmdSecretInject)

	cmdSecret.AddCommand(secret.NewCmdSecretGet(handlerCLI))

	cmdSecret.AddCommand(secret.NewCmdSecretRemove(handlerCLI))

	cmdRoot.AddCommand(cmdSecret)

	// -- user
	cmdUser := user.NewCmdUser()
	cmdUser.AddCommand(user.NewCmdUserRegister(handlerCLI))
	cmdRoot.AddCommand(cmdUser)

	// update help and version for all subcommands
	helpers.WalkCmd(cmdRoot, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
		c.Flags().BoolP("version", "v", false, "Print version information")
	})

	if err := cmdRoot.Execute(); err != nil {
		os.Exit(1)
	}
}
