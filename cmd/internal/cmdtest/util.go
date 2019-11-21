package cmdtest

import (
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
)

// ObjAddressFromInterface translates a dawg.Address interface into a obj.Address.
func ObjAddressFromInterface(a dawg.Address) *obj.Address {
	return &obj.Address{
		Street:   a.LineOne(),
		CityName: a.City(),
		State:    a.StateCode(),
		Zipcode:  a.Zip(),
	}
}

// TestAddress returns a testing address
func TestAddress() *obj.Address {
	return &obj.Address{
		Street:   "1600 Pennsylvania Ave NW",
		CityName: "Washington",
		State:    "DC",
		Zipcode:  "20500",
	}
}
