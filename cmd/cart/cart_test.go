package cart

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestToppings(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	defer r.CleanUp()

	cart := New(r)
	cart.SetOutput(ioutil.Discard)
	order := cart.finder.Store().NewOrder()
	testo := cmdtest.NewTestOrder()
	order.FirstName = testo.FirstName
	order.LastName = testo.LastName
	order.OrderName = cmdtest.OrderName

	code := "12SCREEN"
	p := &dawg.OrderProduct{ItemCommon: dawg.ItemCommon{
		Code: code,
		Tags: map[string]interface{}{"DefaultToppings": "X=1,C=1"},
	},
		Opts: map[string]interface{}{},
		Qty:  1,
	}
	order.Products = []*dawg.OrderProduct{p}
	tests.Fatal(data.SaveOrder(order, cart.out, r.DataBase))
	tests.Fatal(cart.SetCurrentOrder(cmdtest.OrderName))

	if cart.CurrentOrder == order {
		t.Error("pointers are the save, cart should have gotten order from disk")
	}
	if cart.CurrentOrder.Products[0].Code != code {
		t.Error("did not store the order correctly")
	}
	if _, ok := cart.CurrentOrder.Products[0].Tags["DefaultToppings"]; !ok {
		t.Error("should have the default toppings")
	}

	tests.Check(cart.AddToppings(code, []string{
		"B:left",
		"Rr:riGHt:2",
		"K:Full:1.5",
		"Pm:LefT:2.0",
	}))

	checktoppings := func(opts map[string]interface{}) {
		for _, tc := range []struct {
			top, side, amount string
		}{
			{"B", dawg.ToppingLeft, "1.0"},
			{"Rr", dawg.ToppingRight, "2.0"},
			{"K", dawg.ToppingFull, "1.5"},
			{"Pm", dawg.ToppingLeft, "2.0"},
		} {
			top, ok := opts[tc.top]
			if !ok {
				t.Error("options should have", tc.top, "as a topping")
			}
			topping, ok := top.(map[string]string)
			if !ok {
				t.Errorf("topping should be a map; its a %T\n", top)
				continue
			}
			if amt, ok := topping[tc.side]; !ok {
				t.Errorf("topping side expected was %s, got %s\n", tc.side, amt)
			} else {
				if amt != tc.amount {
					t.Errorf("wrong amount; want %s, got %s\n", tc.amount, amt)
				}
			}
		}
	}
	checktoppings(cart.CurrentOrder.Products[0].Opts)

	tests.Check(cart.SaveAndReset())
	o := dawg.Order{}
	bytes, err := r.DataBase.Get(data.OrderPrefix + cmdtest.OrderName)
	tests.Check(err)
	tests.Check(json.Unmarshal(bytes, &o))
	// checktoppings(o.Products[0].Opts)
}
