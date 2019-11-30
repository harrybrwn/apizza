package obj

import (
	"testing"

	"github.com/harrybrwn/apizza/dawg"
)

func TestAddressStr(t *testing.T) {
	a := &Address{
		Street: "1600 Pennsylvania Ave NW", CityName: "Washington",
		State: "dc", Zipcode: "20500",
	}
	expected := []string{
		`1600 Pennsylvania Ave NW
Washington, DC 20500`,
		`1600 Pennsylvania Ave NW
   Washington, DC 20500`,
	}
	formatted := []string{AddressFmt(a), AddressFmtIndent(a, 3)}
	for i, exp := range expected {
		if exp != formatted[i] {
			t.Errorf("unexpected output...\ngot:\n%s\nwanted:\n%s\n", formatted[i], exp)
		}
	}

	a = &Address{
		Street: "1600 Pennsylvania Ave NW", CityName: "Washington DC",
		State: "", Zipcode: "20500",
	}
	expected = []string{
		`1600 Pennsylvania Ave NW
Washington DC, 20500`,
		`1600 Pennsylvania Ave NW
   Washington DC, 20500`,
	}
	formatted = []string{AddressFmt(a), AddressFmtIndent(a, 3)}
	for i, exp := range expected {
		if exp != formatted[i] {
			t.Errorf("unexpected output...\ngot:\n%s\nwanted:\n%s\n", formatted[i], exp)
		}
	}
	addr := FromAddress(dawg.StreetAddrFromAddress(a))
	if addr.LineOne() != a.LineOne() {
		t.Error("wrong lineone")
	}
	if addr.StateCode() != a.StateCode() {
		t.Error("wrong state code")
	}
	if addr.City() != a.City() {
		t.Error("wrong city")
	}
	if addr.Zip() != a.Zip() {
		t.Error("wrong zip")
	}

	if AddrIsEmpty(addr) {
		t.Error("should not be empty")
	}
	addr = &Address{}
	if !AddrIsEmpty(addr) {
		t.Error("addr should be empty")
	}
}
