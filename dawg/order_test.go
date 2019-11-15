package dawg

import (
	"fmt"
	"testing"
	"time"
)

func TestGetOrderPrice(t *testing.T) {
	o := Order{}
	if o.cli == nil {
		o.cli = orderClient
	}
	_, err := getOrderPrice(o)
	if err == nil {
		t.Error("Should have returned an error")
	}
	if !IsFailure(err) {
		t.Error("this error should only be a failure")
		t.Error(err.Error())
	}

	order := Order{
		cli:          orderClient,
		LanguageCode: DefaultLang, ServiceMethod: "Delivery",
		StoreID: "4336", Payments: []*orderPayment{&orderPayment{}}, OrderID: "",
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
	if err := ValidateOrder(&order); IsFailure(err) {
		t.Error(err)
	}
	resp, err := getOrderPrice(order)
	if e, ok := err.(*DominosError); ok && IsFailure(err) {
		fmt.Printf("%+v\n", resp)
		t.Error("\n\b", e)
	}
	if len(order.Payments) == 0 {
		t.Fatal("order.Payments should be empty because tests were about to place an order")
	}
	if err = order.PlaceOrder(); err == nil {
		t.Error("expected error")
	}
	order.StoreID = "" // should cause dominos to reject the order and send an error
	_, err = getOrderPrice(order)
	if err == nil {
		t.Error("Should have raised an error", "\n\b", err)
	}

	err = order.prepare()
	if !IsFailure(err) {
		t.Error("Should have returned a dominos failure", err)
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

func TestPrepareOrder(t *testing.T) {
	st := testingStore()
	o := st.MakeOrder("Bob", "Smith", "bobsmith@aol.com")
	if o.FirstName != "Bob" {
		t.Error("wrong first name")
	}
	if o.LastName != "Smith" {
		t.Error("bad last name")
	}
	if o.Email != "bobsmith@aol.com" {
		t.Error("bad email")
	}
	if o.price > 0.0 {
		t.Error("order should not be initialized with a price above zero")
	}
	if len(o.OrderID) != 0 {
		t.Error("a new order should be initialized with an order id by default")
	}

	menu, err := st.Menu()
	if err != nil {
		t.Error(err)
	}
	if err = o.AddProduct(menu.FindItem("10SCREEN")); err != nil {
		t.Error(err)
	}

	if err = o.prepare(); err != nil {
		t.Error("Should not have returned error:\n", err)
	}
	if o.price <= 0.0 {
		t.Error("cached price should not be zero or less")
	}
	if len(o.OrderID) == 0 {
		t.Error("prepare should give the order an OrderID")
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

func TestOrderProduct(t *testing.T) {
	s := testingStore()
	menu, err := s.Menu()
	if err != nil {
		t.Error(err)
	}
	item := menu.FindItem("14SCEXTRAV")

	op := OrderProductFromItem(item)
	if err := op.AddTopping("X", "1/1", "1"); err != nil {
		t.Error(err)
	}

	if op.Price() != -1 {
		t.Error("bad price")
	}
	if op.Size() != -1 {
		t.Error("bad size")
	}
}

func TestCard(t *testing.T) {
	c := NewCard("1234123412341234", "01/10", 111)
	if c.Num() != "1234123412341234" {
		t.Error("go wrong card number")
	}

	tm := c.ExpiresOn()
	if tm.Month() != time.January {
		t.Error("wrong expiration month:", tm.Month())
	}
	if tm.Year() != 2010 {
		t.Error("bad expiration year:", tm.Year())
	}
	if c.Code() != "111" {
		t.Error("bad cvv")
	}
	if formatDate(tm) != "0110" {
		t.Error("bad date format:", formatDate(tm))
	}

	m, y := parseDate("01/10")
	if m != 1 {
		t.Error("parseDate failed to parse month")
	}
	if y != 2010 {
		t.Error("parseDate failed to parse year")
	}

	c = NewCard("", "", 0)
	if c != nil {
		t.Error("expected an nil value here")
	}

	ttDates := []string{"hello", "no/30", "2/no"}
	for _, tc := range ttDates {
		m, y = parseDate(tc)
		if m >= 0 || y >= 0 {
			t.Error("expected negative values")
		}
	}

	p, ok := NewCard("0000000000000000", "9/08", 123).(*Payment)
	if ok {
		tm = p.ExpiresOn()
		if tm.Year() != 2008 {
			t.Error("bad year:", tm.Year())
		}
		if tm.Month() != time.September {
			t.Error("bad month")
		}
	} else {
		t.Error("the default Card changed, go fix the tests")
	}
	p = ToPayment(NewCard("0000000000000000", "1/08", 123))
	if ok {
		p.Expiration = "08"
		tm = p.ExpiresOn()
		if tm != badExpiration {
			t.Error("a bad expiration date should have given the badExpiration variable")
		}
	} else {
		t.Error("the default Card changed, go fix the tests")
	}

	c = NewCard("0000000000000000", "08/08", 123)
	op := makeOrderPaymentFromCard(c)
	if op.Number != c.Num() {
		t.Error("bad number")
	}
	if op.Expiration != formatDate(c.ExpiresOn()) {
		t.Error("bad expiration")
	}
	if op.SecurityCode != c.Code() {
		t.Error("bad cvv")
	}
}
