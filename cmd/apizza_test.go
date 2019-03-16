package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/harrybrwn/apizza/pkg/cache"
)

func TestRunner(t *testing.T) {
	var funcname = func(a interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	}

	setupTests()
	defer teardownTests()

	var tests = []func(*testing.T){
		dummyCheckForinit,
		testOrderNew,
		testFindProduct,
		testApizzaCmdRun,
		withDummyDB(testApizzaResetflag),
		testMenuRun,
	}

	for _, f := range tests {
		t.Run(funcname(f), f)
	}
}

func testApizzaCmdRun(t *testing.T) {
	c := newApizzaCmd().(*apizzaCmd)
	buf := &bytes.Buffer{}
	c.output = buf
	c.command().SetOutput(buf)

	c.command().ParseFlags([]string{})
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}

	c.command().ParseFlags([]string{"--test"})
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
}

func testApizzaResetflag(t *testing.T) {
	c := newApizzaCmd().(*apizzaCmd)
	buf := &bytes.Buffer{}
	c.output = buf
	c.command().SetOutput(buf)
	c.command().ParseFlags([]string{"--clear-cache"})
	if err := c.run(c.command(), []string{}); err != nil {
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
}

func withDummyDB(fn func(*testing.T)) func(*testing.T) {
	wd, err := os.Getwd()
	check(err, "working dir")

	dbPath := filepath.Join(wd, "testdata", "testing_dummyDB.db")
	newDatabase, err := cache.GetDB(dbPath)
	check(err, "dummy database")
	err = newDatabase.Put("test", []byte("this is a testvalue"))

	oldDatabase := db
	return func(t *testing.T) {
		db = newDatabase
		defer func() {
			db = oldDatabase
			check(newDatabase.Close(), "deleting dummy database")
			os.Remove(dbPath)
		}()
		fn(t)
	}
}

func setupTests() {
	wd, err := os.Getwd()
	check(err, "working dir")

	db, err = cache.GetDB(filepath.Join(wd, "testdata", "test.db"))
	check(err, "database")
	err = db.Put("test", []byte("this is some test data"))
	check(err, "database put")

	raw := []byte(`
{
	"Name":"joe",
	"Email":"nojoe@mail.com",
	"Address":{
		"Street":"1600 Pennsylvania Ave NW",
		"City":"Washington DC",
		"State":"",
		"Zip":"20500"
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

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err = os.Remove(filepath.Join(wd, "testdata", "test.db")); err != nil {
		panic(err)
	}
	if err = os.Remove(filepath.Join(wd, "testdata")); err != nil {
		panic(err)
	}
}

func check(e error, msg string) {
	if e != nil {
		fmt.Printf("test setup failed: %s - %s\n", e, msg)
		os.Exit(1)
	}
}
