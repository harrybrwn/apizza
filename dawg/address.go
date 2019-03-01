package dawg

import (
	"fmt"
	"regexp"
)

var (
	addressRegex = regexp.MustCompile(
		fmt.Sprintf(`^(?P<%s>[0-9]{1,4})\s(?P<%s>[A-Za-z ]+\.|\n).?\n?\s*(?P<%s>[A-Za-z ]+),\s(?P<%s>[A-Z]{2})\s(?P<%s>[0-9]{5}).*$`,
			"street_num", "street", "city", "state", "zipcode"),
	)
)

// ParseAddress will parse a raw address and return an address object.
// This method is prone to all the mishaps that arise when trying to work
// with addresses, it may not work for all cases.
func ParseAddress(raw string) *Address {
	parsed := parse([]byte(raw))
	return &Address{
		StreetNum:  string(parsed[1]),
		Street:     string(parsed[1]) + " " + string(parsed[2]),
		StreetName: string(parsed[2]),
		City:       string(parsed[3]),
		State:      string(parsed[4]),
		Zip:        string(parsed[5]),
	}
}

func parse(raw []byte) [][]byte {
	return addressRegex.FindAllSubmatch(raw, -1)[0]
}

// Address represents a street address.
type Address struct {
	Street     string `json:"Street"`
	StreetNum  string `json:"StreetNumber"`
	City       string `json:"City"`
	State      string `json:"Region"`
	Zip        string `json:"PostalCode"`
	AddrType   string `json:"Type"`
	StreetName string `json:"StreetName"`
}
