package cmd

import (
	"apizza/pkg/config"
	"fmt"

	"github.com/spf13/cobra"
)

type apizzaCmd struct {
	basecmd
	address string
	service string
	storeID string
	test    bool
}

func (c *apizzaCmd) run(cmd *cobra.Command, args []string) (err error) {
	if c.test {
		fmt.Println("config file:", config.File())
	} else {
		err = cmd.Usage()
	}
	return err
}

func newApizzaCmd() cliCommand {
	c := &apizzaCmd{address: "", service: cfg.Service}

	c.basecmd.cmd = &cobra.Command{
		Use:   "apizza",
		Short: "Dominos pizza from the command line.",
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return config.Save()
		},
	}
	c.cmd.RunE = c.run

	c.cmd.PersistentFlags().StringVar(&c.address, "address", c.address, "use a specific address")
	c.cmd.PersistentFlags().StringVar(&c.service, "service", c.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	c.cmd.Flags().BoolVarP(&c.test, "test", "t", false, "testing flag")
	c.cmd.Flags().MarkHidden("test")
	return c
}
