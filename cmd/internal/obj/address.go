package obj

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/harrybrwn/apizza/dawg"
)

var _ dawg.Address = (*Address)(nil)

// Address represents a street address
type Address struct {
	Street   string `config:"street" json:"street"`
	CityName string `config:"cityname" json:"cityname"`
	State    string `config:"state" json:"state"`
	Zipcode  string `config:"zipcode" json:"zipcode"`
}

// FromAddress makes an obj.Address from an address interface.
func FromAddress(a dawg.Address) *Address {
	return &Address{
		Street:   a.LineOne(),
		CityName: a.City(),
		State:    a.StateCode(),
		Zipcode:  a.Zip(),
	}
}

// LineOne returns the first line of the address
func (a *Address) LineOne() string {
	return a.Street
}

// StateCode returns the state code of the address
func (a *Address) StateCode() string {
	if len(a.State) == 2 {
		return strings.ToUpper(a.State)
	} else if len(a.State) == 0 {
		return ""
	}
	return ""
}

// City returns the city
func (a *Address) City() string {
	return a.CityName
}

// Zip returns the zip code.
func (a *Address) Zip() string {
	if strings.Contains(a.Zipcode, " ") {
		panic(fmt.Sprintf("bad zipcode %s", a.Zipcode))
	}
	if len(a.Zipcode) == 5 {
		return a.Zipcode
	}
	return ""
}

// AddressFmt returns a formatted address string from and Address interface.
func AddressFmt(a dawg.Address) string {
	return AddressFmtIndent(a, 0)
}

// AddressFmtIndent returns AddressFmt but with the second line indented to some
// length l.
func AddressFmtIndent(a dawg.Address, l int) string {
	var format string
	if len(a.StateCode()) == 0 {
		format = "%s\n%s%s, %s%s"
	} else {
		format = "%s\n%s%s, %s %s"
	}

	return fmt.Sprintf(format,
		a.LineOne(),
		strings.Repeat(" ", l),
		a.City(),
		a.StateCode(),
		a.Zip(),
	)
}

func (a Address) String() string {
	return AddressFmt(&a)
}

// AsGob converts the Address into a binary format using the gob package.
func AsGob(a *Address) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(a)
	return buf.Bytes(), err
}

// FromGob will create a new address from binary data encoded using the gob package.
func FromGob(raw []byte) (*Address, error) {
	a := &Address{}
	return a, gob.NewDecoder(bytes.NewReader(raw)).Decode(a)
}

// AsJSON converts the Address to json format.
func AsJSON(a *Address) ([]byte, error) {
	return json.Marshal(a)
}

// AddrIsEmpty will tell if an address is empty.
func AddrIsEmpty(a dawg.Address) bool {
	if a.LineOne() == "" &&
		a.Zip() == "" &&
		a.City() == "" &&
		a.StateCode() == "" {
		return true
	}
	return false
}
