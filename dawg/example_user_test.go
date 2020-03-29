package dawg_test

import (
	"log"
	"os"

	"github.com/harrybrwn/apizza/dawg"
)

var (
	username = os.Getenv("DOMINOS_TEST_USER")
	password = os.Getenv("DOMINOS_TEST_PASS")

	address = dawg.StreetAddr{
		Street:   "600 Mountain Ave bldg 5",
		CityName: "New Providence",
		State:    "NJ",
		Zipcode:  "07974",
		AddrType: "Business",
	}
)

func Example_dominosAccount() {
	user, err := dawg.SignIn(username, password)
	if err != nil {
		log.Fatal(err)
	}
	user.AddAddress(&address)
	store, err := user.NearestStore(dawg.Carryout)
	if err != nil {
		log.Fatal(err)
	}
	store.Menu()

	// Output:
}
