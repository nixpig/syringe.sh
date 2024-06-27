package main

import (
	"context"
	"fmt"
	"os"

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

	rootCmd.PersistentFlags().StringP("identity", "i", "", "Path to SSH key (if not provided, SSH agent is used)")

	projectCmd := project.New(nil)
	projectCmd.AddCommand(project.ListCmd(run))
	projectCmd.AddCommand(project.AddCmd(run))
	projectCmd.AddCommand(project.RenameCmd(run))
	projectCmd.AddCommand(project.RemoveCmd(run))
	rootCmd.AddCommand(projectCmd)

	environmentCmd := environment.New(nil)
	environmentCmd.AddCommand(environment.ListCmd(run))
	environmentCmd.AddCommand(environment.AddCmd(run))
	environmentCmd.AddCommand(environment.RenameCmd(run))
	environmentCmd.AddCommand(environment.RemoveCmd(run))
	rootCmd.AddCommand(environmentCmd)

	secretCmd := secret.New(nil)
	secretCmd.AddCommand(secret.ListCmd(run))
	secretCmd.AddCommand(secret.SetCmd(run))
	secretCmd.AddCommand(secret.GetCmd(run))
	secretCmd.AddCommand(secret.RemoveCmd(run))
	rootCmd.AddCommand(secretCmd)

	userCmd := user.New(nil)
	userCmd.AddCommand(user.RegisterCmd(run))
	rootCmd.AddCommand(userCmd)

	injectCmd := inject.New(nil)
	injectCmd.RunE = run
	rootCmd.AddCommand(injectCmd)

	helpers.WalkCmd(rootCmd, func(c *cobra.Command) {
		c.Flags().BoolP("help", "h", false, "Help for the "+c.Name()+" command")
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
