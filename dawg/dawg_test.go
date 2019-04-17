package dawg

import (
	"encoding/json"
	"errors"
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

func TestAddressTable(t *testing.T) {
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
		addr := ParseAddress(tc.raw)
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
	_, err := get("/", nil)
	if err == nil {
		t.Error("expected error")
	}
	_, err = get("/invalid path", nil)
	if err == nil {
		t.Error("expected error")
	}
	b, err := post("/invalid path", make([]byte, 1))
	if len(b) != 0 {
		t.Error("exepcted zero length response")
	}
	if err == nil {
		t.Error("expected error")
	}
	_, err = post("invalid path", nil)
	if err == nil {
		t.Error("expected error")
	}
	_, err = post("/power/price-order", []byte{})
	if err == nil {
		t.Error("expected error")
	}
	original := cli
	cli = &http.Client{
		Transport: &http.Transport{
			DialTLS: func(string, string) (net.Conn, error) {
				return nil, errors.New("stop")
			},
		},
	}
	resp, err := get("/power/store/4336/profile", nil)
	if err == nil {
		t.Error("expected error")
	}
	if resp != nil {
		t.Error("should not have gotten any responce data")
	}
	b, err = post("/invalid path", make([]byte, 1))
	if err == nil {
		t.Error("expected error")
	}
	if b != nil {
		t.Error("exepcted zero length response")
	}
	cli = original
}

func TestDominosErrors(t *testing.T) {
	// fmt.Println("try to find the fields that dominos gives when giving errors")
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
	resp, err := post("/power/price-order", order.rawData())
	if err != nil {
		t.Error(err)
	}
	if err := dominosErr(resp); err != nil && IsFailure(err) {
		t.Error(err)
		// err, _ := err.(*DominosError)
		// for k, v := range err.fullErr {
		// 	if k != "Order" {
		// 		fmt.Println(k, v)
		// 	}
		// }
		// print("\n")
		// for k, v := range err.fullErr["Order"].(map[string]interface{}) {
		// 	fmt.Println(k, v)
		// }
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
	expected := `Dominos Failure:
    Code: 'Failure':
        test order msg
    Code: 'SomeOtherCode'
    PulseCode 1:
        this isn't the real error format`
	if e.Error() != expected {
		t.Errorf("\nexpected:\n'%s'\ngot:\n'%s'\n", expected, e.Error())
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
