package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func testConfigStruct(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg.printAll(buf)
	output := string(buf.Bytes())

	phrases := []string{
		"Service Carryout",
		"Name joe",
		"Email nojoe@mail.com",
		"Washington DC",
		"1600 Pennsylvania Ave NW",
		"20500",
	}

	for _, phrase := range phrases {
		if !strings.Contains(output, phrase) {
			t.Error("wrong output")
		}
	}

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
	if string(buf.Bytes()) != "" {
		t.Error("got unexpected output")
	}
	buf.Reset()

	c.dir = true
	if err = c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "" {
		t.Error("got unexpected output")
	}
	c.dir = false
	buf.Reset()

	builder := newBuilder()
	c.resetCache = true
	if err = c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("expected error")
	}
	if !strings.Contains(err.Error(), builder.dbPath()) {
		t.Error("error given does not match dbPath()")
	}
	c.resetCache = false
	buf.Reset()

	phrases := []string{
		"Service Carryout",
		"Name joe",
		"Email nojoe@mail.com",
		"Washington DC",
		"1600 Pennsylvania Ave NW",
		"20500",
	}
	c.getall = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	getAllOutput := string(buf.Bytes())
	for _, phrase := range phrases {
		if !strings.Contains(getAllOutput, phrase) {
			t.Error("wrong output")
		}
	}
	c.getall = false
	buf.Reset()

	cmdUseage := c.Cmd().UsageString()
	if err = c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != cmdUseage {
		t.Error("usage does not match")
	}
	buf.Reset()

	if err = c.basecmd.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != c.Cmd().UsageString() {
		t.Error("usage does not match")
	}
}

func testConfigGet(t *testing.T) {
	c := newConfigGet().(*configGetCmd)

	buf := &bytes.Buffer{}
	c.SetOutput(buf)

	if err := c.Run(c.Cmd(), []string{"email", "name"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "nojoe@mail.com\njoe\n" {
		t.Error("wrong email config output")
	}
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
