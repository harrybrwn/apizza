package dawg_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/harrybrwn/apizza/dawg"
)

func TestURLParams(t *testing.T) {
	expected := []string{
		"a=what&b=7&c=false",
		"b=7&c=false&a=what",
		"c=false&a=what&b=7",
	}
	p := dawg.Params{"a": "what", "b": 7, "c": false}
	enc := p.Encode()

	func() {
		for _, expec := range expected {
			// exit the closure if one of the possible encodings is found
			if enc == expec {
				return
			}
		}
		t.Error("bad url encoding")
	}()

	p = dawg.Params{"byteobj": []byte("data")}
	if p.Encode() != "byteobj=data" {
		t.Error("bad encoding for bytes")
	}

	t.Run("bad Param type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic")
			}
		}()
		type test struct{ a string }
		p = dawg.Params{"struct": test{"no"}}
		p.Encode()
	})

	p = nil
	enc = p.Encode()
	if enc != "" {
		t.Error("should be empty string")
	}
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
	pizza.AddTopping("P", dawg.ToppingFull, 1.5)
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
	pizza.AddTopping("P", dawg.ToppingLeft, 1.0)  // pepperoni on the left
	pizza.AddTopping("K", dawg.ToppingRight, 2.0) // double bacon on the right
	order.AddProduct(pizza)
}
