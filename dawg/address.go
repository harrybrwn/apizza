package dawg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
func ParseAddress(raw string) *StreetAddr {
	parsed := parse([]byte(raw))
	return &StreetAddr{
		StreetNum:     string(parsed[1]),
		StreetLineOne: string(parsed[1]) + " " + string(parsed[2]),
		StreetName:    string(parsed[2]),
		CityName:      string(parsed[3]),
		State:         string(parsed[4]),
		Zipcode:       string(parsed[5]),
	}
}

func parse(raw []byte) [][]byte {
	return addressRegex.FindAllSubmatch(raw, -1)[0]
}

// Address represents a street address.
// type Address struct {
// 	Street     string `json:"Street"`
// 	StreetNum  string `json:"StreetNumber"`
// 	City       string `json:"City"`
// 	State      string `json:"Region"`
// 	Zip        string `json:"PostalCode"`
// 	AddrType   string `json:"Type"`
// 	StreetName string `json:"StreetName"`
// }

// Address is a guid for how addresses should be used as input
type Address interface {
	Street() string
	StateCode() string
	City() string
	Zip() string
}

var _ Address = (*StreetAddr)(nil)

// StreetAddr represents a street address
type StreetAddr struct {
	StreetLineOne string `json:"Street"`
	StreetNum     string `json:"StreetNumber"`
	CityName      string `json:"City"`
	State         string `json:"Region"`
	Zipcode       string `json:"PostalCode"`
	AddrType      string `json:"Type"`
	StreetName    string `json:"StreetName"`
}

// StreetAddrFromAddress returns a StreetAddr pointer from an Address interface.
func StreetAddrFromAddress(addr Address) *StreetAddr {
	parts := strings.Split(addr.Street(), " ")
	var streetNum, streetName string
	if _, err := strconv.Atoi(parts[0]); err == nil {
		streetNum = parts[0]
		streetName = strings.Join(parts[1:], " ")
	}

	return &StreetAddr{
		StreetLineOne: addr.Street(),
		StreetNum:     streetNum,
		CityName:      addr.City(),
		State:         addr.StateCode(),
		Zipcode:       addr.Zip(),
		StreetName:    streetName,
	}
}

// Street gives the street in the format of
//
// <number> <name> <type>
// 123 Example St.
func (s *StreetAddr) Street() string {
	if s.StreetNum != "" && s.StreetName != "" {
		return fmt.Sprintf("%s %s", s.StreetNum, s.StreetName)
	}
	return s.StreetLineOne
}

// Zip returns the zipcode of the address
func (s *StreetAddr) Zip() string {
	return s.Zipcode
}

// StateCode is the code for the state of the address
func (s *StreetAddr) StateCode() string {
	return s.State
}

// City returns the city of the address
func (s *StreetAddr) City() string {
	return s.CityName
}
