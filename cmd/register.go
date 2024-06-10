package cmd

import (
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/spf13/cobra"
)

func NewRegisterCommand(
	sess ssh.Session,
	appHandlers handlers.SshHandlers,
) *cobra.Command {
	return &cobra.Command{
		Use:     "register",
		Aliases: []string{"r"},
		Short:   "Register new user",
		Long:    "Register a new user",
		Example: "syringe register",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := appHandlers.RegisterUser(sess.User(), sess.PublicKey()); err != nil {
				fmt.Println("FUCKKK!!!", err)
			}
		},
	}
}
