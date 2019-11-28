package dawg

import (
	"errors"
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
func ParseAddress(raw string) (*StreetAddr, error) {
	parsed, err := parse([]byte(raw))
	if err != nil {
		return nil, err
	}
	addr := &StreetAddr{
		StreetNum:  string(parsed[1]),
		Street:     string(parsed[1]) + " " + string(parsed[2]),
		StreetName: string(parsed[2]),
		CityName:   string(parsed[3]),
		State:      string(parsed[4]),
		Zipcode:    string(parsed[5]),
	}
	return addr, nil
}

func parse(raw []byte) ([][]byte, error) {
	res := addressRegex.FindAllSubmatch(raw, -1)
	if len(res) > 0 {
		return res[0], nil
	}
	return nil, errors.New("address parsing error")
}

// Address is a guid for how addresses should be used as input
type Address interface {
	LineOne() string
	StateCode() string
	City() string
	Zip() string
}

var _ Address = (*StreetAddr)(nil)

// StreetAddr represents a street address
type StreetAddr struct {
	// Street should be the street number followed by the street name.
	Street string `json:"Street"`

	// StreetNum is just the street number.
	StreetNum string `json:"StreetNumber"`

	// StreetName is just the street name.
	StreetName string `json:"StreetName"`

	CityName string `json:"City"`
	State    string `json:"Region"`
	Zipcode  string `json:"PostalCode"`

	// This is a dominos specific field, and should one of the following...
	// "House", "Apartment", "Business", "Campus/Base", "Hotel", or "Other"
	AddrType string `json:"Type"`
}

// StreetAddrFromAddress returns a StreetAddr pointer from an Address interface.
func StreetAddrFromAddress(addr Address) *StreetAddr {
	parts := strings.Split(addr.LineOne(), " ")
	var streetNum, streetName string

	if _, err := strconv.Atoi(parts[0]); err == nil {
		streetNum = parts[0]
	}
	streetName = strings.Join(parts[1:], " ")

	if res, ok := addr.(*StreetAddr); ok {
		if len(res.StreetNum) == 0 {
			res.StreetNum = streetNum
		}
		if len(res.StreetName) == 0 {
			res.StreetName = streetName
		}
		return res
	}

	return &StreetAddr{
		Street:     addr.LineOne(),
		StreetNum:  streetNum,
		CityName:   addr.City(),
		State:      addr.StateCode(),
		Zipcode:    addr.Zip(),
		StreetName: streetName,
	}
}

// LineOne gives the street in the following format
//
// <number> <name> <type>
// 123 Example St.
func (s *StreetAddr) LineOne() string {
	if s.StreetNum != "" && s.StreetName != "" {
		return fmt.Sprintf("%s %s", s.StreetNum, s.StreetName)
	}
	return s.Street
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
