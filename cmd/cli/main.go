package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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

	handler := newCliHandler(cmdRoot.OutOrStdout())

	cmdRoot.PersistentFlags().StringP("identity", "i", "", "Path to SSH key (if not provided, SSH agent is used)")

	cmdProject := project.NewCmdProject()
	cmdProject.AddCommand(project.NewCmdProjectList(handler))
	cmdProject.AddCommand(project.NewCmdProjectAdd(handler))
	cmdProject.AddCommand(project.NewCmdProjectRename(handler))
	cmdProject.AddCommand(project.NewCmdProjectRemove(handler))
	cmdRoot.AddCommand(cmdProject)

	cmdEnvironment := environment.NewCmdEnvironment()
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentList(handler))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentAdd(handler))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentRename(handler))
	cmdEnvironment.AddCommand(environment.NewCmdEnvironmentRemove(handler))
	cmdRoot.AddCommand(cmdEnvironment)

	cmdSecret := secret.NewCmdSecret()
	cmdSecret.AddCommand(secret.NewCmdSecretList(handler))
	cmdSecret.AddCommand(secret.NewCmdSecretSet(handler))
	cmdSecret.AddCommand(secret.NewCmdSecretGet(handler))
	cmdSecret.AddCommand(secret.NewCmdSecretRemove(handler))
	cmdRoot.AddCommand(cmdSecret)

	cmdUser := user.NewCmdUser()
	cmdUser.AddCommand(user.NewCmdUserRegister(handler))
	cmdRoot.AddCommand(cmdUser)

	cmdInject := inject.NewCmdInject(injectCLIHandler)
	cmdRoot.AddCommand(cmdInject)

	helpers.WalkCmd(cmdRoot, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
		c.Flags().BoolP("version", "v", false, "Print version information")
	})

	if err := cmdRoot.Execute(); err != nil {
		os.Exit(1)
	}
}

func injectCLIHandler(cmd *cobra.Command, args []string) error {
	w := bytes.NewBufferString("")
	injectHandler := newCliHandler(w)

	if err := injectHandler(cmd, args); err != nil {
		return err
	}

	injection, err := io.ReadAll(w)
	if err != nil {
		return err
	}

	env := strings.Split(string(injection), " ")

	var command string
	var arguments []string

	if len(args) > 0 {
		command = args[0]
	}

	if len(args) > 1 {
		arguments = args[1:]
	}

	hostCmd := exec.Command(command, arguments...)
	hostCmd.Env = append(hostCmd.Environ(), env...)
	hostCmd.Stdout = cmd.OutOrStdout()

	if err := hostCmd.Run(); err != nil {
		cmd.SilenceUsage = true
		return err
	}

	return nil
}
