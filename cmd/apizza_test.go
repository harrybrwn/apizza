package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestRunner(t *testing.T) {
	app := CreateApp(cmdtest.TempDB(), &cli.Config{}, nil)
	builder := cmdtest.NewRecorder()
	builder.ConfigSetup([]byte(cmdtest.TestConfigjson))

	tsts := []func(*testing.T){
		cli.WithCmds(testOrderNew, NewCartCmd(builder), newAddOrderCmd(builder)),
		cli.WithCmds(testAddOrder, NewCartCmd(builder), newAddOrderCmd(builder)),
		cli.WithCmds(testOrderNewErr, newAddOrderCmd(builder)),
		cli.WithCmds(testOrderRunAdd, NewCartCmd(builder)),
		withCartCmd(builder, testOrderPriceOutput),
		withCartCmd(builder, testAddToppings),
		withCartCmd(builder, testOrderRunDelete),
		withAppCmd(testAppRootCmdRun, app),
	}

	for i, tst := range tsts {
		t.Run(fmt.Sprintf("Test %d", i), tst)
	}

	builder.CleanUp()
	if err := app.db.Destroy(); err != nil {
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
	err = a.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestAppResetFlag(t *testing.T) {
	r := cmdtest.NewRecorder()
	a := CreateApp(r.ToApp())
	r.ConfigSetup([]byte(cmdtest.TestConfigjson))

	a.Cmd().ParseFlags([]string{"--clear-cache"})
	a.gOpts.ClearCache = true
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
	a := CreateApp(r.ToApp())

	store := a.Store()
	if store == nil {
		t.Error("what")
	}
}

func setupTests() {
	// config.SetNonFileConfig(cfg) // don't want it to over ride the file on disk
	// check(json.Unmarshal([]byte(testconfigjson), cfg), "json")
}

func teardownTests() {}

func withAppCmd(f func(*testing.T, *bytes.Buffer, *App), c cli.CliCommand) func(*testing.T) {
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
	b cli.Builder,
	f func(*cartCmd, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	return func(t *testing.T) {
		cart := NewCartCmd(b).(*cartCmd)
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

func TestExecute(t *testing.T) {
	var (
		exp string
		err error
		buf *bytes.Buffer
	)

	tt := []struct {
		args    []string
		exp     string
		outfunc func() string
		test    func(*testing.T)
		cleanup bool
	}{
		{
			args: []string{"config", "-f"}, test: nil, cleanup: false,
			outfunc: func() string {
				return fmt.Sprintf("setting up config file at %s\n%s\n", config.File(), config.File())
			},
		},
		{args: []string{"--delete-menu", "config", "-d"}, outfunc: func() string { return config.Folder() + "\n" }},
		{args: []string{"--service=Delivery", "config", "-d"}, outfunc: func() string { return config.Folder() + "\n" }},
		{args: []string{"--log=log.txt", "config", "-d"}, outfunc: func() string { return config.Folder() + "\n" }, test: func(t *testing.T) {
			if _, err = os.Stat(filepath.Join(config.Folder(), "logs", "log.txt")); os.IsNotExist(err) {
				// t.Error("file should exist")
			}
		}},
		{args: []string{"config", "-d"}, outfunc: func() string { return config.Folder() + "\n" }, cleanup: true},
	}

	var errmsg *ErrMsg
	for i, tc := range tt {
		buf, err = tests.CaptureOutput(func() {
			errmsg = Execute(tc.args, ".apizza/.tests")
		})
		if err != nil {
			t.Error(err)
		}
		if errmsg != nil {
			t.Error(errmsg.Msg, errmsg.Err)
		}

		if len(tc.exp) == 0 && tc.outfunc != nil {
			exp = tc.outfunc()
		} else {
			exp = tc.exp
		}

		if tc.test != nil {
			t.Run(fmt.Sprintf("Exec test: %d", i), tc.test)
		}

		tests.Compare(t, buf.String(), exp)

		config.Save()
		if tc.cleanup {
			err := os.RemoveAll(config.Folder())
			if err != nil {
				t.Error("could not remove test dir:", err)
			}
		}
	}
}
