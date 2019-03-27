package base

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CliCommand is an interface for base commands
type CliCommand interface {
	Runner
	Cmd() *cobra.Command
	Addcmd(...CliCommand) CliCommand
	SetOutput(io.Writer)
	Output() io.Writer
	Printf(string, ...interface{})
	Println(...interface{})
	Flags() *pflag.FlagSet
}

// Runner defines an interface for an object that can be run.
type Runner interface {
	Run(*cobra.Command, []string) error
}

// NewCommand returns a new base command.
func NewCommand(use, short string, f runFunction) *Command {
	return &Command{
		cmd: &cobra.Command{
			Use:           use,
			Short:         short,
			RunE:          f,
			SilenceErrors: true,
			SilenceUsage:  true,
		},
		output: os.Stdout,
	}
}

// Command is a cli command
type Command struct {
	cmd    *cobra.Command
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

// Output returns the command's output writer.
func (c *Command) Output() io.Writer {
	return c.output
}

// Flags returns the flag set.
func (c *Command) Flags() *pflag.FlagSet {
	return c.cmd.Flags()
}

// Printf prints to the command's output.
func (c *Command) Printf(format string, a ...interface{}) {
	fmt.Fprintf(c.output, format, a...)
}

// Println prints to the command's output.
func (c *Command) Println(a ...interface{}) {
	fmt.Fprintln(c.output, a...)
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

type runFunction func(*cobra.Command, []string) error
