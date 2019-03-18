package cmd

import (
	"github.com/spf13/cobra"
)

type cartCmd struct {
	*basecmd
	price  bool
	delete bool
	add    []string
}

func (c *cartCmd) run(cmd *cobra.Command, args []string) error {
	return nil
}

func newCartCmd() cliCommand {
	c := &cartCmd{}
	c.basecmd = newBaseCommand("cart <order name>", "Manage user created orders", c.run)
	c.basecmd.cmd.Long = `The cart command gets information on all of the user
created orders. Use 'apizza cart <order name>' for info on a specific order`

	c.cmd.Flags().BoolVarP(&c.price, "price", "p", c.price, "show to price of an order")
	c.cmd.Flags().StringSliceVarP(&c.add, "add", "a", c.add, "add any number of products to a specific order")
	c.cmd.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "delete the order from the database")
	return c
}
