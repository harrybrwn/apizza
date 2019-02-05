package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func stackTrace() {
	fpcs := make([]uintptr, 1)
	for i := 0; runtime.Callers(i, fpcs) != 0; i++ {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		_, file, line, _ := runtime.Caller(i)
		if file != "" && line > 0 {
			fmt.Print(file, ": ", line, ", ", fun.Name(), "\n")
		}
	}
}

type testCnfg struct {
	Test string `config:"test" default:"\"this is a test config file\""`
	Msg  string `config:"msg" default:"\"this should have been deleted, please remove it\""`
	More struct {
		One string `config:"one"`
		Two string `config:"two"`
	} `config:"more"`
}

func (c *testCnfg) Get(key string) interface{}            { return nil }
func (c *testCnfg) Set(key string, val interface{}) error { return nil }

func TestConfigGetandSet(t *testing.T) {
	var c = &testCnfg{}
	if Get(c, "msg") != Get(c, "Msg") {
		t.Error("the Get function should auto convert 'msg' to 'Msg'.")
	}
	if Get(c, "msg").(string) != c.Msg {
		t.Error("The Get function should be returning the same value as acessing the struct literal.")
	}
	Set(c, "more.one", "hey is this shit workin")
	if c.More.One != "hey is this shit workin" {
		t.Error("Setting variables using dot notation in the key didn't work")
	}
	Set(c, "Test", "this config is part of a test. it should't be here")
	test := "this config is part of a test. it should't be here"
	if c.Test != test {
		t.Errorf("Error in 'Set':\n\twant: %s\n\tgot: %s", test, c.Test)
	}
	if Get(c, "invalidkey") != nil {
		t.Error("an invalid key should have resulted in a nil value")
	}
	if err := Set(c, "invalidkey", ""); err == nil {
		t.Error("this should have raised an error")
	}
}

func TestSetConfig(t *testing.T) {
	var c = &testCnfg{}
	err := SetConfig(".testconfig", c)
	if err != nil {
		if e, ok := err.(Error); ok {
			fmt.Println(e.file, e.fun, e.line)
		}
		t.Error(err)
	}

	b, err := ioutil.ReadFile(Folder() + "/config.json")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(b, c)
	if err != nil {
		t.Error(err)
	}
	if c.Test != "this is a test config file" {
		t.Error("config default value failed")
	}
	if c.Msg != "this should have been deleted, please remove it" {
		t.Error("default config var failed")
	}
	if _, err := os.Stat(Folder()); os.IsNotExist(err) {
		t.Error("The config folder is not where it is supposed to be, you should probably find it")
	} else {
		os.Remove(Folder() + "/config.json")
		os.Remove(Folder())
	}
}

func TestEmptyConfig(t *testing.T) {
	elem := reflect.ValueOf(&testCnfg{}).Elem()
	expected := `{
    "Test": "this is a test config file",
    "Msg": "this should have been deleted, please remove it",
    "More": {
		"One": "",
		"Two": ""
    }
}`
	expected = strings.Replace(expected, "\t", "    ", -1)
	raw := emptyConfig(elem.Type(), 0)

	if raw != expected {
		t.Errorf("the emptyConfig funtion returned:\n%s\nand should have returned\n%s", raw, expected)
	}
}
