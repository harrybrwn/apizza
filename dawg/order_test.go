package dawg

import (
	"fmt"
	"testing"
)

func TestGetOrderPrice(t *testing.T) {
	var err error
	o := Order{}
	if _, err = getOrderPrice(o); err == nil {
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
		Address: &StreetAddr{
			StreetNum:  "1600",
			StreetName: "Pennsylvania Ave.",
			CityName:   "Washington",
			State:      "DC",
			Zipcode:    "20500",
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
	if err = order.PlaceOrder(); err == nil {
		t.Error("expected error")
	}

	order.StoreID = "" // should cause dominos to reject the order and send an error
	_, err = getOrderPrice(order)
	if err == nil {
		t.Error("Should have raised an error", "\n\b", err)
	}
}

func TestNewOrder(t *testing.T) {
	addr := &StreetAddr{
		StreetNum:  "1600",
		StreetName: "Pennsylvania Ave.",
		CityName:   "Washington",
		State:      "DC",
		Zipcode:    "20500",
		AddrType:   "House",
	}
	s, err := NearestStore(addr, "Carryout")
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("store is <nil>")
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
	o.SetName("test_order")
	if o.OrderName != o.Name() || o.OrderName != "test_order" {
		t.Error("incorrect order name")
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

func TestOrder_Err(t *testing.T) {
	// store, err := NearestStore(testAddress(), "Delivery")
	addr := testAddress()
	addr.Street = ""
	store, err := NearestStore(addr, "Delivery")
	if err != nil {
		t.Error(err)
	}
	o := store.NewOrder()

	p, err := store.GetProduct("2LDCOKE")
	if err != nil {
		t.Error(err)
	}
	o.AddProduct(p)

	price, err := o.Price()
	if err == nil {
		t.Error(err)
	}
	if price != -1.0 {
		t.Error("expected bad price")
	}
}

func TestRemoveProduct(t *testing.T) {
	s, err := NearestStore(testAddress(), "Carryout")
	if err != nil {
		t.Error(err)
	}
	order := s.NewOrder()
	menu, err := s.Menu()
	if err != nil {
		t.Error(err)
	}
	productCodes := []string{"2LDCOKE", "12SCREEN", "PSANSABC", "B2PCLAVA"}
	for _, code := range productCodes {
		p, err := menu.GetProduct(code)
		if err != nil {
			t.Error(err)
		}
		order.AddProduct(p)
	}

	err = order.RemoveProduct("12SCREEN")
	if err != nil {
		t.Error(err)
	}
	err = order.RemoveProduct("B2PCLAVA")
	if err != nil {
		t.Error(err)
	}

	for _, p := range order.Products {
		if p.Code == "12SCREEN" || p.Code == "B2PCLAVA" {
			t.Error("should have been removed")
		}
	}
	if err = order.RemoveProduct("nothere"); err == nil {
		t.Error("expected error")
	}
}
