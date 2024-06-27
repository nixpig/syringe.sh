package helpers

import "github.com/spf13/cobra"

func WalkCmd(c *cobra.Command, f func(*cobra.Command)) {
	f(c)
	for _, c := range c.Commands() {
		WalkCmd(c, f)
	}
}
