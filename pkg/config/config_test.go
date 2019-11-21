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

	"github.com/harrybrwn/apizza/pkg/tests"
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
	Test    string      `config:"test" default:"this is a test config file"`
	Msg     string      `config:"msg" default:"this should have been deleted, please remove it"`
	Number  int         `config:"number" default:"50"`
	Number2 int         `config:"number2"`
	NullVal interface{} `config:"nullval"`
	More    struct {
		One string `config:"one"`
		Two string `config:"two"`
	} `config:"more"`
	F   float64 `config:"f"`
	Pie float64 `config:"pi" default:"3.14159"`
}

func (c *testCnfg) Get(key string) interface{}            { return nil }
func (c *testCnfg) Set(key string, val interface{}) error { return nil }

func TestConfigGetandSet(t *testing.T) {
	var c = &testCnfg{}
	cfg = configfile{conf: c}
	if GetField(c, "msg") != GetField(c, "Msg") {
		t.Error("the Get function should auto convert 'msg' to 'Msg'.")
	}
	if Get("msg") != GetField(c, "Msg") {
		t.Error("the Get function should auto convert 'msg' to 'Msg'.")
	}
	if GetField(c, "msg").(string) != c.Msg {
		t.Error("The Get function should be returning the same value as acessing the struct literal.")
	}
	SetField(c, "more.one", "hey is this shit workin")
	if c.More.One != "hey is this shit workin" {
		t.Error("Setting variables using dot notation in the key didn't work")
	}
	SetField(c, "Test", "this config is part of a test. it should't be here")
	test := "this config is part of a test. it should't be here"
	if c.Test != test {
		t.Errorf("Error in 'Set':\n\twant: %s\n\tgot: %s", test, c.Test)
	}
	if GetField(c, "invalidkey") != nil {
		t.Error("an invalid key should have resulted in a nil value")
	}
	if err := SetField(c, "invalidkey", ""); err == nil {
		t.Error("this should have raised an error")
	}
	if v := GetField(c, "number"); v == nil {
		t.Errorf("Get(c, `number`) should have returned and integer, got %T", v)
	}
	if GetField(c, "More") == nil {
		t.Error("didn't get struct value")
	}
	err := SetField(c, "number", "what")
	if err == nil {
		t.Error("should have returned a bad type error")
	}
	err = SetField(c, "number2", int64(3))
	if err != nil {
		t.Error(err)
	}
	err = SetField(c, "number", int64(6))
	if err != nil {
		t.Error(err)
	}
	if GetField(c, "number").(int64) != int64(6) {
		t.Error("wrong number")
	}

	err = SetField(c, "msg", 5)
	if err == nil {
		t.Error("expected error")
	}
	err = SetField(c, "more", struct{ a string }{"what"})
	if err == nil {
		t.Error("expected error")
	}
	err = SetField(c, "pi", 6.9)
	if err != nil {
		t.Error(err)
	}
	err = SetField(c, "pi", "what")
	if err == nil {
		t.Error("expected error")
	}
	err = SetField(c, "msg", 9.5)
	if err == nil {
		t.Error("expected error")
	}
}

func TestSetConfig(t *testing.T) {
	var c = &testCnfg{}
	err := SetConfig(".testconfig", c)
	if err != nil {
		t.Error(err)
	}
	err = Save()
	if err != nil {
		t.Error(err)
	}
	err = SetConfig(".testconfig", c)
	if err == nil {
		t.Error("The second call to SetConfig should have returned an error")
	}
	tests.Compare(t, err.Error(), "cannot set multiple config files")

	err = Reset()
	if err != nil {
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
	if c.Number != 50 {
		t.Error("number should be 50")
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
		os.Remove(File())
		os.Remove(Folder())
	}
}

func TestEmptyConfig(t *testing.T) {
	elem := reflect.ValueOf(&testCnfg{}).Elem()
	expected := `{
    "Test": "this is a test config file",
	"Msg": "this should have been deleted, please remove it",
	"Number": 50,
	"Number2": 0,
	"NullVal": null,
    "More": {
		"One": "",
		"Two": ""
	},
	"F": 0.0,
	"Pie": 3.14159
}`
	expected = strings.Replace(expected, "\t", "    ", -1)
	raw := emptyJSONConfig(elem.Type(), 0)

	if raw != expected {
		t.Errorf("the emptyConfig function returned:\n%s\nand should have returned\n%s", raw, expected)
	}
}

func TestIsField(t *testing.T) {
	var c Config = &testCnfg{}

	if !IsField(c, "more.one") {
		t.Error("should register as field")
	}
	if IsField(c, "not_a_field") {
		t.Error("should not register as a field")
	}
}

func TestFieldName(t *testing.T) {
	var c Config = &testCnfg{}

	if FieldName(c, "msg") != "Msg" {
		t.Error("bad field name")
	}
	if FieldName(c, "more.one") != "More.One" {
		t.Error("bad field name")
	}
	if FieldName(c, "more") != "More" {
		t.Error("bad field name")
	}
	if FieldName(c, "badFieldName") != "" {
		t.Error("bad field name")
	}
}

func TestGeters(t *testing.T) {
	var c Config = new(testCnfg)
	if err := SetNonFileConfig(c); err != nil {
		t.Error(err)
	}

	if GetString("msg") != "this should have been deleted, please remove it" {
		t.Error("bad result; GetString")
	}
	if GetString("test") != "this is a test config file" {
		t.Error("bad result; GetString")
	}
	if GetInt("number") != 50 {
		t.Error("bad result; GetInt")
	}
	if GetFloat("pi") != 3.14159 {
		t.Error("bad result; GetFloat")
	}
	if Object().(*testCnfg).Pie != GetFloat("pi") {
		t.Error("no")
	}
	if err := Set("pi", 3.14159*2); err != nil {
		t.Error(err)
	}
	if GetFloat("pi") != 3.14159*2 {
		t.Error("wrong result; GetFloat after Set")
	}
}

func TestPrintAll(t *testing.T) {
	var c Config = &testCnfg{}
	expected := "test: \"\"\nmsg: \"\"\nnumber: 0\nnumber2: 0\nnullval: null\nmore:\n  one: \"\"\n  two: \"\"\nf: 0\npi: 0\n"

	tests.CompareOutput(t, expected, func() {
		if err := PrintAll(c); err != nil {
			t.Error(err)
		}
	})

	tests.CompareOutput(t, expected, func() {
		if err := FprintAll(os.Stdout, c); err != nil {
			t.Error(err)
		}
	})
}

func TestEditor(t *testing.T) {
	cfg = configfile{file: tests.TempFile()}
	if err := os.Setenv("EDITOR", "cat"); err != nil {
		t.Error(err)
	}
	data := "testing my editor func"
	err := ioutil.WriteFile(cfg.file, []byte(data), 0777)
	if err != nil {
		t.Error(err)
	}

	tests.CompareOutput(t, data, func() {
		if err := Edit(); err != nil {
			t.Error(err)
		}
	})

	data = "written to temp file for tests"
	f := tests.TempFile()
	err = ioutil.WriteFile(f, []byte(data), 0777)
	if err != nil {
		t.Error(err)
	}

	tests.CompareOutput(t, data, func() {
		if err := EditFile(f); err != nil {
			t.Error(err)
		}
	})

	FileHasChanged()

	if cfg.changed == false {
		t.Error("FileHasChanged should make this true")
	}
	if cfg.save() != nil {
		t.Error("should have returned nil")
	}

}
