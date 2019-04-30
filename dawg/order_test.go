package dawg

import (
	"fmt"
	"testing"
)

func TestGetOrderPrice(t *testing.T) {
	o := Order{}
	if _, err := getOrderPrice(o); err == nil {
		t.Error("Should have returned an error")
	}
	order := Order{
		LanguageCode: DefaultLang, ServiceMethod: "Delivery",
		StoreID: "4336", Payments: []Payment{Payment{}}, OrderID: "",
		Products: []*OrderProduct{
			&OrderProduct{
				item: item{
					Code: "12SCREEN",
				},
				Opts: map[string]interface{}{
					"C": map[string]string{"1/1": "1"},
					"P": map[string]string{"1/1": "1.5"},
				},
				Qty: 1,
			},
		},
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
	if e, ok := err.(*DominosError); ok && IsFailure(err) {
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
	s := testingStore()
	if _, err := s.GetProduct("S_PIZZA"); err != nil {
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
	v, err := s.GetVariant("2LDCOKE")
	if err != nil {
		t.Fatal(err)
	}
	err = o.AddProduct(v)
	if err != nil {
		t.Error(err)
	}
	if o.Products == nil {
		t.Error("Products should not be empty")
	}
	pizza, err := s.GetVariant("14TMEATZA")
	if err != nil {
		t.Error(err)
	}
	if pizza == nil {
		t.Error("product is nil")
	}
	pizza.AddTopping("X", ToppingFull, "1.5")
	err = o.AddProduct(pizza)
	if err != nil {
		t.Error(err)
	}
	price, err := o.Price()
	if IsFailure(err) {
		t.Error(err)
	}
	if price == -1.0 {
		t.Error("Order.Price() failed")
	}
	if err != nil {
		t.Error(err)
	}
}

func TestOrder_Err(t *testing.T) {
	addr := testAddress()
	addr.Street = ""
	store, err := NearestStore(addr, "Delivery")
	if err != nil {
		t.Error(err)
	}
	o := store.NewOrder()
	v, err := store.GetVariant("2LDCOKE")
	if err != nil {
		t.Error(err)
	}
	if v == nil {
		t.Fatal("got nil variant")
	}
	err = o.AddProduct(v)
	if err != nil {
		t.Error(err)
	}
	price, err := o.Price()
	if err == nil {
		t.Error(err)
	}
	if price != -1.0 {
		t.Error("expected bad price")
	}
}

func TestRemoveProduct(t *testing.T) {
	s := testingStore()
	order := s.NewOrder()
	menu, err := s.Menu()
	if err != nil {
		t.Error(err)
	}
	productCodes := []string{"2LDCOKE", "12SCREEN", "PSANSABC", "B2PCLAVA"}
	for _, code := range productCodes {
		v, err := menu.GetVariant(code)
		if err != nil {
			t.Error(err)
		}
		order.AddProduct(v)
	}
	if err = order.RemoveProduct("12SCREEN"); err != nil {
		t.Error(err)
	}
	if err = order.RemoveProduct("B2PCLAVA"); err != nil {
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
