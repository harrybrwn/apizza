package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func testConfig(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg.printAll(buf)
	output := string(buf.Bytes())

	phrases := []string{
		"Service Carryout",
		"Name joe",
		"Email nojoe@mail.com",
		"City:Washington DC",
		"Street:1600 Pennsylvania Ave NW",
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

func testConfigGet(t *testing.T) {
	c := newConfigGet().(*configGetCmd)

	buf := &bytes.Buffer{}
	c.setOutput(buf)

	if err := c.run(c.command(), []string{"name"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "joe\n" {
		t.Error("wrong name output")
	}
	buf.Reset()

	if err := c.run(c.command(), []string{"email"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "nojoe@mail.com\n" {
		t.Error("wrong email config output")
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
}
