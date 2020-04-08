package obj

import (
	"testing"

	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestAddressStr(t *testing.T) {
	tests.InitHelpers(t)
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
	tests.StrEq(addr.LineOne(), a.LineOne(), "wrong lineone")
	tests.StrEq(addr.StateCode(), a.StateCode(), "wrong state code")
	tests.StrEq(addr.City(), a.City(), "wrong city")
	tests.StrEq(addr.Zip(), a.Zip(), "wrong zip")
	if AddrIsEmpty(addr) {
		t.Error("should not be empty")
	}
	addr = &Address{}
	if !AddrIsEmpty(addr) {
		t.Error("addr should be empty")
	}
}
