package user

import (
	"github.com/spf13/cobra"
)

func UserCommand() *cobra.Command {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "Manage users",
	}

	return userCmd
}
