package cmd

import (
	"bytes"
	"testing"
)

func testMenuRun(t *testing.T) {
	c := newBuilder().newMenuCmd().(*menuCmd)
	buf := &bytes.Buffer{}
	c.output = buf
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}

	c.item = "not a thing"
	if err := c.run(c.command(), []string{}); err == nil {
		t.Error("should raise error")
	}

	store = nil
	c.item = "10SCREEN"
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}

	c.item = ""
	c.toppings = true
	if err := c.run(c.command(), []string{}); err != nil {
		t.Error(err)
	}
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
