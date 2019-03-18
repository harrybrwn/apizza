package cmd

import (
	"github.com/spf13/cobra"
)

type cartCmd struct {
	*basecmd
}

func (c *cartCmd) run(cmd *cobra.Command, args []string) error {
	return nil
}

func newCartCmd() cliCommand {
	c := &cartCmd{}
	c.basecmd = newBaseCommand("cart <order name>", "Manage user created orders", c.run)

	return c
}
