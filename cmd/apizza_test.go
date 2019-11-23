package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/config"
)

func TestMain(m *testing.M) {
	config.SetNonFileConfig(cfg) // don't want it to over ride the file on disk
	check(json.Unmarshal([]byte(testconfigjson), cfg), "json")
	m.Run()
}

func TestRunner(t *testing.T) {
	// builder := newapp(cmdtest.TempDB(), cfg, nil)
	app := newapp(cmdtest.TempDB(), cfg, nil)
	builder := cmdtest.NewRecorder()

	tsts := []func(*testing.T){
		base.WithCmds(testOrderNew, newCartCmd(builder), newAddOrderCmd(builder)),
		base.WithCmds(testAddOrder, newCartCmd(builder), newAddOrderCmd(builder)),
		base.WithCmds(testOrderNewErr, newAddOrderCmd(builder)),
		base.WithCmds(testOrderRunAdd, newCartCmd(builder)),
		withCartCmd(builder, testOrderPriceOutput),
		withCartCmd(builder, testAddToppings),
		withCartCmd(builder, testOrderRunDelete),
		withAppCmd(testAppRootCmdRun, app),
	}
	for i, tst := range tsts {
		t.Run(fmt.Sprintf("Test %d", i), tst)
	}

	builder.CleanUp()
	if err := app.db.Close(); err != nil {
		t.Error(err)
	}
}

func testAppRootCmdRun(t *testing.T, buf *bytes.Buffer, a *App) {
	a.Cmd().ParseFlags([]string{})
	if err := a.Run(a.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if buf.String() != a.Cmd().UsageString() {
		t.Error("wrong output")
	}

	a.Cmd().ParseFlags([]string{"--test"})
	if err := a.Run(a.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	test = false
	buf.Reset()

	err := a.prerun(a.Cmd(), []string{})
	if err != nil {
		t.Error("should not return an error")
	}
	err = a.postrun(a.Cmd(), []string{})
	if err != nil {
		t.Error("should not return an error")
	}

	if len(a.Cmd().Commands()) != 0 {
		t.Error("should not have commands yet")
	}
	err = a.exec()
	if err != nil {
		t.Error(err)
	}
	if len(a.Cmd().Commands()) == 0 {
		t.Error("should have commands")
	}
}

func TestAppResetFlag(t *testing.T) {
	r := cmdtest.NewRecorder()
	a := newapp(r.ToApp())

	a.Cmd().ParseFlags([]string{"--clear-cache"})
	a.clearCache = true
	test = false
	if err := a.Run(a.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(a.DB().Path()); os.IsExist(err) {
		t.Error("database should not exitst")
	} else if !os.IsNotExist(err) {
		t.Error("database should not exitst")
	}
	r.Compare(t, fmt.Sprintf("removing %s\n", a.DB().Path()))
	r.ClearBuf()
}

func TestAppStoreFinder(t *testing.T) {
	r := cmdtest.NewRecorder()
	defer r.CleanUp()
	a := newapp(r.ToApp())

	store := a.store()
	if store == nil {
		t.Error("what")
	}
}

func setupTests() {
	config.SetNonFileConfig(cfg) // don't want it to over ride the file on disk
	check(json.Unmarshal([]byte(testconfigjson), cfg), "json")
}

func teardownTests() {}

func withAppCmd(f func(*testing.T, *bytes.Buffer, *App), c base.CliCommand) func(*testing.T) {
	return func(t *testing.T) {
		cmd, ok := c.(*App)
		if !ok {
			t.Error("not an *App")
		}
		buf := new(bytes.Buffer)
		cmd.SetOutput(buf)
		f(t, buf, cmd)
	}
}

func withCartCmd(
	b base.Builder,
	f func(*cartCmd, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	return func(t *testing.T) {
		cart := newCartCmd(b).(*cartCmd)
		buf := &bytes.Buffer{}
		cart.SetOutput(buf)

		f(cart, buf, t)
	}
}

func check(e error, msg string) {
	if e != nil {
		fmt.Printf("test setup failed: %s - %s\n", e, msg)
		os.Exit(1)
	}
}
