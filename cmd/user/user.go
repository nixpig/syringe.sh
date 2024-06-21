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
	userCmd.AddCommand(UserDeleteCommand())
	userCmd.AddCommand(UserAddKeyCommand())
	userCmd.AddCommand(UserRemoveKeyCommand())
	userCmd.AddCommand(UserListKeysCommand())

	return userCmd
}

func userRegisterCommand() *cobra.Command {
	userRegisterCmd := &cobra.Command{
		Use:     "register [flags] [USERNAME]",
		Aliases: []string{"r"},
		Short:   "Register user",
		Example: "syringe user register -i ~/.ssh/id_rsa",
		Args:    cobra.NoArgs,
		RunE:    userRegisterRunE,
	}

	userRegisterCmd.Flags().StringP("identity", "i", "", "Path to SSH key")

	return userRegisterCmd
}

func userRegisterRunE(cmd *cobra.Command, args []string) error {
	identity, err := cmd.Flags().GetString("identity")
	if err != nil {
		return err
	}

	fmt.Println("register user...", identity)

	return nil
}

func UserDeleteCommand() *cobra.Command {
	userDeleteCmd := &cobra.Command{
		Use:     "delete [flags] [USERNAME]",
		Aliases: []string{"d"},
		Short:   "Delete a user",
		Example: "syringe user delete -i ~/.ssh/id_rsa",
		Args:    cobra.NoArgs,
		RunE:    userDeleteRunE,
	}

	userDeleteCmd.Flags().StringP("identity", "i", "", "Path to SSH key")

	return userDeleteCmd
}

func userDeleteRunE(cmd *cobra.Command, args []string) error {
	identity, err := cmd.Flags().GetString("identity")
	if err != nil {
		return err
	}

	fmt.Println("delete user...", identity)

	return nil
}

func UserListKeysCommand() *cobra.Command {
	userListKeysCmd := &cobra.Command{
		Use:     "list-keys [flags]",
		Aliases: []string{"kl"},
		Short:   "List keys",
		Example: "syringe user list-keys",
		Args:    cobra.NoArgs,
		RunE:    userListKeysRunE,
	}

	return userListKeysCmd
}

func userListKeysRunE(cmd *cobra.Command, args []string) error {
	return nil
}

func UserAddKeyCommand() *cobra.Command {
	userAddKeyCmd := &cobra.Command{
		Use:     "add-key [flags]",
		Aliases: []string{"ka"},
		Short:   "Add a SSH key",
		Example: "syringe user add-key -i ~/.ssh/id_rsa",
		Args:    cobra.NoArgs,
		RunE:    userAddKeyRunE,
	}

	userAddKeyCmd.Flags().StringP("identity", "i", "", "Path to SSH key")

	return userAddKeyCmd
}

func userAddKeyRunE(cmd *cobra.Command, args []string) error {
	identity, err := cmd.Flags().GetString("identity")
	if err != nil {
		return err
	}

	fmt.Println("add key...", identity)

	return nil
}

func UserRemoveKeyCommand() *cobra.Command {
	userRemoveKeyCmd := &cobra.Command{
		Use:     "remove-key [flags]",
		Aliases: []string{"kr"},
		Short:   "Remove a SSH key",
		Example: "syringe user remove-key -i ~/.ssh/id_rsa",
		Args:    cobra.NoArgs,
		RunE:    userRemoveKeyRunE,
	}

	return userRemoveKeyCmd
}

func userRemoveKeyRunE(cmd *cobra.Command, args []string) error {
	identity, err := cmd.Flags().GetString("identity")
	if err != nil {
		return err
	}

	fmt.Println("remove key...", identity)

	return nil
}

func initUserContext(cmd *cobra.Command, args []string) error {
	fmt.Println("initialise user context")

	return nil
}
