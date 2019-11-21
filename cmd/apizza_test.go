package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/dawg"
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
		withDummyDB(withApizzaCmd(testApizzaResetflag, newApizzaCmd())),
		testMenuRun, testConfigStruct, testConfigCmd,
		testConfigGet, testConfigSet, withDummyDB(testExec),
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
	a := newapp(makedummydb(), new(Config), nil)
	defer a.db.Destroy()
	a.SetOutput(new(bytes.Buffer))

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
	tests.Compare(
		t, a.Output().(*bytes.Buffer).String(),
		fmt.Sprintf("removing %s\n", a.DB().Path()))
}

func TestAppStoreFinder(t *testing.T) {
	buf := new(bytes.Buffer)
	cf := &Config{
		Service: dawg.Delivery,
		Address: *cmdtest.TestAddress(),
	}
	a := newapp(makedummydb(), cf, buf)
	defer a.db.Destroy()

	store := a.store()
	if store == nil {
		// t.Error("what")
	}
}

func testApizzaResetflag(t *testing.T, buf *bytes.Buffer, c *apizzaCmd) {
	c.Cmd().ParseFlags([]string{"--clear-cache"})
	c.clearCache = true
	test = false
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(db.Path()); os.IsExist(err) {
		t.Error("database should not exitst")
	} else if !os.IsNotExist(err) {
		t.Error("database should not exitst")
	}
	tests.Compare(t, buf.String(), fmt.Sprintf("removing %s\n", db.Path()))
}

// testExec must be run last.
func testExec(t *testing.T) {
	b := newBuilder()
	buf := &bytes.Buffer{}
	b.root.Cmd().SetOutput(buf)
	if err := b.exec(); err != nil {
		t.Error(err)
	}
	// Execute()
	// tests.Compare(t, buf.String(), b.root.Cmd().UsageString())
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

func makedummydb() *cache.DataBase {
	newDatabase, err := cache.GetDB(tests.NamedTempFile("testdata", "apizza_dummy.db"))
	check(err, "dummy database")
	return newDatabase
}

func withDummyDB(fn func(*testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		newDatabase := makedummydb()
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
	raw := []byte(`
{
	"name":"joe","email":"nojoe@mail.com",
	"address":{
		"street":"1600 Pennsylvania Ave NW",
		"cityName":"Washington DC",
		"state":"","zipcode":"20500"
	},
	"card":{"number":"","expiration":"","cvv":""},
	"service":"Carryout"}`)
	check(json.Unmarshal(raw, cfg), "json")
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
