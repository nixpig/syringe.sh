package user

import (
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewCmdUser() *cobra.Command {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "Manage users",
	}

	return userCmd
}

func NewCmdUserRegister(handler pkg.CobraHandler) *cobra.Command {
	registerCmd := &cobra.Command{
		Use:     "register [flags] [USERNAME]",
		Aliases: []string{"r"},
		Short:   "Register user",
		Example: "syringe user register",
		Args:    cobra.NoArgs,
		RunE:    handler,
	}

	return registerCmd
}
