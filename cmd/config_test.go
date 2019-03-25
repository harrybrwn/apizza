package cmd

import (
	"bytes"
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
	c.SetOutput(buf)

	c.file = true
	if err = c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	c.file = false
	tests.Compare(t, string(buf.Bytes()), "\n")
	buf.Reset()

	c.dir = true
	if err = c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), "\n")
	c.dir = false
	buf.Reset()

	c.getall = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
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

	cmdUseage := c.Cmd().UsageString()
	if err = c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), cmdUseage)
	buf.Reset()

	if err = c.basecmd.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), c.Cmd().UsageString())
}

func testConfigGet(t *testing.T) {
	c := newConfigGet().(*configGetCmd)

	buf := &bytes.Buffer{}
	c.SetOutput(buf)

	if err := c.Run(c.Cmd(), []string{"email", "name"}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), "nojoe@mail.com\njoe\n")
	buf.Reset()

	if err := c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}

	if err := c.Run(c.Cmd(), []string{"nonExistantKey"}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func testConfigSet(t *testing.T) {
	c := newConfigSet().(*configSetCmd)

	if err := c.Run(c.Cmd(), []string{"name=someNameOtherThanJoe"}); err != nil {
		t.Error(err)
	}
	if cfg.Name != "someNameOtherThanJoe" {
		t.Error("did not set the name correctly")
	}

	if err := c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}

	if err := c.Run(c.Cmd(), []string{"nonExistantKey=someValue"}); err == nil {
		t.Error("expected error")
	}

	if err := c.Run(c.Cmd(), []string{"badformat"}); err == nil {
		t.Error(err)
	} else if err.Error() != "use '<key>=<value>' format (no spaces)" {
		t.Error("wrong error message, got:", err.Error())
	}
}
