package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func testConfigStruct(t *testing.T) {
	if cfg.Get("name").(string) != "joe" {
		t.Error("wrong value")
	}
	if err := cfg.Set("name", "not joe"); err != nil {
		t.Error(err)
	}
	if cfg.Get("Name").(string) != "not joe" {
		t.Error("wrong value")
	}
	if err := cfg.Set("name", "joe"); err != nil {
		t.Error(err)
	}
}

func testConfigCmd(t *testing.T) {
	var err error
	c := newConfigCmd().(*configCmd)
	buf := &bytes.Buffer{}
	c.setOutput(buf)

	c.file = true
	if err = c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
	c.file = false
	tests.Compare(t, string(buf.Bytes()), "\n")
	buf.Reset()

	c.dir = true
	if err = c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), "\n")
	c.dir = false
	buf.Reset()

	builder := newBuilder()
	c.resetCache = true
	if err = c.run(c.command(), []string{}); err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), builder.dbPath()) {
		t.Error("error given does not match dbPath()")
	}
	c.resetCache = false
	buf.Reset()

	c.getall = true
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}

	expected := `name: "joe"
email: "nojoe@mail.com"
address: 
  street: "1600 Pennsylvania Ave NW"
  cityname: "Washington DC"
  state: ""
  zipcode: "20500"
card: 
  number: ""
  expiration: ""
  cvv: ""
service: "Carryout"
`
	tests.Compare(t, string(buf.Bytes()), expected)
	c.getall = false
	buf.Reset()

	cmdUseage := c.command().UsageString()
	if err = c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != cmdUseage {
		t.Error("usage does not match")
	}
	buf.Reset()

	if err = c.basecmd.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != c.command().UsageString() {
		t.Error("usage does not match")
	}
}

func testConfigGet(t *testing.T) {
	c := newConfigGet().(*configGetCmd)

	buf := &bytes.Buffer{}
	c.setOutput(buf)

	if err := c.run(c.command(), []string{"email", "name"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "nojoe@mail.com\njoe\n" {
		t.Error("wrong email config output")
	}
	buf.Reset()

	if err := c.run(c.command(), []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}

	if err := c.run(c.command(), []string{"nonExistantKey"}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func testConfigSet(t *testing.T) {
	c := newConfigSet().(*configSetCmd)

	if err := c.run(c.command(), []string{"name=someNameOtherThanJoe"}); err != nil {
		t.Error(err)
	}
	if cfg.Name != "someNameOtherThanJoe" {
		t.Error("did not set the name correctly")
	}

	if err := c.run(c.command(), []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}

	if err := c.run(c.command(), []string{"nonExistantKey=someValue"}); err == nil {
		t.Error("expected error")
	}

	if err := c.run(c.command(), []string{"badformat"}); err == nil {
		t.Error(err)
	} else if err.Error() != "use '<key>=<value>' format (no spaces)" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func TestAddressStr(t *testing.T) {
	a := &address{
		Street: "1600 Pennsylvania Ave NW", CityName: "Washington",
		State: "DC", Zipcode: "20500",
	}
	expected := `1600 Pennsylvania Ave NW
Washington, DC 20500`

	formatted := addressStr(a)
	if addressStr(a) != expected {
		t.Errorf("unexpected output...\ngot:\n%s\nwanted:\n%s\n", formatted, expected)
	}
}
