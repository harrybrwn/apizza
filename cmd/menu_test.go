package cmd

import (
	"apizza/pkg/cache"
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

	defer func() {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}()

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

	var tests = []func(*testing.T){
		testFindProduct,
	}

	for _, f := range tests {
		t.Run(funcname(f), f)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Error("error in cleanup:", err)
	}
	err = os.RemoveAll(filepath.Join(wd, "testdata"))
	if err != nil {
		t.Error("error in cleanup:", err)
	}
}

func testFindProduct(t *testing.T) {
	t.Error("test not implimented")
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

// omg, i can't beleve i haven't been putting this in my other tests, this is great
func init() {
	var check = func(e error, msg string) {
		if e != nil {
			fmt.Printf("test setup failed: %s: %s\n", e, msg)
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

	raw := `{"Name":"joe","Email":"nojoe@mail.com","Address":{"Street":"","City":"","State":"","Zip":""},"Card":{"Number":"","Expiration":"","CVV":""},"Service":"Carryout","MyOrders":null}`
	// raw, err := ioutil.ReadFile(filepath.Join(dir, "testconfig"))
	// check(err, "read config file")
	err = json.Unmarshal([]byte(raw), cfg)
	check(err, "json")
}

func testcleanup() {
	var check = func(e error, msg string) {
		if e != nil {
			fmt.Printf("test cleanup failed: %s: %s\n", e, msg)
			os.Exit(1)
		}
	}

	check(db.Close(), "database close")

	wd, err := os.Getwd()
	check(err, "cwd")

	check(
		os.RemoveAll(filepath.Join(wd, "testdata")),
		"delete data",
	)
}
