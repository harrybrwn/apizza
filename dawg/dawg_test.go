package dawg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"testing"
)

func TestFormat(t *testing.T) {
	url := format("https://order.dominos.com/power/%s", "store-locator")
	expected := "https://order.dominos.com/power/store-locator"
	if url != expected {
		t.Errorf("Expected: %s, Got: %s", expected, url)
	}
}

func TestOrderAddressConvertion(t *testing.T) {
	exp := &StreetAddr{StreetNum: "1600", StreetName: "Pennsylvania Ave.",
		Street: "1600 Pennsylvania Ave.", CityName: "Washington",
		State: "DC", Zipcode: "20500", AddrType: "House"}

	addr := &UserAddress{
		Street:     "1600 Pennsylvania Ave.",
		CityName:   "Washington",
		PostalCode: "20500",
		Region:     "DC",
	}

	res := StreetAddrFromAddress(addr)
	if res.City() != exp.City() {
		t.Error("wrong city")
	}
	if res.LineOne() != exp.LineOne() {
		t.Error("wrong lineone")
	}
	if res.StateCode() != exp.StateCode() {
		t.Error("wrong state code")
	}
	if res.Zip() != exp.Zip() {
		t.Error("wrong zip code")
	}
	if res.StreetNum != exp.StreetNum {
		t.Error("wrong street number")
	}
	if res.StreetName != exp.StreetName {
		t.Error("wrong street name")
	}
}

func TestParseAddressTable(t *testing.T) {
	var cases = []struct {
		raw      string
		expected StreetAddr
	}{
		{
			raw: `1600 Pennsylvania Ave. Washington, DC 20500`,
			expected: StreetAddr{StreetNum: "1600", StreetName: "Pennsylvania Ave.",
				Street: "1600 Pennsylvania Ave.", CityName: "Washington",
				State: "DC", Zipcode: "20500", AddrType: "House"},
		},
		{
			raw: `378 James St. Chicago, IL 60621`,
			expected: StreetAddr{StreetNum: "378", StreetName: "James St.",
				Street: "378 James St.", CityName: "Chicago", State: "IL",
				Zipcode: "60621"},
		},
	}

	for _, tc := range cases {
		addr, err := ParseAddress(tc.raw)
		if err != nil {
			t.Error(err)
		}
		if addr.StreetNum != tc.expected.StreetNum {
			t.Error("wrong street num")
		}
		if addr.Street != tc.expected.Street {
			t.Error("wrong street")
		}
		if addr.CityName != tc.expected.CityName {
			t.Error("wrong city")
		}
		if addr.State != tc.expected.State {
			t.Error("wrong state")
		}
		if addr.Zipcode != tc.expected.Zipcode {
			t.Error("wrong zip")
		}
	}
}

func TestNetworking_Err(t *testing.T) {
	_, err := orderClient.get("/", nil)
	if err == nil {
		t.Error("expected error")
	}
	_, err = orderClient.get("/invalid path", nil)
	if err == nil {
		t.Error("expected error")
	}
	b, err := orderClient.post("/invalid path", nil, bytes.NewReader(make([]byte, 1)))
	if len(b) != 0 {
		t.Error("exepcted zero length response")
	}
	if err == nil {
		t.Error("expected error")
	}
	_, err = orderClient.post("invalid path", nil, bytes.NewReader(nil))
	if err == nil {
		t.Error("expected error")
	}
	_, err = orderClient.post("/power/price-order", nil, bytes.NewReader([]byte{}))
	if err == nil {
		t.Error("expected error")
	}
	cli := &client{
		Client: &http.Client{
			Transport: &http.Transport{
				DialTLS: func(string, string) (net.Conn, error) {
					return nil, errors.New("stop")
				},
			},
		},
	}
	resp, err := cli.get("/power/store/4336/profile", nil)
	if err == nil {
		t.Error("expected error")
	}
	if resp != nil {
		t.Error("should not have gotten any response data")
	}
	b, err = cli.post("/invalid path", nil, bytes.NewReader(make([]byte, 1)))
	if err == nil {
		t.Error("expected error")
	}
	if b != nil {
		t.Error("exepcted zero length response")
	}
	req, err := http.NewRequest("GET", "https://www.google.com/", nil)
	if err != nil {
		t.Error(err)
	}
	resp, err = orderClient.do(req)
	if err == nil {
		t.Error("expected an error because we found an html page")
		fmt.Println(string(resp))
	} else if err.Error() != "got html response" {
		t.Error("got an unexpected error:", err.Error())
	}

	req, err = http.NewRequest("GET", "https://hjfkghfdjkhgfjkdhgjkdghfdjk.com", nil)
	if err != nil {
		t.Error(err)
	}
	resp, err = orderClient.do(req)
	if err == nil {
		t.Error("expected an error")
	}
}

func TestDominosErrors(t *testing.T) {
	order := &Order{
		LanguageCode:  "en",
		ServiceMethod: "Delivery",
		Products: []*OrderProduct{
			&OrderProduct{
				item: item{Code: "12SCREEN"},
				Opts: map[string]interface{}{
					"C": map[string]string{"1/1": "1"},
					"P": map[string]string{"1/1": "1.5"},
				},
				Qty: 1,
			},
		},
		StoreID: "4336",
		OrderID: "",
		Address: testAddress(),
	}
	resp, err := orderClient.post("/power/price-order", nil, order.raw())
	if err != nil {
		t.Error(err)
	}
	if err := dominosErr(resp); err != nil && IsFailure(err) {
		t.Error(err)
	}
}

func TestDominosErrorInit(t *testing.T) {
	err := dominosErr([]byte("bad data"))
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf("got wrong error type: %T\n", err)
	}
}

func TestDominosErrorFailure(t *testing.T) {
	e := dominosErr([]byte(`
{
	"Status":-1,
	"StatusItems": [{"Code":"Failure","Message":"test msg"}],
	"Order": {"Status": -1,
		"StatusItems": [
			{"Code":"Failure","Message":"test order msg"},
			{"Code":"SomeOtherCode"},
			{"PulseCode": 1, "PulseText": "this isn't the real error format"}
		]}}`))
	if e == nil {
		t.Error("dominos error should not be nil")
	}
	expected := `Dominos Failure (-1)
    Failure Code: 'Failure':
        test order msg
    Failure Code: 'SomeOtherCode'
        PulseCode 1: this isn't the real error format`
	if e.Error() != expected {
		t.Errorf("\nexpected:\n'%s'\ngot:\n'%s'\n", expected, e.Error())
	}
	if len(e.Error()) < 5 {
		t.Error("the error message here seems too small:\n", e.Error())
	}
	dErr := e.(*DominosError)
	if IsOk(dErr) {
		t.Error("no... its not ok!")
	}
	if IsWarning(dErr) {
		t.Error("error is not a warning")
	}
	if !IsFailure(dErr) {
		t.Error("should be a failure")
	}
	if IsWarning(nil) {
		t.Error("nil should not be interpreted as an error")
	}
	if !IsOk(nil) {
		t.Error("IsOk(nil) should be true because a nil error is ok")
	}
	if IsFailure(nil) {
		t.Error("nil is not a failure")
	}
}

func TestErrPair(t *testing.T) {
	tt := []struct {
		err error
		exp string
	}{
		{errpair(errors.New("one"), errors.New("two")), "error 1. one\nerror 2. two"},
		{errpair(errors.New("one"), nil), "one"},
		{errpair(nil, errors.New("two")), "two"},
	}
	for i, tc := range tt {
		if tc.err.Error() != tc.exp {
			t.Errorf("test case %d for errpair gave wrong result", i)
		}
	}
	err := errpair(nil, nil)
	if err != nil {
		t.Error("a pair of nil errors should result in one nil error")
	}
}

func testingStore() *Store {
	var service string

	if rand.Intn(2) == 1 {
		service = "Carryout"
	} else {
		service = "Delivery"
	}

	s, err := NearestStore(testAddress(), service)
	if err != nil {
		panic(err)
	}
	return s
}

func testingMenu() *Menu {
	m, err := testingStore().Menu()
	if err != nil {
		panic(err)
	}
	return m
}
