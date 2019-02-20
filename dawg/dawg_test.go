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

func TestURLParams(t *testing.T) {
	expected := []string{
		"a=what&b=7&c=false",
		"b=7&c=false&a=what",
		"c=false&a=what&b=7",
	}
	p := Params{"a": "what", "b": 7, "c": false}
	enc := p.Encode()

	func() {
		for _, expec := range expected {
			// exit the closure if one of the possible encodings is found
			if enc == expec {
				return
			}
		}
		t.Error("bad url encoding")
	}()

	p = Params{"byteobj": []byte("data")}
	if p.Encode() != "byteobj=data" {
		t.Error("bad encoding for bytes")
	}

	t.Run("bad Param type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic")
			}
		}()
		type test struct{ a string }
		p = Params{"struct": test{"no"}}
		p.Encode()
	})

	p = nil
	enc = p.Encode()
	if enc != "" {
		t.Error("should be empty string")
	}
}

func TestAddressTable(t *testing.T) {
	var cases = []struct {
		raw      string
		expected Address
	}{
		{
			`1600 Pennsylvania Ave. Washington, DC 20500`,
			Address{StreetNum: "1600", StreetName: "Pennsylvania Ave.",
				Street: "1600 Pennsylvania Ave.", City: "Washington",
				State: "DC", Zip: "20500", AddrType: "House"},
		},
		{
			`378 James St. Chicago, IL 60621`,
			Address{StreetNum: "378", StreetName: "James St.",
				Street: "378 James St.", City: "Chicago", State: "IL",
				Zip: "60621"},
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
		if addr.City != tc.expected.City {
			t.Error("wrong city")
		}
		if addr.State != tc.expected.State {
			t.Error("wrong state")
		}
		if addr.Zip != tc.expected.Zip {
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

	b, err = post("invalid path", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestApizza(t *testing.T) {}
