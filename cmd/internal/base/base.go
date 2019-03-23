package base

import (
	"errors"
	"io"

	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/spf13/cobra"
)

// CliCommand is an interface for base commands
type CliCommand interface {
	Cmd() *cobra.Command
	Addcmd(...CliCommand) CliCommand
	Run(*cobra.Command, []string) error
	SetOutput(io.Writer)
}

// Command is a cli command
type Command struct {
	cmd    *cobra.Command
	Addr   *obj.Address
	Menu   *dawg.Menu
	output io.Writer
}

// Cmd returns the internal cobra.Command
func (c *Command) Cmd() *cobra.Command {
	return c.cmd
}

// Addcmd adds a command to the command tree
func (c *Command) Addcmd(cmds ...CliCommand) CliCommand {
	for _, cmd := range cmds {
		c.cmd.AddCommand(cmd.Cmd())
	}
	return c
}

// Run runs the command.
func (c *Command) Run(cmd *cobra.Command, args []string) error {
	return errors.New("not implimented")
}

// SetOutput sets the command output
func (c *Command) SetOutput(out io.Writer) {
	c.output = out
	c.cmd.SetOutput(c.output)
}
