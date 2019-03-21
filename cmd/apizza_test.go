package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestRunner(t *testing.T) {
	b := newBuilder()

	var apizzaTests = []func(*testing.T){
		withCmds(testOrderNew, b.newCartCmd(), b.newAddOrderCmd()),
		withCmds(testAddOrder, b.newCartCmd(), b.newAddOrderCmd()),
		withCmds(testOrderNewErr, b.newAddOrderCmd()),
		withCmds(testOrderRunAdd, b.newCartCmd()),
		withCartCmd(b, testOrderPriceOutput),
		withCartCmd(b, testOrderRunDelete),
		testFindProduct,
		withApizzaCmd(newApizzaCmd(), testApizzaCmdRun),
		withDummyDB(withApizzaCmd(newApizzaCmd(), testApizzaResetflag)),
		testMenuRun,
		testExec,
		testConfigStruct,
		testConfigCmd,
		testConfigGet,
		testConfigSet,
	}
	runtests(t, apizzaTests)
}

func runtests(t *testing.T, pTests []func(*testing.T)) {
	var funcName = func(a interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	}

	allTests := append([]func(*testing.T){dummyCheckForinit}, pTests...)

	setupTests()
	defer teardownTests()

	for _, f := range allTests {
		t.Run(funcName(f), f)
	}
}

func testApizzaCmdRun(c cliCommand, t *testing.T) {
	c.command().ParseFlags([]string{})
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}

	c.command().ParseFlags([]string{"--test"})
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
}

func testApizzaResetflag(c cliCommand, t *testing.T) {
	c.command().ParseFlags([]string{"--clear-cache"})
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
}

func testExec(t *testing.T) {
	b := newBuilder()
	b.root.command().SetOutput(&bytes.Buffer{})

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
			os.Remove(newDatabase.Path) // ignoring errors because it may already be gone
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

	raw := []byte(`
{
	"Name":"joe",
	"Email":"nojoe@mail.com",
	"Address":{
		"Street":"1600 Pennsylvania Ave NW",
		"CityName":"Washington DC",
		"State":"",
		"Zipcode":"20500"
	},
	"Card":{
		"Number":"",
		"Expiration":"",
		"CVV":""
	},
	"Service":"Carryout",
	"MyOrders":null
}`)
	check(json.Unmarshal(raw, cfg), "json")
}

func teardownTests() {
	if err := db.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(db.Path); err != nil {
		panic(err)
	}
}

func withCmds(test func(*testing.T, *bytes.Buffer, ...cliCommand), cmds ...cliCommand) func(*testing.T) {
	return func(t *testing.T) {
		buf := &bytes.Buffer{}
		for i := range cmds {
			cmds[i].setOutput(buf)
		}
		test(t, buf, cmds...)
	}
}

func withApizzaCmd(c cliCommand, f func(cliCommand, *testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		c.setOutput(&bytes.Buffer{})
		f(c, t)
	}
}

func withCartCmd(
	b *cliBuilder,
	f func(*cartCmd, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	return func(t *testing.T) {
		cart := b.newCartCmd().(*cartCmd)
		buf := &bytes.Buffer{}
		cart.setOutput(buf)

		f(cart, buf, t)
	}
}

func check(e error, msg string) {
	if e != nil {
		fmt.Printf("test setup failed: %s - %s\n", e, msg)
		os.Exit(1)
	}
}
