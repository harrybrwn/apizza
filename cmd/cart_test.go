package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func testOrderNew(t *testing.T, buf *bytes.Buffer, cmds ...cli.CliCommand) {
	cart, add := cmds[0], cmds[1]
	add.Cmd().ParseFlags([]string{"--name=testorder", "--product=12SCMEATZA"})
	err := add.Run(add.Cmd(), []string{})
	if err != nil {
		t.Error(err)
	}
	buf.Reset()

	if err := cart.Run(cart.Cmd(), []string{"testorder"}); err != nil {
		t.Error(err)
	}
	expected := `testorder
  products:
    Medium (12") Hand Tossed MeatZZa
      code:     12SCMEATZA
      options:
         B: 1/1 1
         C: 1/1 1.5
         H: 1/1 1
         P: 1/1 1
         S: 1/1 1
         X: 1/1 1
      quantity: 1
  storeID: 4336
  method:  Carryout
  address: 1600 Pennsylvania Ave NW
  	       Washington, DC 20500
`
	tests.Compare(t, buf.String(), strings.Replace(expected, "\t", "  ", -1))
}

func testAddOrder(t *testing.T, buf *bytes.Buffer, cmds ...cli.CliCommand) {
	cart, add := cmds[0], cmds[1]
	tests.Check(add.Run(add.Cmd(), []string{"testing"}))
	if buf.String() != "" {
		t.Errorf("wrong output: should have no output: '%s'", buf.String())
	}
	buf.Reset()
	cart.Cmd().ParseFlags([]string{"-d"})
	tests.Check(cart.Run(cart.Cmd(), []string{"testing"}))
	buf.Reset()
}

func testOrderNewErr(t *testing.T, buf *bytes.Buffer, cmds ...cli.CliCommand) {
	if err := cmds[0].Run(cmds[0].Cmd(), []string{}); err == nil {
		t.Error("expected error")
	}
}

func testOrderRunAdd(t *testing.T, buf *bytes.Buffer, cmds ...cli.CliCommand) {
	cart := cmds[0]
	tests.Check(cart.Run(cart.Cmd(), []string{}))
	tests.Compare(t, buf.String(), "Your Orders:\n  testorder\n")
	buf.Reset()
	cart.Cmd().ParseFlags([]string{"--add", "10SCPFEAST,PSANSAMV"})
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	tests.Compare(t, buf.String(), "order successfully updated.\n")
}

func testOrderPriceOutput(cart *cartCmd, buf *bytes.Buffer, t *testing.T) {
	cart.price = true

	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	tests.Exp(cart.Run(cart.Cmd(), []string{"to-many", "args"}))
	m := cart.Menu()
	m2 := cart.Menu()
	if m != m2 {
		t.Error("should have cached the menu")
	}
}

func testOrderRunDelete(cart *cartCmd, buf *bytes.Buffer, t *testing.T) {
	cart.delete = true
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	tests.Compare(t, buf.String(), "testorder successfully deleted.\n")
	cart.delete = false
	buf.Reset()
	cart.Cmd().ParseFlags([]string{})
	tests.Check(cart.Run(cart.Cmd(), []string{}))
	tests.Compare(t, buf.String(), "No orders saved.\n")
	buf.Reset()
	tests.Exp(cart.Run(cart.Cmd(), []string{"not_a_real_order"}))
	cart.topping = false
	cart.validate = true
	tests.Check(cart.Run(cart.Cmd(), []string{}))
}

func testAddToppings(cart *cartCmd, buf *bytes.Buffer, t *testing.T) {
	cart.add = []string{"10SCREEN"}
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	cart.add = nil

	cart.product = "10SCREEN"
	cart.add = []string{"P", "K"}
	cart.topping = false
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))

	cart.product = ""
	cart.add = []string{}
	cart.topping = false
	buf.Reset()
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))

	expected := `Small (10") Hand Tossed Pizza
      code:     10SCREEN
      options:
         C: 1/1 1
         K: 1/1 1.0
         P: 1/1 1.0
         X: 1/1 1
      quantity: 1`

	if !strings.Contains(buf.String(), expected) {
		t.Error("bad output")
	}
	buf.Reset()

	cart.topping = false
	cart.product = "10SCREEN"
	cart.remove = "C"
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	buf.Reset()
	cart.topping = false
	cart.product = ""
	cart.remove = ""
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	expected = `    Small (10") Hand Tossed Pizza
      code:     10SCREEN
      options:
         C: 1/1 1
         K: 1/1 1.0
         P: 1/1 1.0
         X: 1/1 1
      quantity: 1`
	if !strings.Contains(buf.String(), expected) {
		fmt.Println("got:")
		fmt.Println(buf.String())
		fmt.Println("expected:")
		fmt.Print(expected)
		t.Error("bad output")
	}
	buf.Reset()

	cart.topping = false
	cart.remove = "10SCREEN"
	tests.Check(cart.Run(cart.Cmd(), []string{"testorder"}))
	if strings.Contains(buf.String(), expected) {
		t.Error("bad output")
	}
}

func TestOrder(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	defer r.CleanUp()

	ordercmd := NewOrderCmd(r)
	err := ordercmd.Run(ordercmd.Cmd(), []string{})
	tests.Check(err)
	tests.Exp(ordercmd.Run(ordercmd.Cmd(), []string{"one", "two"}))
	tests.Exp(ordercmd.Run(ordercmd.Cmd(), []string{"anorder"}))
	cmd := ordercmd.(*orderCmd)
	cmd.cvv = 100
	tests.Exp(cmd.Run(cmd.Cmd(), []string{"nothere"}))
	cmd.cvv = 0
}

func TestEitherOr(t *testing.T) {
	if eitherOr("one", "") != "one" {
		t.Error("wrong result from 'eitherOr'")
	}
	if eitherOr("", "two") != "two" {
		t.Error("wrong result from 'eitherOr'")
	}
	if eitherOr("a", "b") != "a" {
		t.Error("wrong result from 'eitherOr'")
	}
}
