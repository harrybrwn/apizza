// Copyright © 2019 Harrison Brown harrybrown98@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/cart"
	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
)

// `apizza cart`
type cartCmd struct {
	cli.CliCommand
	cart *cart.Cart

	validate bool
	price    bool
	delete   bool
	verbose  bool

	add     []string
	remove  string // yes, you can only remove one thing at a time
	product string

	topping bool // not actually a flag anymore
}

// TODO: changing a cart item needs to be more intuitive.

func (c *cartCmd) Run(cmd *cobra.Command, args []string) (err error) {
	out.SetOutput(cmd.OutOrStdout())
	c.cart.SetOutput(c.Output())
	if len(args) < 1 {
		return c.cart.PrintOrders(c.verbose)
	}

	if c.topping && c.product == "" {
		return errors.New("must specify an item code with '--product' to edit an order's toppings")
	} else if !c.topping && c.product != "" {
		c.topping = true
	}

	name := args[0]

	if c.delete {
		if err = c.cart.DeleteOrder(name); err != nil {
			return err
		}
		c.Printf("%s successfully deleted.\n", name)
		return nil
	}
	// Set the order that will be used will the cart functions
	if err = c.cart.SetCurrentOrder(args[0]); err != nil {
		return err
	}

	if c.validate {
		// validate the current order and stop
		return c.cart.Validate()
	}

	if len(c.remove) > 0 {
		if c.topping {
			for _, p := range c.cart.CurrentOrder.Products {
				if _, ok := p.Options()[c.remove]; ok || p.Code == c.product {
					delete(p.Opts, c.remove)
					break
				}
			}
		} else {
			if err = c.cart.CurrentOrder.RemoveProduct(c.remove); err != nil {
				return err
			}
		}
		return c.cart.Save()
	}

	if len(c.add) > 0 {
		if c.topping {
			err = c.cart.AddToppings(c.product, c.add)
		} else {
			err = c.cart.AddProducts(c.add)
		}
		if err != nil {
			return err
		}
		// stave order and return early before order is printed out
		return c.cart.SaveAndReset()
	}
	return out.PrintOrder(c.cart.CurrentOrder, true, c.price)
}

func onlyFailures(e error) error {
	if e == nil || dawg.IsWarning(e) {
		return nil
	}
	return e
}

// NewCartCmd creates a new cart command.
func NewCartCmd(b cli.Builder) cli.CliCommand {
	c := &cartCmd{
		cart:    cart.New(b),
		price:   false,
		delete:  false,
		verbose: false,
		topping: false,
	}

	c.CliCommand = b.Build("cart <order name>", "Manage user created orders", c)
	cmd := c.Cmd()

	cmd.Long = `The cart command gets information on all of the user
created orders.`

	cmd.PreRunE = cartPreRun()

	c.Flags().BoolVar(&c.validate, "validate", c.validate, "send an order to the dominos order-validation endpoint.")
	c.Flags().BoolVar(&c.price, "price", c.price, "show to price of an order")
	c.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "delete the order from the database")

	c.Flags().StringSliceVarP(&c.add, "add", "a", c.add, "add any number of products to a specific order")
	c.Flags().StringVarP(&c.remove, "remove", "r", c.remove, "remove a product from the order")
	c.Flags().StringVarP(&c.product, "product", "p", "", "give the product that will be effected by --add or --remove")

	c.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose, "print cart verbosely")

	c.Addcmd(newAddOrderCmd(b))
	return c
}

func cartPreRun() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("cannot handle multiple orders")
		}
		return nil
	}
}

// `apizza cart new` command
type addOrderCmd struct {
	cli.CliCommand
	client.StoreFinder
	db *cache.DataBase

	name     string
	product  string
	toppings []string
}

func (c *addOrderCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if c.name == "" && len(args) < 1 {
		return internal.ErrNoOrderName
	}
	order := c.Store().NewOrder()

	if c.name == "" {
		order.SetName(args[0])
	} else {
		order.SetName(c.name)
	}

	if c.product != "" {
		prod, err := c.Store().GetVariant(c.product)
		if err != nil {
			return err
		}
		for _, t := range c.toppings {
			err = prod.AddTopping(t, dawg.ToppingFull, "1.0")
			if err != nil {
				return err
			}
		}
		if err = order.AddProduct(prod); err != nil {
			return err
		}
	} else if len(c.toppings) > 0 {
		return errors.New("cannot add just a toppings without products")
	}
	return data.SaveOrder(order, &bytes.Buffer{}, c.db)
}

func newAddOrderCmd(b cli.Builder) cli.CliCommand {
	c := &addOrderCmd{name: "", product: ""}
	c.CliCommand = b.Build("new <order name>",
		"Create a new order that will be stored in the cart.", c)
	c.db = b.DB()
	c.StoreFinder = client.NewStoreGetter(b)

	c.Flags().StringVarP(&c.name, "name", "n", c.name, "set the name of a new order")
	c.Flags().StringVarP(&c.product, "product", "p", c.product, "product codes for the new order")
	c.Flags().StringSliceVarP(&c.toppings, "toppings", "t", c.toppings, "toppings for the products being added")
	return c
}

// `apizza order`
type orderCmd struct {
	cli.CliCommand
	db *cache.DataBase

	verbose bool
	track   bool

	email, phone string
	fname, lname string
	cvv          int
	number       string
	expiration   string

	logonly    bool
	getaddress func() dawg.Address
}

func (c *orderCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if len(args) < 1 {
		return data.PrintOrders(c.db, c.Output(), c.verbose)
	} else if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}

	if c.cvv == 0 {
		return errors.New("must have cvv number. (see --cvv)")
	}
	order, err := data.GetOrder(args[0], c.db)
	if err != nil {
		return err
	}

	order.AddCard(dawg.NewCard(
		eitherOr(c.number, config.GetString("card.number")),
		eitherOr(c.expiration, config.GetString("card.expiration")),
		c.cvv))

	names := strings.Split(config.GetString("name"), " ")
	if len(names) >= 1 {
		order.FirstName = eitherOr(c.fname, names[0])
	}
	if len(names) >= 2 {
		order.LastName = eitherOr(c.lname, strings.Join(names[1:], " "))
	}
	order.Email = eitherOr(c.email, config.GetString("email"))
	order.Phone = eitherOr(c.phone, config.GetString("phone"))
	order.Address = dawg.StreetAddrFromAddress(c.getaddress())

	c.Printf("Ordering dominos for %s to %s\n\n", order.ServiceMethod, strings.Replace(obj.AddressFmt(order.Address), "\n", " ", -1))

	if c.logonly {
		log.Println("logging order:", dawg.OrderToJSON(order))
		return nil
	}

	if !yesOrNo(os.Stdin, "Would you like to purchase this order? (y/n)") {
		return nil
	}

	c.Printf("sending order '%s'...\n", order.Name())
	// TODO: save the order id as a traced order and give it a timeout of
	// an hour or two.
	err = order.PlaceOrder()
	// logging happens after so any data from placeorder is included
	log.Println("sending order:", dawg.OrderToJSON(order))
	if err != nil {
		return err
	}
	c.Printf("sent to %s %s\n", order.Address.LineOne(), order.Address.City())

	if c.verbose {
		if order.ServiceMethod == dawg.Delivery {
			c.Printf("sent by %s to %s %s\n", order.ServiceMethod,
				order.Address.LineOne(), order.Address.City())
		} else {
			c.Printf("sent order for %s\n", order.ServiceMethod)
		}
		c.Printf("%+v\n", order)
	}
	return nil
}

func eitherOr(s1, s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	return s1
}

// NewOrderCmd creates a new order command.
func NewOrderCmd(b cli.Builder) cli.CliCommand {
	c := &orderCmd{
		verbose:    false,
		getaddress: b.Address,
	}
	c.CliCommand = b.Build("order", "Send an order from the cart to dominos.", c)
	c.db = b.DB()
	c.Cmd().Long = `The order command is the final destination for an order. This is where
the order will be populated with payment information and sent off to dominos.

The --cvv flag must be specified, and the config file will never store the
cvv. In addition to keeping the cvv safe, payment information will never be
stored the program cache with orders.
`
	c.Cmd().PreRunE = cartPreRun()

	flags := c.Cmd().Flags()
	flags.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "output the order command verbosely")

	flags.StringVar(&c.phone, "phone", "", "Set the phone number that will be used for this order")
	flags.StringVar(&c.email, "email", "", "Set the email that will be used for this order")
	flags.StringVar(&c.fname, "first-name", "", "Set the first name that will be used for this order")
	flags.StringVar(&c.fname, "last-name", "", "Set the last name that will be used for this order")

	flags.IntVar(&c.cvv, "cvv", 0, "Set the card's cvv number for this order")
	flags.StringVar(&c.number, "number", "", "the card number used for orderings")
	flags.StringVar(&c.expiration, "expiration", "", "the card's expiration date")

	flags.BoolVar(&c.logonly, "log-only", false, "")
	flags.MarkHidden("log-only")
	return c
}
