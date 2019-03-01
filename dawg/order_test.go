package dawg

import (
	"fmt"
	"testing"
)

func TestGetOrderPrice(t *testing.T) {
	o := Order{}
	_, err := getOrderPrice(o)
	if err == nil {
		t.Error("Should have returned an error")
	}
	order := Order{
		LanguageCode:  DefaultLang,
		ServiceMethod: "Delivery",
		StoreID:       "4336",
		Products: []*Product{
			&Product{
				Code: "12SCREEN",
				Options: map[string]interface{}{
					"C": map[string]string{"1/1": "1"},
					"P": map[string]string{"1/1": "1.5"},
				},
				Qty: 1,
			},
		},
		Payments: []Payment{Payment{}},
		OrderID:  "",
		Address: Address{
			StreetNum:  "1600",
			Street:     "1600 Pennsylvania Ave.",
			StreetName: "Pennsylvania Ave.",
			City:       "Washington",
			State:      "DC",
			Zip:        "20500",
			AddrType:   "House",
		},
	}
	resp, err := getOrderPrice(order)
	if e, ok := err.(*DominosError); ok && e.IsFailure() {
		fmt.Printf("%+v\n", resp)
		t.Error("\n\b", e)
	}
	if order.Payments == nil {
		t.Error("order.Payments should not be nil after getOrderPrice")
	}

	order.StoreID = ""
	_, err = getOrderPrice(order)
	if err != nil {
		t.Log(err.Error())
	}
	if err == nil {
		t.Error("Should have raised an error", "\n\b", err)
	}
}

func TestNewOrder(t *testing.T) {
	var addr = &Address{
		StreetNum:  "1600",
		StreetName: "Pennsylvania Ave NW",
		// Street: "1600 Pennsylvania Ave NW",
		City:     "Washington",
		State:    "DC",
		Zip:      "20500",
		AddrType: "House",
	}
	s, err := NearestStore(addr, "Delivery")
	if err != nil {
		t.Error(err)
	}
	_, err = s.GetProduct("S_PIZZA")
	if err == nil {
		t.Error("should have returned an error")
	}
	p, err := s.GetProduct("2LDCOKE")
	if err != nil {
		t.Error(err)
	}

	o := s.NewOrder()
	if o == nil {
		t.Error("NewOrder should not be nil")
	}
	o.AddProduct(p)
	if o.Products == nil {
		t.Error("Products should not be empty")
	}
	pizza, err := s.GetProduct("12SCREEN")
	pizza.AddTopping("X", ToppingFull, 1.5)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("adding a bad topping coverage should panic")
			}
		}()
		pizza.AddTopping("P", ToppingLeft, 1.2)
	}()
	if err != nil {
		t.Error(err)
	}
	if pizza == nil {
		t.Error("product is nil")
	}
	o.AddProduct(pizza)
	o.AddPayment(Payment{Number: "", Expiration: "", CVV: ""})
	price, err := o.Price()
	if err != nil {
		t.Error(err)
	}
	if price == -1.0 {
		t.Error("Order.Price() failed")
	}
}
