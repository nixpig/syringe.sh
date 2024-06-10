package cmd

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/spf13/cobra"
)

func NewRegisterCommand(
	sess ssh.Session,
	appService services.AppService,
) *cobra.Command {
	return &cobra.Command{
		Use:     "register",
		Aliases: []string{"r"},
		Short:   "Register new user",
		Long:    "Register a new user",
		Example: "syringe register",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := appService.RegisterUser(services.RegisterUserRequest{
				Username:  sess.User(),
				Email:     "not_used_yet@example.org",
				PublicKey: sess.PublicKey(),
			})
			if err != nil {
				return
			}
		},
	}
}
