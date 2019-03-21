package dawg

import (
	"fmt"
	"testing"
)

func TestFormat(t *testing.T) {
	url := format("https://order.dominos.com/power/%s", "store-locator")
	expected := "https://order.dominos.com/power/store-locator"
	if url != expected {
		t.Error(fmt.Sprintf("Expected: %s, Got: %s", expected, url))
	}
}

func TestAddressTable(t *testing.T) {
	var cases = []struct {
		raw      string
		expected StreetAddr
	}{
		{
			`1600 Pennsylvania Ave. Washington, DC 20500`,
			StreetAddr{StreetNum: "1600", StreetName: "Pennsylvania Ave.",
				StreetLineOne: "1600 Pennsylvania Ave.", CityName: "Washington",
				State: "DC", Zipcode: "20500", AddrType: "House"},
		},
		{
			`378 James St. Chicago, IL 60621`,
			StreetAddr{StreetNum: "378", StreetName: "James St.",
				StreetLineOne: "378 James St.", CityName: "Chicago", State: "IL",
				Zipcode: "60621"},
		},
	}

	for _, tc := range cases {
		addr := ParseAddress(tc.raw)
		if addr.StreetNum != tc.expected.StreetNum {
			t.Error("wrong street num")
		}
		if addr.StreetLineOne != tc.expected.StreetLineOne {
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
}

func TestDominosErrors(t *testing.T) {
	fmt.Println("try to find the fields that dominos gives when giving errors")
	order := &Order{
		LanguageCode:  "en",
		ServiceMethod: "Delivery",
		// ServiceMethod: "",
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
		StoreID: "4336",
		OrderID: "",
		Address: &StreetAddr{
			// StreetNum: "1600",
			StreetLineOne: "1600 Pennsylvania Ave.",
			// StreetName: "Pennsylvania Ave.",
			CityName: "Washington",
			State:    "DC",
			Zipcode:  "20500",
			AddrType: "House",
		},
	}
	resp, err := post("/power/price-order", order.rawData())
	if err != nil {
		t.Error(err)
	}
	if err := dominosErr(resp); err != nil {
		t.Error(err)
		for k, v := range err.fullErr {
			if k != "Order" {
				fmt.Println(k, v)
			}
		}
		print("\n")
		for k, v := range err.fullErr["Order"].(map[string]interface{}) {
			fmt.Println(k, v)
		}
	} else {
		fmt.Printf("we chillin\n%+v\n", order)
	}
	// fmt.Println(err)
	// e := &DominosError{}
	// if err := e.init(resp); err != nil {
	// 	panic(err)
	// }
	// if e.IsFailure() {
	// 	fmt.Println(e)
	// }
}
