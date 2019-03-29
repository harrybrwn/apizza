package base

import (
	"bytes"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
	"github.com/spf13/cobra"
)

type testCommand struct {
	*Command
}

var _ CliCommand = (*testCommand)(nil)

func (c *testCommand) Run(cmd *cobra.Command, args []string) error {
	c.Printf("command")
	c.Println("test output")
	return nil
}

func TestNewCommand(t *testing.T) {
	c := &testCommand{}
	c.Command = NewCommand("test", "test for base.Command", c.Run)
	c.SetOutput(os.Stdout)
	if c.Output() != os.Stdout {
		t.Error("wrong output")
	}
	t.Run("inner_test", WithCmds(testCmd, c))
	if c.Cmd().Use != "test" {
		t.Error("bad cobra.Command")
	}
	if c.Cmd().Short != "test for base.Command" {
		t.Error("bad cobra.Command")
	}
	c.Flags().String("testflag", "", "test flags")
	c.Addcmd(&testCommand{Command: NewCommand("subcmd1", "descr.", func(*cobra.Command, []string) error { return nil })})
	c.AddCobraCmd(&cobra.Command{Use: "subcmd2", Short: "descr."})
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.CompareOutput(t, "", func() {
		if err := c.Command.Run(c.Cmd(), []string{}); err != nil {
			t.Error(err)
		}
	})
}

func testCmd(t *testing.T, buf *bytes.Buffer, cmds ...CliCommand) {
	c := cmds[0]
	if c == nil {
		t.Error("nil command")
	}
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, buf.String(), "commandtest output\n")
}
