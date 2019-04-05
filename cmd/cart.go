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

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/dawg"
)

type cartCmd struct {
	*basecmd
	price      bool
	delete     bool
	verbose    bool
	add        []string
	removeProd string
}

func (c *cartCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if len(args) < 1 {
		return data.PrintOrders(db, c.Output(), c.verbose)
	} else if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}
	name := args[0]

	if c.delete {
		if err = db.Delete(data.OrderPrefix + name); err != nil {
			return err
		}
		c.Printf("%s successfully deleted.\n", name)
		return nil
	}

	var order *dawg.Order
	if order, err = data.GetOrder(name, db); err != nil {
		return err
	}

	if len(c.removeProd) > 0 {
		if err = order.RemoveProduct(c.removeProd); err != nil {
			return err
		}
		return data.SaveOrder(order, c.Output(), db)
	}

	if len(c.add) > 0 {
		if err := db.UpdateTS("menu", c); err != nil {
			return err
		}
		for _, newP := range c.add {
			p, err := c.menu.GetVariant(newP)
			if err != nil {
				return err
			}
			order.AddProduct(p)
		}
		return data.SaveOrder(order, c.Output(), db)
	}

	return c.printOrder(name, order)
}

func (c *cartCmd) printOrder(name string, o *dawg.Order) (err error) {
	buffer := &bytes.Buffer{}
	out.SetOutput(buffer)

	if err := out.PrintOrder(o, true); err != nil {
		return err
	}
	fmt.Fprintf(buffer, "  address: %s\n", obj.AddressFmtIndent(o.Address, 11))

	if c.price {
		p, err := o.Price()
		if err != nil {
			return err
		}
		fmt.Fprintf(buffer, "  price: $%0.2f\n", p)
	}
	_, err = c.Output().Write(buffer.Bytes())
	return err
}

func (b *cliBuilder) newCartCmd() base.CliCommand {
	c := &cartCmd{price: false, delete: false, verbose: false}
	c.basecmd = b.newCommand("cart <order name>", "Manage user created orders", c)
	c.basecmd.Cmd().Long = `The cart command gets information on all of the user
created orders.`

	c.Flags().BoolVarP(&c.price, "price", "p", c.price, "show to price of an order")
	c.Flags().StringSliceVarP(&c.add, "add", "a", c.add, "add any number of products to a specific order")
	c.Flags().StringVarP(&c.removeProd, "remove-product", "r", c.removeProd, "remove a product from the order")
	c.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "delete the order from the database")
	c.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose, "print cart verbosly")
	return c
}

type addOrderCmd struct {
	*basecmd
	name     string
	products []string
	toppings []string
}

func (c *addOrderCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if c.name == "" && len(args) < 1 {
		return errors.New("No order name... use '--name=<order name>' or give name as an argument")
	}
	order := c.store().NewOrder()

	if c.name == "" {
		order.SetName(args[0])
	} else {
		order.SetName(c.name)
	}

	if len(c.products) > 0 {
		for i, p := range c.products {
			prod, err := c.store().GetVariant(p)
			if err != nil {
				return err
			}
			if i < len(c.toppings) {
				err = prod.AddTopping(c.toppings[i], dawg.ToppingFull, "1.0")
				if err != nil {
					return err
				}
			}
			order.AddProduct(prod)
		}
	} else if len(c.toppings) > 0 {
		return errors.New("cannot add just a toppings without products")
	}
	return data.SaveOrder(order, &bytes.Buffer{}, db)
}

func (b *cliBuilder) newAddOrderCmd() base.CliCommand {
	c := &addOrderCmd{name: "", products: []string{}}
	c.basecmd = b.newCommand("add <new order name>",
		"Create a new order that will be stored in the cart.", c)

	c.Flags().StringVarP(&c.name, "name", "n", c.name, "set the name of a new order")
	c.Flags().StringSliceVarP(&c.products, "products", "p", c.products, "product codes for the new order")
	c.Flags().StringSliceVarP(&c.toppings, "toppings", "t", c.toppings, "toppings for the products being added")
	return c
}
