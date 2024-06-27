package helpers

import "github.com/spf13/cobra"

func CmdWalker(c *cobra.Command, f func(*cobra.Command)) {
	f(c)
	for _, c := range c.Commands() {
		CmdWalker(c, f)
	}
}
