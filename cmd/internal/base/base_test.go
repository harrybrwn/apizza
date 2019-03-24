package base

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

type testCommand struct {
	*Command
}

var _ CliCommand = (*testCommand)(nil)

func (c *testCommand) Run(cmd *cobra.Command, args []string) error {
	c.Println("test output")
	return nil
}

func TestCommand(t *testing.T) {
	c := &testCommand{}
	c.Command = NewCommand("test", "test for base.Command", c.Run)

	t.Run("inner_test", WithCmds(testCmd, c))
}

func testCmd(t *testing.T, buf *bytes.Buffer, cmds ...CliCommand) {
	c := cmds[0]
	if c == nil {
		t.Error("nil command")
	}

	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "test output" {
		t.Error("wrong output")
	}
}
