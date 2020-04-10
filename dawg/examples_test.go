package dawg_test

import (
	"fmt"
	"log"
	"os"

	"github.com/harrybrwn/apizza/dawg"
)

var (
	user     = dawg.UserProfile{}
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

func Example_getStore() {
	// This can be anything that satisfies the dawg.Address interface.
	var addr = dawg.StreetAddr{
		Street:   "600 Mountain Ave bldg 5",
		CityName: "New Providence",
		State:    "NJ",
		Zipcode:  "07974",
		AddrType: "Business",
	}
	store, err := dawg.NearestStore(&addr, dawg.Delivery)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(store.WaitTime())
}

func ExampleNearestStore() {
	var addr = &dawg.StreetAddr{
		Street:   "1600 Pennsylvania Ave.",
		CityName: "Washington",
		State:    "DC",
		Zipcode:  "20500",
		AddrType: "House",
	}
	store, err := dawg.NearestStore(addr, "Delivery")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(store.City)
	fmt.Println(store.ID)

	// Output:
	// Washington
	// 4336
}

func ExampleSignIn() {
	user, err := dawg.SignIn(username, password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user.Email == username)
	fmt.Printf("%T\n", user)
}

func ExampleUserProfile() {
	// Obtain a dawg.UserProfile with the dawg.SignIn function
	user, err := dawg.SignIn(username, password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user.Email == username)
	fmt.Printf("%T\n", user)
}

func ExampleUserProfile_GetCards() {
	cards, err := user.GetCards()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Test Card name:", cards[0].NickName) // This is dependant on the account
}

func ExampleUserProfile_AddAddress() {
	// The address that is stored in a dawg.UserProfile are the address that dominos
	// keeps track of and an address may need to be added.
	fmt.Printf("%+v\n", user.Addresses)

	user.AddAddress(&dawg.StreetAddr{
		Street:   "600 Mountain Ave bldg 5",
		CityName: "New Providence",
		State:    "NJ",
		Zipcode:  "07974",
		AddrType: "Business",
	})
	fmt.Println(len(user.Addresses) > 0)
	fmt.Printf("%T\n", user.DefaultAddress())

	// Output:
	// []
	// true
	// *dawg.UserAddress
}

func ExampleOrder_AddProduct() {
	store, err := dawg.NearestStore(&address, "Delivery")
	if err != nil {
		log.Fatal(err)
	}
	order := store.NewOrder()

	pizza, err := store.GetVariant("14SCREEN")
	if err != nil {
		log.Fatal(err)
	}
	pizza.AddTopping("P", dawg.ToppingFull, "1.5")
	order.AddProduct(pizza)

	fmt.Println(order.Products[0].Name)

	// Output:
	// Large (14") Hand Tossed Pizza
}

func ExampleProduct_AddTopping() {
	store, err := dawg.NearestStore(&address, "Delivery")
	if err != nil {
		log.Fatal(err)
	}
	order := store.NewOrder()

	pizza, err := store.GetVariant("14SCREEN")
	if err != nil {
		log.Fatal(err)
	}
	pizza.AddTopping("P", dawg.ToppingLeft, "1.0")  // pepperoni on the left
	pizza.AddTopping("K", dawg.ToppingRight, "2.0") // double bacon on the right
	order.AddProduct(pizza)

	fmt.Println(pizza.Options()["P"])
	fmt.Println(pizza.Options()["K"])

	// Output:
	// map[1/2:1.0]
	// map[2/2:2.0]
}
