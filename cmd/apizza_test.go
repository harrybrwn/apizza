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
	var funcname = func(a interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	}

	setupTests()
	defer teardownTests()

	b := newBuilder()

	var tests = []func(*testing.T){
		dummyCheckForinit,
		withTwoCmd(b.newCartCmd(), b.newAddOrderCmd(), testOrderNew),
		withTwoCmd(b.newCartCmd(), b.newAddOrderCmd(), testAddOrder),
		withCmd(b.newAddOrderCmd(), testOrderNewErr),
		withCmd(b.newCartCmd(), testOrderRunAdd),
		withCartCmd(b, testOrderPriceOutput),
		withCartCmd(b, testOrderRunDelete),
		testFindProduct,
		withApizza(newApizzaCmd(), testApizzaCmdRun),
		withDummyDB(withApizza(newApizzaCmd(), testApizzaResetflag)),
		testMenuRun,
		testExec,
		testConfigStruct,
		testConfigCmd,
		testConfigGet,
		testConfigSet,
	}

	for _, f := range tests {
		t.Run(funcname(f), f)
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
	newDatabase, err := cache.GetDB(tests.NamedTempFile("testdata", "testing_dummyDB.db"))
	check(err, "dummy database")
	err = newDatabase.Put("test", []byte("this is a testvalue"))
	check(err, "db.Put")

	oldDatabase := db
	return func(t *testing.T) {
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

func withCmd(
	c cliCommand,
	f func(cliCommand, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	return func(t *testing.T) {
		buf := &bytes.Buffer{}
		c.setOutput(buf)
		f(c, buf, t)
	}
}

func withTwoCmd(
	c1, c2 cliCommand,
	f func(cliCommand, cliCommand, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	buf := &bytes.Buffer{}
	c1.setOutput(buf)
	c2.setOutput(buf)
	return func(t *testing.T) { f(c1, c2, buf, t) }
}

func withApizza(c cliCommand, f func(cliCommand, *testing.T)) func(*testing.T) {
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
