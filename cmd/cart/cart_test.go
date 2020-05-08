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

var testProduct = &dawg.OrderProduct{ItemCommon: dawg.ItemCommon{
	Code: "12SCREEN",
	Tags: map[string]interface{}{"DefaultToppings": "X=1,C=1"},
},
	Opts: map[string]interface{}{},
	Qty:  1,
}

func TestToppings(t *testing.T) {
	r, cart, order := setup(t)
	defer r.CleanUp()

	tests.Exp(addTopping("", testProduct))
	order.Products = []*dawg.OrderProduct{testProduct}
	tests.Fatal(data.SaveOrder(order, cart.out, r.DataBase))
	tests.Fatal(cart.SetCurrentOrder(cmdtest.OrderName))
	code := testProduct.Code

	if cart.CurrentOrder == order {
		t.Error("pointers are the save, cart should have gotten order from disk")
	}
	if cart.CurrentOrder.Products[0].Code != code {
		t.Error("did not store the order correctly")
	}
	if _, ok := cart.CurrentOrder.Products[0].Tags["DefaultToppings"]; !ok {
		t.Error("should have the default toppings")
	}

	tests.Check(cart.AddToppings(testProduct.Code, []string{
		"P",
		"B:left",
		"Rr:riGHt:2",
		"K:Full:1.5",
		"Pm:LefT:2.0",
	}))

	// TODO: add test cases that make sure errors are raised when bad inputs are given

	checktoppings := func(opts map[string]interface{}) {
		for _, tc := range []struct {
			top, side, amount string
		}{
			{"P", dawg.ToppingFull, "1.0"},
			{"B", dawg.ToppingLeft, "1.0"},
			{"Rr", dawg.ToppingRight, "2.0"},
			{"K", dawg.ToppingFull, "1.5"},
			{"Pm", dawg.ToppingLeft, "2.0"},
		} {
			top, ok := opts[tc.top]
			if !ok {
				t.Error("options should have", tc.top, "as a topping")
			}
			var amount string
			switch topping := top.(type) {
			case map[string]string:
				amount, ok = topping[tc.side]
				if !ok {
					t.Errorf("topping side expected was %s, got %s\n", tc.side, amount)
				}
			case map[string]interface{}:
				a, ok := topping[tc.side]
				if !ok {
					t.Errorf("topping side expected was %s, got %v\n", tc.side, a)
				}
				amount = a.(string)
			default:
				t.Fatal("expected a map[string]string or map[string]interface{}")
				continue
			}
			if !ok {
				t.Errorf("topping side expected was %s, got %s\n", tc.side, amount)
			} else if amount != tc.amount {
				t.Errorf("wrong amount; want %s, got %s\n", tc.amount, amount)
			}
		}
	}
	checktoppings(cart.CurrentOrder.Products[0].Opts)
	tests.Check(cart.Validate())

	tests.Check(cart.SaveAndReset())
	tests.Exp(cart.AddToppings(testProduct.Code, []string{"P"}))
	// tests.Check(cart.AddToppingsToOrder(cmdtest.OrderName, code, []string{"P"}))
	o := dawg.Order{}
	bytes, err := r.DataBase.Get(data.OrderPrefix + cmdtest.OrderName)
	tests.Check(err)
	tests.Check(json.Unmarshal(bytes, &o))
	tests.Check(cart.ValidateOrder(cmdtest.OrderName))

	checktoppings(o.Products[0].Opts)
}

func TestValidate_Err(t *testing.T) {
	r, cart, o := setup(t)
	defer r.CleanUp()

	o.Address = &dawg.StreetAddr{}
	b, err := json.Marshal(o)
	tests.Check(err)
	tests.Check(r.DataBase.Put(data.OrderPrefix+o.Name(), b))
	tests.Exp(cart.ValidateOrder(cmdtest.OrderName))
	tests.Exp(cart.ValidateOrder(""))
	tests.Check(cart.SetCurrentOrder(cmdtest.OrderName))
	tests.Exp(cart.Validate())
	tests.Exp(cart.Save())
	if cart.CurrentOrder == nil {
		t.Error("current order should not be nil")
	}
	orders, err := cart.ListOrders()
	tests.Check(err)
	if orders[0] != cmdtest.OrderName {
		t.Error("did not list correct order name")
	}
	orders, _ = cart.OrdersCompletion(nil, []string{}, "")
	if orders[0] != cmdtest.OrderName {
		t.Error("did not list correct order name")
	}
	tests.Exp(cart.SaveAndReset())
	if cart.CurrentOrder != nil {
		t.Error("current order should be nil")
	}
	if cart.Validate() != ErrNoCurrentOrder {
		t.Error("wrong error")
	}
	tests.Check(cart.DeleteOrder(cmdtest.OrderName))
	_, err = cart.GetOrder(cmdtest.OrderName)
	tests.Exp(err)
	tests.Exp(cart.ValidateOrder(cmdtest.OrderName))
}

func TestProducts(t *testing.T) {
	r, cart, order := setup(t)
	defer r.CleanUp()

	order.Products = []*dawg.OrderProduct{}
	b, err := json.Marshal(order)
	tests.Check(err)
	tests.Check(r.DataBase.Put(data.OrderPrefix+order.Name(), b))
	codes := []string{"12SCREEN", "W08PBBQW", "10THIN", "10SCMEATZA"}

	tests.Check(cart.SetCurrentOrder(cmdtest.OrderName))
	tests.Check(cart.AddProducts(codes))
	for i, c := range codes {
		tests.StrEq(cart.CurrentOrder.Products[i].Code, c, "set wrong code")
	}
	tests.Check(cart.SaveAndReset())
	tests.Exp(cart.AddProducts(codes))

	o := dawg.Order{}
	bytes, err := r.DataBase.Get(data.OrderPrefix + cmdtest.OrderName)
	tests.Check(err)
	tests.Check(json.Unmarshal(bytes, &o))
	for i, c := range codes {
		tests.StrEq(o.Products[i].Code, c, "stored wrong code")
	}
	tests.Check(cart.PrintOrders(false))
}

func TestHelpers_Err(t *testing.T) {
	r, cart, o := setup(t)
	defer r.CleanUp()
	m, err := cart.finder.Store().Menu()
	tests.Check(err)
	if m == nil {
		t.Fatal("nil menu")
	}
	tests.Exp(addProducts(o, m, []string{"nope", "not a thing"}))
	tests.Check(addProducts(o, m, []string{"12SCREEN"}))
	tests.Exp(addToppingsToOrder(o, "nothere", []string{"K", "B"}))
	tests.Exp(addToppingsToOrder(o, "", []string{"K", "B"}))
	tests.Exp(addToppingsToOrder(o, "12SCREEN", []string{""}))
}

func setup(t *testing.T) (*cmdtest.Recorder, *Cart, *dawg.Order) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	cart := New(r)
	cart.SetOutput(ioutil.Discard)
	return r, cart, cmdtest.NewTestOrder()
}
