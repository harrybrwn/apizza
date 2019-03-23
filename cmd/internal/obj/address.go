package obj

import (
	"fmt"
	"strings"

	"github.com/harrybrwn/apizza/dawg"
)

// Address represents a street address
type Address struct {
	Street   string `config:"street"`
	CityName string `config:"cityname"`
	State    string `config:"state"`
	Zipcode  string `config:"zipcode"`
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
	panic(fmt.Sprintf("bad statecode %s", a.State))
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
	panic(fmt.Sprintf("bad zipcode %s", a.Zipcode))
}

func addressStr(a dawg.Address) string {
	return addressStrIndent(a, 0)
}

func addressStrIndent(a dawg.Address, tablen int) string {
	var format string
	if len(a.StateCode()) == 0 {
		format = "%s\n%s%s, %s%s"
	} else {
		format = "%s\n%s%s, %s %s"
	}

	return fmt.Sprintf(format,
		a.LineOne(), strings.Repeat(" ", tablen), a.City(), a.StateCode(), a.Zip())
}

func (a Address) String() string {
	return addressStr(&a)
}
