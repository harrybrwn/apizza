package cmd

import (
	"bytes"
	"testing"
)

func testMenuRun(t *testing.T) {
	c := newBuilder().newMenuCmd().(*menuCmd)
	c.SetOutput(&bytes.Buffer{})
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	c.item = "not a thing"
	if err := c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("should raise error")
	}
	c.dstore = nil
	c.item = "10SCREEN"
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	c.item = ""
	c.toppings = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
}

func testFindProduct(t *testing.T) {
	c := newBuilder().newMenuCmd().(*menuCmd)
	buf := &bytes.Buffer{}
	c.SetOutput(buf)
	if err := db.UpdateTS("menu", c); err != nil {
		t.Error(err)
	}
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
