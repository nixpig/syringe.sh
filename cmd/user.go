package cmd

import (
	"github.com/spf13/cobra"
)

func userCommand() *cobra.Command {
	userCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"u"},
		Short:   "Manage users",
	}

	return userCmd
}
