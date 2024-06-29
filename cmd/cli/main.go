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

	cmdRoot.PersistentFlags().StringP("identity", "i", "", "Path to SSH key (if not provided, SSH agent is used)")

	cmdProject := project.NewCmdProject()
	cmdProject.AddCommand(project.NewCmdProjectList(handlerCLI))
	cmdProject.AddCommand(project.NewCmdProjectAdd(handlerCLI))
	cmdProject.AddCommand(project.NewCmdProjectRename(handlerCLI))
	cmdProject.AddCommand(project.NewCmdProjectRemove(handlerCLI))
	cmdRoot.AddCommand(cmdProject)

	cmdEnvironment := environment.NewCmdEnvironment()
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentList(handlerCLI))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentAdd(handlerCLI))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentRename(handlerCLI))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentRemove(handlerCLI))
	cmdRoot.AddCommand(cmdEnvironment)

	cmdSecret := secret.NewCmdSecret()
	cmdSecret.AddCommand(secret.NewCmdSecretList(handlerCLI))
	cmdSecret.AddCommand(secret.NewCmdSecretSet(handlerCLI))
	cmdSecret.AddCommand(secret.NewCmdSecretGet(handlerCLI))
	cmdSecret.AddCommand(secret.NewCmdSecretRemove(handlerCLI))
	cmdRoot.AddCommand(cmdSecret)

	cmdUser := user.NewCmdUser()
	cmdUser.AddCommand(user.NewCmdUserRegister(handlerCLI))
	cmdRoot.AddCommand(cmdUser)

	cmdInject := inject.NewCmdInject(handlerInjectCLI)
	cmdRoot.AddCommand(cmdInject)

	helpers.WalkCmd(cmdRoot, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
		c.Flags().BoolP("version", "v", false, "Print version information")
	})

	if err := cmdRoot.Execute(); err != nil {
		os.Exit(1)
	}
}
