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
	rootCmd := root.New(context.Background())

	handler := newCliHandler(rootCmd.OutOrStdout())

	rootCmd.PersistentFlags().StringP("identity", "i", "", "Path to SSH key (if not provided, SSH agent is used)")

	projectCmd := project.NewCmdProject()
	projectCmd.AddCommand(project.NewCmdProjectList(handler))
	projectCmd.AddCommand(project.NewCmdProjectAdd(handler))
	projectCmd.AddCommand(project.NewCmdProjectRename(handler))
	projectCmd.AddCommand(project.NewCmdProjectRemove(handler))
	rootCmd.AddCommand(projectCmd)

	environmentCmd := environment.NewCmdEnvironment(nil)
	environmentCmd.AddCommand(environment.NewCmdEnvironmentList(handler))
	environmentCmd.AddCommand(environment.NewCmdEnvironmentAdd(handler))
	environmentCmd.AddCommand(environment.NewCmdEnvironmentRename(handler))
	environmentCmd.AddCommand(environment.NewCmdEnvironmentRemove(handler))
	rootCmd.AddCommand(environmentCmd)

	secretCmd := secret.NewCmdSecret(nil)
	secretCmd.AddCommand(secret.NewCmdSecretList(handler))
	secretCmd.AddCommand(secret.NewCmdSecretSet(handler))
	secretCmd.AddCommand(secret.NewCmdSecretGet(handler))
	secretCmd.AddCommand(secret.NewCmdSecretRemove(handler))
	rootCmd.AddCommand(secretCmd)

	userCmd := user.NewCmdUser(nil)
	userCmd.AddCommand(user.NewCmdUserRegister(handler))
	rootCmd.AddCommand(userCmd)

	injectCmd := inject.NewCmdInjectWithHandler(nil, injectCLIHandler)
	rootCmd.AddCommand(injectCmd)

	helpers.WalkCmd(rootCmd, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the '%s' command", c.Name()))
		c.Flags().BoolP("version", "v", false, "Print version information")
	})

	if err := rootCmd.Execute(); err != nil {
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
