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
		"The menu command will show the dominos menu. To show a subdivition of the menu, ",
		"give an item or category to the --category and --item flags or give them as an ",
		"argument to the command itself.",
	}
	s := `The menu command will show the dominos menu. To show a subdivition of the menu, give an item or category to the --category and --item flags or give them as an argument to the command itself.`
	for i, line := range FormatLine(s, 80) {
		if exp[i] != line {
			t.Error("wrong line format")
		}
	}

	expected := "The menu command will show the dominos menu. To show a subdivition of the menu, \n    give an item or category to the --category and --item flags or give them as an \n    argument to the command itself."
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
	o := testStore.MakeOrder("Jimbo", "Jones", "blahblah@aol.com")
	o.SetName("TestOrder")
	pizza, err := testStore.GetVariant("14SCREEN")
	err = errs.Pair(err, o.AddProduct(pizza))
	if err != nil {
		t.Error(err)
	}

	buf := new(bytes.Buffer)
	SetOutput(buf)
	err = PrintOrder(o, false, false)
	if err != nil {
		t.Error(err)
	}
	tests.CompareV(t, buf.String(), "  TestOrder -  14SCREEN, \n")
	buf.Reset()
	if err = PrintOrder(o, true, false); err != nil {
		t.Error(err)
	}
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

	if err = PrintOrder(o, true, true); err != nil {
		t.Error(err)
	}
	tests.Compare(t, buf.String(), expected+"  price:   $20.15\n")
	ResetOutput()
}

func TestPrintItems(t *testing.T) {

}
