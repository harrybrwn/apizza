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

func TestAddress(t *testing.T) {
	var testAddr = Address{
		StreetNum:  "1600",
		StreetName: "Pennsylvania Ave.",
		Street:     "1600 Pennsylvania Ave.",
		City:       "Washington",
		State:      "DC",
		Zip:        "20500",
		AddrType:   "House",
	}
	rawAddr := `1600 Pennsylvania Ave. Washington, DC 20500`

	addr := ParseAddress(rawAddr)
	if addr.StreetNum != testAddr.StreetNum {
		t.Error("wrong street num")
	}
	if addr.Street != testAddr.Street {
		t.Error("wrong street")
	}
	if addr.City != testAddr.City {
		t.Error("wrong city")
	}
	if addr.State != testAddr.State {
		t.Error("wrong state")
	}
	if addr.Zip != testAddr.Zip {
		t.Error("wrong zip")
	}

	parsed := parse([]byte(rawAddr))
	if string(parsed[1]) != testAddr.StreetNum {
		t.Error("wrong street num")
	}
	if string(parsed[1])+" "+string(parsed[2]) != testAddr.Street {
		t.Error("wrong street")
	}
	if string(parsed[3]) != testAddr.City {
		t.Error("wrong city")
	}
	if string(parsed[4]) != testAddr.State {
		t.Error("wrong state")
	}
	if string(parsed[5]) != testAddr.Zip {
		t.Error("wrong zip")
	}
}

func TestApizza(t *testing.T) {}
