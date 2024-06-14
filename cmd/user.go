package cmd

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/spf13/cobra"
)

func userCommand(
	sess ssh.Session,
	appService services.AppService,
) *cobra.Command {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "Manage users",
	}

	userCmd.AddCommand(userRegisterCommand(sess, appService))

	return userCmd
}

func userRegisterCommand(sess ssh.Session, appService services.AppService) *cobra.Command {
	userRegisterCmd := &cobra.Command{
		Use:     "register [flags]",
		Aliases: []string{"r"},
		Short:   "Register new user",
		Example: "syringe register",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := appService.RegisterUser(services.RegisterUserRequest{
				Username:  sess.User(),
				Email:     "not_used_yet@example.org",
				PublicKey: sess.PublicKey(),
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	return userRegisterCmd
}
