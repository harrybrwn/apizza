package cmd

import (
	"errors"
	"fmt"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"

	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/spf13/cobra"
)

type orderCmd struct {
	*basecmd
	verbose bool
	track   bool
}

func (c *orderCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if len(args) < 1 {
		return data.PrintOrders(db, c.Output(), c.verbose)
	} else if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}

	name := args[0]
	order, err := data.GetOrder(name, db)
	if err != nil {
		return err
	}

	err = dawg.ValidateOrder(order)
	if dawg.IsFailure(err) {
		return fmt.Errorf("invalid order:\n%s", err.Error())
	}
	if yesOrNo("Would you like to purchas this order? (y/n)") {
		c.Printf("ordering '%s'...\n", order.Name())
	}
	return nil
}

func newOrderCmd() base.CliCommand {
	c := &orderCmd{verbose: false}
	c.basecmd = newCommand("order", "Send an order from the cart to dominos.", c)

	c.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose, "output the order command verbosly")
	c.Flags().BoolVarP(&c.track, "track", "t", c.track, "enable tracking for the purchased order")
	return c
}
