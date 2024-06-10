package cmd

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/spf13/cobra"
)

func Execute(
	sess ssh.Session,
	handlers handlers.SshHandlers,
) error {

	rootCmd := &cobra.Command{
		Use:   "syringe",
		Short: "A terminal-based utility to securely manage environment variables across projects and environments.",
		Long:  "A terminal-based utility to securely manage environment variables across projects and environments.",
	}

	rootCmd.AddCommand(NewRegisterCommand(sess, handlers))
	rootCmd.AddCommand(NewSecretCommand(sess))

	rootCmd.SetArgs(sess.Command())
	rootCmd.SetIn(sess)
	rootCmd.SetOut(sess)
	rootCmd.SetErr(sess.Stderr())
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
