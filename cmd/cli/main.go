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

	projectCmd := project.New(nil)
	projectCmd.AddCommand(project.ListCmd(handler))
	projectCmd.AddCommand(project.AddCmd(handler))
	projectCmd.AddCommand(project.RenameCmd(handler))
	projectCmd.AddCommand(project.RemoveCmd(handler))
	rootCmd.AddCommand(projectCmd)

	environmentCmd := environment.New(nil)
	environmentCmd.AddCommand(environment.ListCmd(handler))
	environmentCmd.AddCommand(environment.AddCmd(handler))
	environmentCmd.AddCommand(environment.RenameCmd(handler))
	environmentCmd.AddCommand(environment.RemoveCmd(handler))
	rootCmd.AddCommand(environmentCmd)

	secretCmd := secret.New(nil)
	secretCmd.AddCommand(secret.ListCmd(handler))
	secretCmd.AddCommand(secret.SetCmd(handler))
	secretCmd.AddCommand(secret.GetCmd(handler))
	secretCmd.AddCommand(secret.RemoveCmd(handler))
	rootCmd.AddCommand(secretCmd)

	userCmd := user.New(nil)
	userCmd.AddCommand(user.RegisterCmd(handler))
	rootCmd.AddCommand(userCmd)

	injectCmd := inject.NewWithHandler(nil, func(cmd *cobra.Command, args []string) error {
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

		hostCmd := exec.Command(args[0], args[1:]...)
		hostCmd.Env = append(hostCmd.Environ(), env...)
		hostCmd.Stdout = cmd.OutOrStdout()

		if err := hostCmd.Run(); err != nil {
			cmd.SilenceUsage = true
			return err
		}

		return nil
	})

	injectCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{
		UnknownFlags: true,
	}

	rootCmd.AddCommand(injectCmd)

	helpers.WalkCmd(rootCmd, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for the %s command", c.Name()))
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error in here...", err)
		os.Exit(1)
	}
}
