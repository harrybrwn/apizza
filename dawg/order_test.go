package dawg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestGetOrderPrice(t *testing.T) {
	defer swapclient(1)()
	o := Order{cli: orderClient}
	_, err := getPricingData(o)
	if err == nil {
		t.Error("should have returned an error")
	}
	if !IsFailure(err) {
		t.Error("this error should only be a failure")
		t.Error(err.Error())
	}

	order := Order{
		cli:          orderClient,
		LanguageCode: DefaultLang, ServiceMethod: "Delivery",
		StoreID: "4336", OrderID: "",
		Payments: []*orderPayment{},
		Products: []*OrderProduct{
			{
				ItemCommon: ItemCommon{
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
	if err := order.Validate(); IsFailure(err) {
		t.Error(err)
	}
	resp, err := getOrderPrice(order)
	if e, ok := err.(*DominosError); ok && IsFailure(err) {
		fmt.Printf("%+v\n", resp)
		t.Error("\n\b", e)
	}
	if len(order.Payments) != 0 {
		t.Fatal("order.Payments should be empty because tests were about to place an order")
	}
	order.StoreID = "" // should cause dominos to reject the order and send an error
	_, err = getOrderPrice(order)
	if err == nil {
		t.Error("Should have raised an error", err)
	}

	err = order.prepare()
	if !IsFailure(err) {
		t.Error("Should have returned a dominos failure", err)
	}
}

func TestNewOrder(t *testing.T) {
	tests.InitHelpers(t)
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
	tests.Check(err)
	tests.Check(o.AddProductQty(v, 2))
	if o.Products == nil {
		t.Error("Products should not be empty")
	}
	pizza, err := s.GetVariant("14TMEATZA")
	tests.Check(err)
	if pizza == nil {
		t.Error("product is nil")
	}
	pizza.AddTopping("X", ToppingFull, "1.5")
	tests.Check(o.AddProduct(pizza))
	price, err := o.Price()
	if IsFailure(err) {
		t.Error(err)
	}
	if price == -1.0 {
		t.Error("Order.Price() failed")
	}
	tests.Check(err)
}

func TestPrepareOrder(t *testing.T) {
	tests.InitHelpers(t)
	st := testingStore()
	o := st.MakeOrder("Bob", "Smith", "bobsmith@aol.com")
	tests.StrEq(o.FirstName, "Bob", "wrong first name")
	tests.StrEq(o.LastName, "Smith", "wrong last name")
	tests.StrEq(o.Email, "bobsmith@aol.com", "wrong email")
	if o.price > 0.0 {
		t.Error("order should not be initialized with a price above zero")
	}
	if len(o.OrderID) != 0 {
		t.Error("a new order should be initialized with an order id by default")
	}

	menu := testingMenu()
	tests.Check(o.AddProduct(menu.FindItem("10SCREEN")))
	tests.Check(o.prepare())
	if o.price <= 0.0 {
		t.Error("cached price should not be zero or less")
	}
	if len(o.OrderID) == 0 {
		t.Error("prepare should give the order an OrderID")
	}
	length := len(o.Payments)
	o.AddPayment(Payment{})
	if len(o.Payments) != length+1 {
		t.Error("failed to add a payment")
	}
	o.AddCard(&Payment{})
	if len(o.Payments) != length+2 {
		t.Error("failed to add a card")
	}
}

func TestOrder_Err(t *testing.T) {
	tests.InitHelpers(t)
	addr := testAddress()
	addr.Street = ""
	store, err := NearestStore(addr, Delivery)
	tests.Check(err)
	o := store.NewOrder()
	v, err := store.GetVariant("2LDCOKE")
	tests.Check(err)
	if v == nil {
		t.Fatal("got nil variant")
	}
	tests.Check(o.AddProduct(v))
	price, err := o.Price()
	tests.Exp(err)
	if price != -1.0 {
		t.Error("expected bad price")
	}
	tests.Exp(o.AddProduct(nil))
	tests.Exp(o.AddProductQty(nil, 50))
	o = new(Order)
	InitOrder(o)
	tests.Exp(o.PlaceOrder())
	itm, err := store.GetVariant("12SCREEN")
	tests.Check(err)
	op := OrderProductFromItem(itm)
	tests.Exp(op.AddTopping("test", "test", "test"))
}

func TestRawOrder(t *testing.T) {
	tests.InitHelpers(t)
	var (
		err   error
		o     *Order
		reset = func() {
			o = &Order{
				ServiceMethod: Delivery,
				Address:       &StreetAddr{},
				Email:         "jake@statefarm.com",
				Phone:         "1234567",
			}
		}
	)
	reset()
	err = o.PlaceOrder()
	if !IsFailure(err) {
		t.Error("placing an empty order should fail")
	}
	reset()
	tests.Exp(o.Validate(), "expected validation error from empty order")
}

func TestRemoveProduct(t *testing.T) {
	tests.InitHelpers(t)
	s := testingStore()
	order := s.NewOrder()
	menu := testingMenu()

	productCodes := []string{"2LDCOKE", "12SCREEN", "PSANSABC", "B2PCLAVA"}
	for _, code := range productCodes {
		v, err := menu.GetVariant(code)
		tests.Check(err)
		tests.Check(order.AddProduct(v))
	}
	tests.Check(order.RemoveProduct("12SCREEN"))
	tests.Check(order.RemoveProduct("B2PCLAVA"))
	for _, p := range order.Products {
		if p.Code == "12SCREEN" || p.Code == "B2PCLAVA" {
			t.Error("should have been removed")
		}
	}
	tests.Exp(order.RemoveProduct("nothere"))
}

func TestOrderProduct(t *testing.T) {
	tests.InitHelpers(t)
	menu := testingMenu() // this will get the menu from the same store but cached
	item := menu.FindItem("14SCEXTRAV")

	op := OrderProductFromItem(item)
	tests.Check(op.AddTopping("X", "1/1", "1"))

	m := op.ReadableOptions()
	if len(m) <= 0 {
		t.Error("should have readable options")
	}

	op.menu = menu
	m = op.ReadableOptions()
	if len(m) <= 0 {
		t.Error("should have readable options")
	}
}

func TestCard(t *testing.T) {
	tests.InitHelpers(t)
	c := NewCard("1234123412341234", "01/10", 111)
	tests.StrEq(c.Num(), "1234123412341234", "go wrong card number")

	tm := c.ExpiresOn()
	if tm.Month() != time.January {
		t.Error("wrong expiration month:", tm.Month())
	}
	if tm.Year() != 2010 {
		t.Error("bad expiration year:", tm.Year())
	}
	tests.StrEq(c.Code(), "111", "bad cvv")
	tests.StrEq(formatDate(tm), "0110", "bad date format: %s", formatDate(tm))

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
	tests.StrEq(op.Number, c.Num(), "bad number")
	tests.StrEq(op.Expiration, formatDate(c.ExpiresOn()), "bad expiration")
	tests.StrEq(op.SecurityCode, c.Code(), "bad cvv")
}

func TestOrderToJSON(t *testing.T) {
	o := new(Order)
	s := OrderToJSON(o)
	if len(s) == 0 {
		t.Error("OrderToJSON should return something")
	}

	data, err := json.Marshal(o)
	if err != nil {
		t.Error(err)
	}
	plain := string(data)
	if len(s) <= len(plain) {
		t.Error("OrderToJSON should return an indented order")
	}

	s = o.raw().String()
	if !strings.HasPrefix(s, "{\"Order\":") {
		t.Error("Order.raw should be prefixed by '{\"Order\":'")
	}
	if !strings.HasSuffix(s, "}") {
		t.Error("Order.raw should end with '}'")
	}
}

func TestOrderCalls(t *testing.T) {
	o := new(Order)
	o.Init()
	err := sendOrder("/power/validate-order", *o)
	if !IsFailure(err) || err == nil {
		t.Error("expected error")
	}

	o = new(Order)
	InitOrder(o)
	err = sendOrder("", *o)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCardTypeRegex(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/cardnums.json")
	if err != nil {
		t.Error(err)
	}
	data := make(map[string][]string)
	if err = json.Unmarshal(b, &data); err != nil {
		t.Error(err)
	}
	for ctype, nums := range data {
		for _, num := range nums {
			pat, ok := cardRegex[ctype]
			if !ok {
				continue
			}
			match := pat.MatchString(num)
			if !match {
				t.Errorf("expected %s to match as %s", num, ctype)
			}
			if findCardType(num) != ctype {
				t.Error("wrong card type")
			}
		}
	}
}
