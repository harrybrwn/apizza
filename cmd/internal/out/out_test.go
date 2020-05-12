package out

import (
	"bytes"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestFormatLine(t *testing.T) {
	exp := []string{
		"The menu command will show the dominos menu. To show a subdivision of the menu, ",
		"give an item or category to the --category and --item flags or give them as an ",
		"argument to the command itself.",
	}
	s := `The menu command will show the dominos menu. To show a subdivision of the menu, give an item or category to the --category and --item flags or give them as an argument to the command itself.`
	for i, line := range FormatLine(s, 80) {
		if exp[i] != line {
			t.Error("wrong line format")
		}
	}

	expected := "The menu command will show the dominos menu. To show a subdivision of the menu, \n    give an item or category to the --category and --item flags or give them as an \n    argument to the command itself."
	tests.Compare(t, FormatLineIndent(s, 80, 4), expected)
}

func TestTmpl(t *testing.T) {
	s := "msg: {{.Msg}}"
	buf := new(bytes.Buffer)
	err := tmpl(buf, s, struct{ Msg string }{"hello tdd"})
	if err != nil {
		t.Error(err)
	}
	if buf.String() != "msg: hello tdd" {
		t.Error("got wrong template result")
	}
}

var testStore *dawg.Store

func init() {
	var err error
	testStore, err = dawg.NearestStore(cmdtest.TestAddress(), dawg.Delivery)
	if err != nil {
		panic(err)
	}
}

func TestPrintOrder(t *testing.T) {
	tests.InitHelpers(t)
	o := testStore.MakeOrder("Jimbo", "Jones", "blahblah@aol.com")
	o.SetName("TestOrder")
	pizza, err := testStore.GetVariant("14SCREEN")
	tests.Check(errs.Pair(err, o.AddProduct(pizza)))

	buf := new(bytes.Buffer)
	SetOutput(buf)
	tests.Check(PrintOrder(o, false, false, false))
	tests.CompareV(t, buf.String(), "  TestOrder -  14SCREEN, \n")
	buf.Reset()
	tests.Check(PrintOrder(o, true, false, false))
	expected := `TestOrder
  products:
    Large (14") Hand Tossed Pizza
      code:     14SCREEN
      options:
         C: full 1
         X: full 1
      quantity: 1
  storeID: 4336
  method:  Delivery
  address: 1600 Pennsylvania Ave NW
           Washington, DC 20500
`
	tests.CompareV(t, buf.String(), expected)
	buf.Reset()
	tests.Check(PrintOrder(o, true, false, true))
	tests.Compare(t, buf.String(), expected+"  price:   $20.15\n")
	ResetOutput()
}

func TestPrintItems(t *testing.T) {
	tests.InitHelpers(t)
	menu, err := testStore.Menu()
	tests.Check(err)
	buf := new(bytes.Buffer)
	SetOutput(buf)
	defer ResetOutput()

	v, err := menu.GetVariant("14SCREEN")
	tests.Check(err)
	tests.Check(ItemInfo(v, menu))
	expected := `Large (14") Hand Tossed Pizza
  Code: 14SCREEN
  Category: Pizza
  Toppings:
    Robust Inspired Tomato Sauce (X): full 1
    Cheese (C): full 1
  Price: 13.99
  Parent Product: 'Pizza' [S_PIZZA]
`
	// we are not testing for the output of the toppings section
	// because the order of the toppings relies on a map and we cannot guarantee
	// that the toppings will always be in the same order.
	tests.Compare(t, buf.String()[:76], expected[:76])
	tests.Compare(t, buf.String()[147:], expected[147:])
	buf.Reset()
}

func TestPrintMenu(t *testing.T) {
	tests.InitHelpers(t)
	menu, err := testStore.Menu()
	tests.Check(err)
	buf := new(bytes.Buffer)
	SetOutput(buf)
	defer ResetOutput()

	tests.Check(PrintMenu(menu.Categorization.Food, 0, menu))
	if buf.Len() < 9000 {
		t.Error("the menu output seems a bit too short")
	}
	buf.Reset()
	tests.Check(PrintMenu(menu.Categorization.Preconfigured, 0, menu))
	if buf.Len() < 1000 {
		t.Error("menu output is too short")
	}
}
