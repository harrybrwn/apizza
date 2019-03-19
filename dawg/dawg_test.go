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
		t.Error("exepcted zero length responce")
	}
	if err == nil {
		t.Error("expected error")
	}

	_, err = post("invalid path", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestApizza(t *testing.T) {}
