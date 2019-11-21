package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestRunner(t *testing.T) {
	r := tests.NewRunner(t, setupTests, teardownTests)
	app := newapp(db, cfg, nil)
	r.AddTest(
		dummyCheck,
		base.WithCmds(testOrderNew, newCartCmd(app.builder), newAddOrderCmd(app.builder)),
		base.WithCmds(testAddOrder, newCartCmd(app.builder), newAddOrderCmd(app.builder)),
		base.WithCmds(testOrderNewErr, newAddOrderCmd(app.builder)),
		base.WithCmds(testOrderRunAdd, newCartCmd(app.builder)),
		withCartCmd(app.builder, testOrderPriceOutput),
		withCartCmd(app.builder, testAddToppings),
		withCartCmd(app.builder, testOrderRunDelete),
		testFindProduct,
		withAppCmd(testAppRootCmdRun, app),
		testMenuRun,
		testConfigCmd,
		testConfigGet, testConfigSet,
	)
	r.Run()
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

func dummyCheck(t *testing.T) {
	data, err := db.Get("test")
	if err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(data), "this is some test data")
	if cfg.Name != "joe" || cfg.Email != "nojoe@mail.com" {
		t.Error("test data is not correct")
	}
	if err = db.Delete("test"); err != nil {
		t.Error("couldn't delete test", err)
	}
}

func withDummyDB(fn func(*testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		newDatabase := cmdtest.TempDB()
		oldDatabase := db
		db = newDatabase
		defer func() {
			db = oldDatabase
			newDatabase.Destroy()
		}()
		fn(t)
	}
}

func setupTests() {
	var err error
	db, err = cache.GetDB(tests.NamedTempFile("testdata", "apizza_test.db"))
	check(err, "database")
	err = db.Put("test", []byte("this is some test data"))
	check(err, "database put")
	config.SetNonFileConfig(cfg) // don't want it to over ride the file on disk
	check(json.Unmarshal([]byte(testconfigjson), cfg), "json")
}

func teardownTests() {
	if err := db.Destroy(); err != nil {
		panic(err)
	}
}

func withApizzaCmd(f func(*testing.T, *bytes.Buffer, *apizzaCmd), c base.CliCommand) func(*testing.T) {
	return func(t *testing.T) {
		cmd, ok := c.(*apizzaCmd)
		if !ok {
			t.Error("not an *apizzaCmd")
		}
		buf := &bytes.Buffer{}
		cmd.SetOutput(buf)
		f(t, buf, cmd)
	}
}

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
	b *cliBuilder,
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
