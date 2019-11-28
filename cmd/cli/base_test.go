package base

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
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

func TestRunFunction(t *testing.T) {
	f := RunFunction(func(cmd *cobra.Command, args []string) error {
		return errs.New("this is a run func")
	})
	// rf = RunFunction(rf)
	err := f.Run(nil, nil)
	if err.Error() != "this is a run func" {
		t.Error("RunFunction gave wrong output")
	}
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

var testconfigjson = `
{
	"name":"joe","email":"nojoe@mail.com",
	"address":{
		"street":"1600 Pennsylvania Ave NW",
		"cityName":"Washington DC",
		"state":"","zipcode":"20500"
	},
	"card":{"number":"","expiration":"","cvv":""},
	"service":"Carryout"
}`

func TestConfigStruct(t *testing.T) {
	c := &Config{}
	config.SetNonFileConfig(c)
	err := json.Unmarshal([]byte(testconfigjson), c)
	if err != nil {
		t.Error(err)
	}

	if c.Get("name").(string) != "joe" {
		t.Error("wrong value")
	}
	if err := c.Set("name", "not joe"); err != nil {
		t.Error(err)
	}
	if c.Get("Name").(string) != "not joe" {
		t.Error("wrong value")
	}
	if err := c.Set("name", "joe"); err != nil {
		t.Error(err)
	}
	if err = c.Set("service", "should fail"); err == nil {
		t.Error("expected error")
	}
}
