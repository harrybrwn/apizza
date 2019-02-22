package cmd

import (
	"apizza/pkg/config"
	"fmt"

	"github.com/spf13/cobra"
)

// TODO: re-write Execute func with the new cli builder system... ex.
// func Execute() {
// 	builder := cliBuilder{}
// }

type cliBuilder struct {
	root     cliCommand
	commands []cliCommand
}

type builder interface {
	build()
	add(cliCommand)
}

func (b *cliBuilder) build() {
	b.addAll()
	b.root.addCommands(b.commands...)
}

func (b *cliBuilder) add(cmd cliCommand) {
	b.commands = append(b.commands, cmd)
}

func (b *cliBuilder) addAll() {
	// menually add all commands here with new<cmd name> functions
	b.commands = []cliCommand{}
}

func (b *cliBuilder) newApizzaCmd() cliCommand {
	return newApizzaCmd(cobra.Command{
		Use:   "apizza",
		Short: "Dominos pizza from the command line.",
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return config.Save()
		},
	})
}

type cliCommand interface {
	command() *cobra.Command
	addCommands(...cliCommand)
	run(*cobra.Command, []string) error
}

type apizzaCmd struct {
	cmd     *cobra.Command
	address string
	service string
	storeID string
	test    bool
}

func (c *apizzaCmd) command() *cobra.Command {
	return c.cmd
}

func (c *apizzaCmd) addCommands(cmds ...cliCommand) {
	for _, cmd := range cmds {
		c.cmd.AddCommand(cmd.command())
	}
}

func (c *apizzaCmd) run(cmd *cobra.Command, args []string) (err error) {
	if c.test {
		fmt.Println("config file:", config.File())
	} else {
		err = cmd.Usage()
	}
	return err
}

func newApizzaCmd(cmd cobra.Command) *apizzaCmd {
	c := &apizzaCmd{cmd: &cmd, address: ""}
	c.cmd.RunE = c.run

	c.cmd.PersistentFlags().StringVar(&c.address, "address", c.address, "use a specific address")
	c.cmd.PersistentFlags().StringVar(&c.service, "service", cfg.Service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	c.cmd.Flags().BoolVarP(&c.test, "test", "t", false, "testing flag")
	c.cmd.Flags().MarkHidden("test")
	return c
}
