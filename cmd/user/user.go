package user

import (
	"fmt"

	"github.com/spf13/cobra"
)

func UserCommand() *cobra.Command {
	userCmd := &cobra.Command{
		Use:               "user",
		Aliases:           []string{"u"},
		Short:             "Manage users",
		PersistentPreRunE: initUserContext,
	}

	userCmd.AddCommand(userRegisterCommand())

	return userCmd
}

func userRegisterCommand() *cobra.Command {
	userRegisterCmd := &cobra.Command{
		Use:     "register [flags] [USERNAME]",
		Aliases: []string{"r"},
		Short:   "Register user",
		Example: "syringe user register nixpig",
		Args:    cobra.RangeArgs(0, 1),
		RunE:    userRegisterRunE,
	}

	return userRegisterCmd
}

func userRegisterRunE(cmd *cobra.Command, args []string) error {
	var username string

	if len(args) > 0 {
		username = args[0]
	}

	fmt.Println("register user...", username)

	return nil
}

func UserDeleteCommand() *cobra.Command {
	userDeleteCmd := &cobra.Command{
		Use:     "delete [flags] [USERNAME]",
		Aliases: []string{"d"},
		Short:   "Delete a user",
		Example: "syringe user delete nixpig",
		Args:    cobra.RangeArgs(0, 1),
		RunE:    userDeleteRunE,
	}

	return userDeleteCmd
}

func userDeleteRunE(cmd *cobra.Command, args []string) error {
	var username string

	if len(args) > 0 {
		username = args[0]
	}

	fmt.Println("delete user...", username)

	return nil
}

func initUserContext(cmd *cobra.Command, args []string) error {
	fmt.Println("initialise user context")

	return nil
}
