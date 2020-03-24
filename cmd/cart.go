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
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
)

// `apizza cart`
type cartCmd struct {
	cli.CliCommand
	data.MenuCacher
	client.StoreFinder
	db *cache.DataBase

	validate bool

	price   bool
	delete  bool
	verbose bool

	add     []string
	remove  string // yes, you can only remove one thing at a time
	product string

	topping bool // not actually a flag anymore
}

func (c *cartCmd) Run(cmd *cobra.Command, args []string) (err error) {
	out.SetOutput(cmd.OutOrStdout())
	if len(args) < 1 {
		return data.PrintOrders(c.db, c.Output(), c.verbose)
	}

	if c.topping && c.product == "" {
		return errors.New("must specify an item code with '--product' to edit an order's toppings")
	} else if !c.topping && c.product != "" {
		c.topping = true
	}

	name := args[0]

	if c.delete {
		if err = c.db.Delete(data.OrderPrefix + name); err != nil {
			return err
		}
		c.Printf("%s successfully deleted.\n", name)
		return nil
	}

	var order *dawg.Order
	if order, err = data.GetOrder(name, c.db); err != nil {
		return err
	}

	if c.validate {
		c.Printf("validating order '%s'...\n", order.Name())
		return onlyFailures(order.Validate())
	}

	if len(c.remove) > 0 {
		if c.topping {
			for _, p := range order.Products {
				if _, ok := p.Options()[c.remove]; ok || p.Code == c.product {
					delete(p.Opts, c.remove)
					break
				}
			}
		} else {
			if err = order.RemoveProduct(c.remove); err != nil {
				return err
			}
		}
		return data.SaveOrder(order, c.Output(), c.db)
	}

	if len(c.add) > 0 {
		if c.topping {
			if c.product == "" {
				return errors.New("what product are these toppings being added to")
			}
			for _, top := range c.add {
				p := getOrderItem(order, c.product)
				if p == nil {
					return fmt.Errorf("cannot find '%s' in the '%s' order", c.product, order.Name())
				}

				err = addTopping(top, p)
				if err != nil {
					return err
				}
			}
		} else {
			if err := c.db.UpdateTS("menu", c); err != nil {
				return err
			}
			menu := c.Menu()
			var itm dawg.Item
			for _, newP := range c.add {
				itm, err = menu.GetVariant(newP)
				if err != nil {
					return err
				}
				err = order.AddProduct(itm)
				if err != nil {
					return err
				}
			}
		}
		return data.SaveOrder(order, c.Output(), c.db)
	}
	return out.PrintOrder(order, true, c.price)
}

func (c *cartCmd) syncWithConfig(o *dawg.Order) error {
	addr := config.Get("address").(obj.Address)
	if obj.AddrIsEmpty(&addr) {
		return errs.New("no address in config file")
	}

	o.Address = dawg.StreetAddrFromAddress(&addr)
	o.StoreID = c.Store().ID
	return onlyFailures(data.SaveOrder(o, c.Output(), c.db))
}

func onlyFailures(e error) error {
	if e == nil || dawg.IsWarning(e) {
		return nil
	}
	return e
}

// adds a topping.
//
// formated as <name>:<side>:<amount>
// name is the only one that is required.
func addTopping(topStr string, p dawg.Item) error {
	var side, amount string

	topping := strings.Split(topStr, ":")

	if len(topping) < 1 {
		return errors.New("incorrect topping format")
	}

	if len(topping) == 1 {
		side = dawg.ToppingFull
	} else if len(topping) >= 2 {
		side = topping[1]
	}

	if len(topping) == 3 {
		amount = topping[2]
	} else {
		amount = "1.0"
	}
	p.AddTopping(topping[0], side, amount)
	return nil
}

func getOrderItem(order *dawg.Order, code string) dawg.Item {
	for _, itm := range order.Products {
		if itm.ItemCode() == code {
			return itm
		}
	}
	return nil
}

// NewCartCmd creates a new cart command.
func NewCartCmd(b cli.Builder) cli.CliCommand {
	c := &cartCmd{
		db:      b.DB(),
		price:   false,
		delete:  false,
		verbose: false,
		topping: false,
	}

	if app, ok := b.(*App); ok {
		c.StoreFinder = app
	} else {
		c.StoreFinder = client.NewStoreGetterFunc(
			func() string { return b.Config().Service }, b.Address)
	}

	c.MenuCacher = data.NewMenuCacher(menuUpdateTime, b.DB(), c.Store)
	c.CliCommand = b.Build("cart <order name>", "Manage user created orders", c)
	c.Cmd().Long = `The cart command gets information on all of the user
created orders.`

	c.Cmd().PersistentPreRunE = c.persistentPreRunE
	// c.Cmd().PreRunE = c.preRun

	c.Flags().BoolVar(&c.validate, "validate", c.validate, "send an order to the dominos order-validation endpoint.")

	c.Flags().BoolVar(&c.price, "price", c.price, "show to price of an order")
	c.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "delete the order from the database")

	c.Flags().StringSliceVarP(&c.add, "add", "a", c.add, "add any number of products to a specific order")
	c.Flags().StringVarP(&c.remove, "remove", "r", c.remove, "remove a product from the order")
	c.Flags().StringVarP(&c.product, "product", "p", "", "give the product that will be effected by --add or --remove")

	c.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose, "print cart verbosly")

	c.Addcmd(newAddOrderCmd(b))
	return c
}

func (c *cartCmd) persistentPreRunE(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}
	return nil
}

func (c *cartCmd) preRun(cmd *cobra.Command, args []string) error {
	return nil
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
		return errors.New("No order name... use '--name=<order name>' or give name as an argument")
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
	c.CliCommand = b.Build("new <new order name>",
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
	flags := c.Cmd().Flags()
	flags.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "output the order command verbosly")

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
