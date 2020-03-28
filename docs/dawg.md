### The Dominos API Wrapper for Go
The DAWG library is the api wrapper used by apizza for interfacing with the dominos pizza api.

```go
package main

import (
	"fmt"
	"log"

	"github.com/harrybrwn/apizza/dawg"
)

var addr = &dawg.StreetAddr{
	Street: "1600 Pennsylvania Ave.",
	City: "Washington",
	State: "DC",
	Zip: "20500",
	AddrType: "House",
}

func main() {
	store, err := dawg.NearestStore(addr, "Delivery")
	if err != nil {
		log.Fatal(err)
	}
	order := store.NewOrder()

	pizza, err := store.GetProduct("16SCREEN")
	if err != nil {
		log.Fatal(err)
	}
	pizza.AddTopping("P", dawg.ToppingLeft, 1.5)
	order.AddProduct(pizza)

	if store.IsOpen {
		fmt.Println(order.Price())
	} else {
		fmt.Println("dominos is not open")
	}
}
```