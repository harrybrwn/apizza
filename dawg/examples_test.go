package dawg_test

import (
	"fmt"
	"log"

	"github.com/harrybrwn/apizza/dawg"
)

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

func ExampleOrder_AddProduct() {
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
	order := store.NewOrder()

	pizza, err := store.GetProduct("16SCREEN")
	if err != nil {
		log.Fatal(err)
	}
	pizza.AddTopping("P", dawg.ToppingFull, "1.5")
	order.AddProduct(pizza)
}

func ExampleProduct_AddTopping() {
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
	order := store.NewOrder()

	pizza, err := store.GetProduct("16SCREEN")
	if err != nil {
		log.Fatal(err)
	}
	pizza.AddTopping("P", dawg.ToppingLeft, "1.0")  // pepperoni on the left
	pizza.AddTopping("K", dawg.ToppingRight, "2.0") // double bacon on the right
	order.AddProduct(pizza)
}
