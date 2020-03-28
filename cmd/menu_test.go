package cmd

import (
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
)

func TestMenuRun(t *testing.T) {
	r := cmdtest.NewRecorder()
	defer r.CleanUp()
	c := NewMenuCmd(r).(*menuCmd)

	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	c.item = "not a thing"
	if err := c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("should raise error")
	}
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

func TestFindProduct(t *testing.T) {
	r := cmdtest.NewRecorder()
	defer r.CleanUp()
	c := NewMenuCmd(r).(*menuCmd)

	if err := r.DB().UpdateTS("menu", c); err != nil {
		t.Error(err)
	}
	c.all = true
	if err := c.printMenu(c.Output(), ""); err != nil { // yes, this is supposed to be an empty string... in this case
		t.Error(err)
	}
	r.ClearBuf()
	c.printToppings()
	if len(r.Out.Bytes()) < 1000 {
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
}
