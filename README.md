# apizza

[![Build Status](https://travis-ci.com/harrybrwn/apizza.svg?branch=master)](https://travis-ci.com/harrybrwn/apizza)
[![GoDoc](https://godoc.org/github.com/github.com/harrybrwn/apizza/dawg?status.svg)](https://godoc.org/github.com/harrybrwn/apizza/dawg)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrybrwn/apizza)](https://goreportcard.com/report/github.com/harrybrwn/apizza)

Dominos pizza from the command line.

### Table of Contents
- [Installation](#installation)
- [Setup](#setup)
- [Commands](#commands)
	- [Config](#config)
	- [Cart](#cart)
	- [Menu](#menu)
- [DAWG](#the-dominos-api-wrapper-for-go)

### Installation
```
go get -u github.com/harrybrwn/apizza
go install github.com/harrybrwn/apizza
```

### Setup
The most you have to do as a user in terms of setting up apizza is fill in the
config variables. The only config variables that are manditory are "Address"
and "Service" but the other config variables contain information that the Dominos
website uses.

> **Note**: The config file won't exist if apizza is not run at least once.

To edit the config file, you can either use the built-in `config get` and
`config set` commands (see [Config](#config)) to configure apizza or you can edit the `config.json` file
in your home path. Both of these setup methods will have the same results If you
add a key-value pair to the `config.json` file that is not already in the file
it will be overwritten the next time the program is run.


### Config
The `config get` and `config set` commands can be used with one config variable
at a time...
```
apizza config set email=bob@example.com
apizza config set name=Bob
apizza config set service=Carryout
```

or they can be moved to one command like so. Make sure that there are no spaces between keys and values and that there is a space between key-value pairs.
```
apizza config set name=Bob email=bob@example.com service=Carryout
```


### Cart
To save a new order, use `apizza cart add` along with at least the `--name` flag or an argument representing an order name. The name flag is the name that the app will use when referring to that order. The `--products` flag takes at least one string but accepts a list of comma separated product codes that can be found in the menu command.

Viewing all of the saved orders is as simple as `apizza cart`.


### Menu
Run `apizza menu` to print the dominos menu.


### The Domios API Wrapper for Go
The DAWG library is the internal api wrapper used by apizza for interfacing with the dominos pizza api.
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
