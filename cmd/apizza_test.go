package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestRunner(t *testing.T) {
	r := tests.NewRunner(t, setupTests, teardownTests)
	b := newBuilder()

	r.AddTest(
		dummyCheckForinit,
		base.WithCmds(testOrderNew, b.newCartCmd(), b.newAddOrderCmd()),
		base.WithCmds(testAddOrder, b.newCartCmd(), b.newAddOrderCmd()),
		base.WithCmds(testOrderNewErr, b.newAddOrderCmd()),
		base.WithCmds(testOrderRunAdd, b.newCartCmd()),
		withCartCmd(b, testOrderPriceOutput),
		withCartCmd(b, testOrderRunDelete),
		testFindProduct,
		withApizzaCmd(testApizzaCmdRun, newApizzaCmd()),
		withDummyDB(withApizzaCmd(testApizzaResetflag, newApizzaCmd())),
		testMenuRun,
		testExec,
		testConfigStruct,
		testConfigCmd,
		testConfigGet,
		testConfigSet,
	)
	r.Run()
}

func testApizzaCmdRun(t *testing.T, buf *bytes.Buffer, c *apizzaCmd) {
	c.Cmd().ParseFlags([]string{})
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}

	c.Cmd().ParseFlags([]string{"--test"})
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	test = false
	buf.Reset()
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

	if string(buf.Bytes()) != fmt.Sprintf("removing %s\n", db.Path()) {
		t.Error("wrong output")
	}
}

func testExec(t *testing.T) {
	b := newBuilder()
	b.root.Cmd().SetOutput(&bytes.Buffer{})

	_, err := b.exec()
	if err != nil {
		t.Error(err)
	}
}

func dummyCheckForinit(t *testing.T) {
	data, err := db.Get("test")
	if err != nil {
		t.Error(err)
	}
	if string(data) != "this is some test data" {
		t.Error("database is broken. go check apizza/pkg/cache tests")
	}
	if cfg.Name != "joe" || cfg.Email != "nojoe@mail.com" {
		t.Error("test data is not correct")
	}
	if err = db.Delete("test"); err != nil {
		t.Error("couldn't delete test", err)
	}
}

func withDummyDB(fn func(*testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		newDatabase, err := cache.GetDB(tests.NamedTempFile("testdata", "testing_dummyDB.db"))
		check(err, "dummy database")
		err = newDatabase.Put("test", []byte("this is a testvalue"))
		check(err, "db.Put")

		oldDatabase := db
		db = newDatabase
		defer func() {
			db = oldDatabase
			check(newDatabase.Close(), "deleting dummy database")
			os.Remove(newDatabase.Path()) // ignoring errors because it may already be gone
		}()
		fn(t)
	}
}

func setupTests() {
	var err error
	db, err = cache.GetDB(tests.NamedTempFile("testdata", "test.db"))
	check(err, "database")
	err = db.Put("test", []byte("this is some test data"))
	check(err, "database put")

	config.SetNonFileConfig(cfg) // don't want it to over ride the file on disk
	raw := []byte(`
{
	"name":"joe",
	"email":"nojoe@mail.com",
	"address":{
		"street":"1600 Pennsylvania Ave NW",
		"cityName":"Washington DC",
		"state":"",
		"zipcode":"20500"
	},
	"card":{
		"number":"",
		"expiration":"",
		"cvv":""
	},
	"service":"Carryout"
}`)
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

func withCartCmd(
	b *cliBuilder,
	f func(*cartCmd, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	return func(t *testing.T) {
		cart := b.newCartCmd().(*cartCmd)
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
