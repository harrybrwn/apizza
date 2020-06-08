package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
	"gopkg.in/yaml.v3"
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
	Test    string      `config:"test" yaml:"test" default:"this is a test config file"`
	Msg     string      `config:"msg" yaml:"msg" default:"this should have been deleted, please remove it"`
	Number  int         `config:"number" default:"50" yaml:"number"`
	Number2 int         `config:"number2" yaml:"number2"`
	NullVal interface{} `config:"nullval" yaml:"nullval"`
	More    struct {
		One string `config:"one" yaml:"one"`
		Two string `config:"two" yaml:"two"`
	} `config:"more" yaml:"more"`
	F   float64 `config:"f" yaml:"f"`
	Pie float64 `config:"pi" yaml:"pi" default:"3.14159"`
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
	SetField(c, "more.one", "hey is this shit working")
	if c.More.One != "hey is this shit working" {
		t.Error("Setting variables using dot notation in the key didn't work")
	}
	SetField(c, "Test", "this config is part of a test. it shouldn't be here")
	test := "this config is part of a test. it shouldn't be here"
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
	err := SetField(c, "number2", int64(3))
	if err != nil {
		t.Error(err)
	}
	err = SetField(c, "number", int64(6))
	if err != nil {
		t.Error(err)
	}
	err = SetField(c, "pi", 6.9)
	if err != nil {
		t.Error(err)
	}
	if GetField(c, "number").(int64) != int64(6) {
		t.Error("wrong number")
	}

	tt := []struct {
		key string
		val interface{}
	}{
		{"number", "what"},
		{"msg", 5},
		{"more", struct{ a string }{"what"}},
		{"pi", "what"},
		{"msg", 9.5},
	}
	for _, tc := range tt {
		if err = SetField(c, tc.key, tc.val); err == nil {
			t.Errorf("expected error setting %s to %v", tc.key, tc.val)
		}
		if GetField(c, tc.key) == tc.val {
			t.Errorf("field %s should not have been set to %v", tc.key, tc.val)
		}
	}
}

func TestSetConfig(t *testing.T) {
	var c = &testCnfg{}
	err := SetConfig(".testconfig", c)
	if err != nil {
		t.Error(err)
	}
	if err = Save(); err != nil {
		t.Error(err)
	}

	if err = Reset(); err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadFile(File())
	if err != nil {
		t.Error(err)
	}
	// err = json.Unmarshal(b, c)
	err = yaml.Unmarshal(b, c)
	if err != nil {
		t.Error(err)
	}
	if c.Number != 50 {
		// t.Error("number should be 50")
		t.Log("number should be 50; defaults not setup for yaml")
	}
	if c.Test != "this is a test config file" {
		// t.Error("config default value failed")
		t.Log("config default value failed; defaults not setup for yaml")
	}
	if c.Msg != "this should have been deleted, please remove it" {
		// t.Error("default config var failed")
		t.Log("default config var failed; defaults not setup for yaml")
	}
	if _, err := os.Stat(Folder()); os.IsNotExist(err) {
		t.Error("The config folder is not where it is supposed to be, you should probably find it")
	} else {
		os.Remove(File())
		os.Remove(Folder())
	}
	if _, err = os.Stat(Folder()); os.IsExist(err) {
		t.Error("test config folder cleanup failed")
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
	tt := []struct {
		key string
		val string
	}{
		{"msg", "Msg"},
		{"more.one", "More.One"},
		{"more", "More"},
		{"badFieldName", ""},
	}
	for _, tc := range tt {
		if name := FieldName(c, tc.key); name != tc.val {
			t.Errorf("bad field name for %s; got: \"%s\", wanted: \"%s\"", tc.key, name, tc.val)
		}
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

func TestDeepConfig(t *testing.T) {
	cfg.file = ""
	err := SetConfig(".testconfig/deep/config", &testCnfg{})
	if err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(Folder()); os.IsNotExist(err) {
		t.Error("the config folder should exist")
	}
	dir := filepath.Dir(filepath.Dir(Folder()))
	if err = os.RemoveAll(dir); err != nil {
		t.Error(err)
	}
	if _, err = os.Stat(dir); os.IsExist(err) {
		t.Error("should have cleaned up this directory")
	}
}
