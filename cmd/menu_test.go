package cmd

import (
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestMenuRun(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	defer r.CleanUp()
	c := NewMenuCmd(r).(*menuCmd)

	tests.Check(c.Run(c.Cmd(), []string{}))
	c.item = "not a thing"
	tests.Exp(c.Run(c.Cmd(), []string{}))
	c.item = "10SCREEN"
	tests.Check(c.Run(c.Cmd(), []string{}))
	c.item = ""
	c.toppings = true
	tests.Check(c.Run(c.Cmd(), []string{}))
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
