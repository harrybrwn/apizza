package base

import (
	"bytes"
	"io"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/obj"
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
	return c.cmd.Usage()
}

// SetOutput sets the command output
func (c *Command) SetOutput(out io.Writer) {
	c.output = out
	c.cmd.SetOutput(c.output)
}

// WithCmds returns a general test given a more specific test function.
//
// This wrapper function is meant for testing only.
func WithCmds(
	test func(*testing.T, *bytes.Buffer, ...CliCommand),
	cmds ...CliCommand,
) func(*testing.T) {
	return func(t *testing.T) {
		buf := &bytes.Buffer{}

		for i := range cmds {
			cmds[i].SetOutput(buf)
		}

		test(t, buf, cmds...)
	}
}
