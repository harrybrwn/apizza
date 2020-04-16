<h1 align="center"><img alt="apizza" src="/docs/logo.png"></h1>

[![Build Status](https://travis-ci.com/harrybrwn/apizza.svg?branch=master)](https://travis-ci.com/harrybrwn/apizza)
[![GoDoc](https://godoc.org/github.com/github.com/harrybrwn/apizza/dawg?status.svg)](https://pkg.go.dev/github.com/harrybrwn/apizza/dawg?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrybrwn/apizza)](https://goreportcard.com/report/github.com/harrybrwn/apizza)
[![codecov](https://codecov.io/gh/harrybrwn/apizza/branch/master/graph/badge.svg)](https://codecov.io/gh/harrybrwn/apizza)
[![TODOs](https://badgen.net/https/api.tickgit.com/badgen/github.com/harrybrwn/apizza)](https://www.tickgit.com/browse?repo=github.com/harrybrwn/apizza)

Dominos pizza from the command line.

### Table of Contents
- [Installation](#installation)
- [Setup](#setup)
- [Commands](#commands)
	- [Config](#config)
	- [Menu](#menu)
	- [Cart](#cart)
	- [Order](#order)
	- [None Pizza with Left Beef](#none-pizza-with-left-beef)

### Installation
Download the precompiled binaries for Mac, Windows, and Linux (only for amd64)
```
wget https://github.com/harrybrwn/apizza/releases/download/v0.0.2/apizza-linux
wget https://github.com/harrybrwn/apizza/releases/download/v0.0.2/apizza-darwin
wget https://github.com/harrybrwn/apizza/releases/download/v0.0.2/apizza-windows
```

Or compile from source
```bash
go get -u github.com/harrybrwn/apizza
```

### Setup
The most you have to do as a user in terms of setting up apizza is fill in the config variables. The only config variables that are mandatory are "Address" and "Service" but the other config variables contain information that the Dominos website uses.https://github.com/harrybrwn/apizzahttps://github.com/harrybrwn/apizzahttps://github.com/harrybrwn/apizza

To edit the config file, you can either use the built-in `config get` and `config set` commands (see [Config](#config)) to configure apizza or you can edit the `$HOME/.config/apizza/config.json` file. Both of these setup methods will have the same results If you add a key-value pair to the `config.json` file that is not already in the file it will be overwritten the next time the program is run.


### Config
For documentation on configuration and configuration fields, see [documentation](/docs/configuration.md)

The `config get` and `config set` commands can be used with one config variable at a time...
```bash
$ apizza config set email='bob@example.com'
$ apizza config set name='Bob'
$ apizza config set service='Carryout'
```

Or they can be moved to one command like so.
```bash
$ apizza config set name=Bob email='bob@example.com' service='Carryout'
```

Or just edit the json config file with
```bash
$ apizza config --edit
```


### Menu
Run `apizza menu` to print the dominos menu.

The menu command will also give more detailed information when given arguments.

The arguments can either be a product code or a category name.
```bash
$ apizza menu pizza      # show all the pizza
$ apizza menu drinks     # show all the drinks
$ apizza menu 10SCEXTRAV # show details on 10SCEXTRAV
```
To see the different menu categories, use the `--show-categories` flag. And to view the different toppings use the `--toppings` flag.


### Cart
To save a new order, use `apizza cart new`
```bash
$ apizza cart new 'testorder' --product=16SCREEN --toppings=P,C,X # pepperoni, cheese, sauce
```
`apizza cart` is the command the shows all the saved orders.

> Note: Adding and removing items from the cart is a little bit weird and it will probably change in the future.

The two flags `--add` and `--remove` are intended for editing an order. They will not work if no order name is given as a command. To add a product from an order, simply give `apizza cart <order> --add=<product>` and to remove a product give `--remove=<product>`.

Editing a product's toppings a little more complicated. The `--product` flag is the key to editing toppings. To edit a topping, give the product that the topping belongs to to the `--product` flag and give the actual topping name to either `--remove` or `--add`.

```bash
$ apizza cart myorder --product=16SCREEN --add=P
```
This command will add pepperoni to the pizza named 16SCREEN, and...
```bash
$ apizza cart myorder --product=16SCREEN --remove=P
```
will remove pepperoni from the 16SCREEN item in the order named 'myorder'.


### Order
To actually send an order from the cart. Use the `order` command.

```bash
$ apizza order myorder --cvv=000
```
Once the command is executed, it will prompt you asking if you are sure you want to send the order. Enter `y` and the order will be sent.

### None Pizza with Left Beef
```bash
$ apizza cart new --name=leftbeef --product=12SCREEN
$ apizza cart leftbeef --remove=C --product=12SCREEN # remove cheese
$ apizza cart leftbeef --remove=X --product=12SCREEN # remove sauce
$ apizza cart leftbeef --add=B:left --product=12SCREEN # add beef to the left
```


### The [Dominos API Wrapper for Go](/docs/dawg.md)
Docs and example code for my Dominos library.

> **Credit**: Logo was made with [Logomakr](https://logomakr.com/).
