package cmd

import (
	"apizza/pkg/cache"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestRunner(t *testing.T) {
	var funcname = func(a interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(a).Pointer()).Name()
	}

	var tests = []func(*testing.T){
		initDummyCheck,
		testOrderNew,
		testFindProduct,
	}

	for _, f := range tests {
		t.Run(funcname(f), f)
	}
	teardownTests()
}

func testFindProduct(t *testing.T) {
	c := newBuilder().newMenuCmd().(*menuCmd)
	if err := c.initMenu(); err != nil {
		t.Error(err)
	}
	buf := &bytes.Buffer{}
	c.output = buf
	c.all = true

	c.printMenu()
	if len(buf.Bytes()) < 10000 {
		t.Error("the menu seems to be a bit short in length")
	}
	buf.Reset()
	c.printToppings()
	if len(buf.Bytes()) < 1000 {
		t.Error("toppings menu seems too short")
	}
}

func TestStringStuff(t *testing.T) {
	if strLen("123456") != 6 {
		t.Error("wrong string len")
	}
	strs := []interface{}{}
	for i := 0; i < 10; i++ {
		strs = append(strs, spaces(i))
		if strLen(strs[i].(string)) != i {
			t.Error("wrong string len")
		}
	}
	if maxStrLen(strs) != 9 {
		t.Error("wrong max length")
	}
}

func initDummyCheck(t *testing.T) {
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

// omg, i can't beleve i haven't been putting this in my other tests, this is great
func init() {
	var check = func(e error, msg string) {
		if e != nil {
			fmt.Printf("test setup failed: %s - %s\n", e, msg)
			os.Exit(1)
		}
	}

	wd, err := os.Getwd()
	check(err, "working dir")
	dir := filepath.Join(wd, "testdata")

	db, err = cache.GetDB(filepath.Join(dir, "test.db"))
	check(err, "database")
	err = db.Put("test", []byte("this is some test data"))
	check(err, "database put")

	raw := `{"Name":"joe","Email":"nojoe@mail.com","Address":{"Street":"1600 Pennsylvania Ave NW","City":"Washington DC","State":"","Zip":"20500"},"Card":{"Number":"","Expiration":"","CVV":""},"Service":"Carryout","MyOrders":null}`
	err = json.Unmarshal([]byte(raw), cfg)
	check(err, "json")
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
