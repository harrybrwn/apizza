# apizza

[![Build Status](https://travis-ci.com/harrybrwn/apizza.svg?branch=master)](https://travis-ci.com/harrybrwn/apizza)
[![GoDoc](https://godoc.org/github.com/github.com/harrybrwn/apizza?status.svg)](https://godoc.org/github.com/harrybrwn/apizza)
[![Coverage Status](https://coveralls.io/repos/github/harrybrwn/apizza/badge.svg?branch=master)](https://coveralls.io/github/harrybrwn/apizza?branch=master)

A cli for ordering domios pizza.

### Table of Contents
- [Installatoin](#installation)
- [Setup](#setup)
- [DAWG](#the-dominos-api-wrapper-for-go)

### Installation
```
go install -u github.com/harrybrwn/apizza
```

### Setup
You can either use the built-in `get` and `set` commands to configure apizza or you can edit the `config.json` file in your home path both methods will have the same results. If you add a key-value pair to the `config.json` file that is not already in the file it will be overwritten the next time the program is run.

The `get` and `set` comands can be used one at a time,
```
apizza config set email=bob@example.com
apizza config set name=Bob
apizza config set service=Carryout
```

or they can be moved to one line like so. Make sure that there are no spaces between keys and values and that there is a space between key-value pairs.
```
apizza config set name=Bob email=bob@example.com service=Carryout
```

### The Domios API Wrapper for Go
The DAWG library is the api wrapper used by apizza for interfacing with the dominos pizza api.
```go
package main

import (
	"fmt"
	"log"

	"github.com/harrybrwn/apizza/dawg"
)

var addr = &dawg.Address{
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
	order, err := store.NewOrder()
	if err != nil {
		log.Fatal(err)
	}

	pizza, err := store.GetProduct("16SCREEN")
	if err != nil {
		log.Fatal(err)
	}
	pizza.AddTopping("P", dawg.ToppingLeft, 1.5)
	order.AddProduct(pizza)

	if store.IsOpen {
		fmt.Println(order.Price()
	} else {
		fmt.Println("dominos is not open")
	}
}
```
