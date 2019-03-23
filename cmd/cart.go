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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/dawg"
)

var orderPrefix = "user_order_"

type cartCmd struct {
	*basecmd
	price   bool
	delete  bool
	verbose bool
	add     []string
}

func (c *cartCmd) run(cmd *cobra.Command, args []string) (err error) {
	if len(args) < 1 {
		return c.printall()
	} else if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}

	if c.delete {
		if err = db.Delete(orderPrefix + args[0]); err != nil {
			return err
		}
		fmt.Fprintln(c.output, args[0], "successfully deleted.")
		return nil
	}

	order, err := getOrder(args[0])
	if err != nil {
		return err
	}

	if len(c.add) > 0 {
		if err := c.getstore(); err != nil {
			return err
		}
		for _, newP := range c.add {
			p, err := store.GetProduct(newP)
			if err != nil {
				return err
			}
			order.AddProduct(p)
		}
		if err := saveOrder(args[0], order); err != nil {
			return err
		}
		fmt.Fprintln(c.output, "updated order successfully saved.")
		return nil
	}

	return c.printOrder(args[0], order)
}

func (c *cartCmd) printall() error {
	all, err := db.GetAll()
	if err != nil {
		return err
	}
	if len(all) < 1 {
		fmt.Fprintln(c.output, "No orders saved.")
		return nil
	}

	fmt.Fprintln(c.output, "Your Orders:")
	for k := range all {
		if strings.Contains(k, orderPrefix) {
			fmt.Fprintln(c.output, " ", strings.Replace(k, orderPrefix, "", -1))
		}
	}
	return nil
}

func (c *cartCmd) printOrder(name string, o *dawg.Order) (err error) {
	buffer := &bytes.Buffer{}

	fmt.Fprintln(buffer, name)
	if c.price {
		p, err := o.Price()
		if err != nil {
			return err
		}
		fmt.Fprintf(buffer, "  Price: %f\n", p)
	}

	fmt.Fprintln(buffer, "  Products:")
	for _, p := range o.Products {
		fmt.Fprintf(buffer, "    %s", p.Code)
		if c.verbose {
			fmt.Fprintf(buffer, " - quantity: %d, options: %v\n", p.Qty, p.Options)
		} else {
			fmt.Fprint(buffer, "\n")
		}
	}
	fmt.Fprintf(buffer, "  StoreID: %s\n", o.StoreID)
	fmt.Fprintf(buffer, "  Method:  %s\n", o.ServiceMethod)
	fmt.Fprintf(buffer, "  Address: %s\n", addressStrIndent(o.Address, 11))
	if test {
		for _, prod := range o.Products {
			fmt.Fprintf(buffer, "%+v\n", prod)
		}
	}
	_, err = c.output.Write(buffer.Bytes())
	return err
}

func (b *cliBuilder) newCartCmd() cliCommand {
	c := &cartCmd{price: false, delete: false, verbose: false}
	c.basecmd = b.newBaseCommand("cart <order name>", "Manage user created orders", c.run)
	c.basecmd.cmd.Long = `The cart command gets information on all of the user
created orders. Use 'apizza cart <order name>' for info on a specific order`

	c.cmd.Flags().BoolVarP(&c.price, "price", "p", c.price, "show to price of an order")
	c.cmd.Flags().StringSliceVarP(&c.add, "add", "a", c.add, "add any number of products to a specific order")
	c.cmd.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "delete the order from the database")
	c.cmd.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose, "print cart verbosly")
	return c
}

type addOrderCmd struct {
	*basecmd
	name     string
	products []string
	toppings []string
}

func (c *addOrderCmd) run(cmd *cobra.Command, args []string) (err error) {
	if c.name == "" && len(args) < 1 {
		return errors.New("No order name... use '--name=<order name>' or give name as an argument")
	}
	var orderName string

	if c.name == "" {
		orderName = args[0]
	} else {
		orderName = c.name
	}

	if err := c.getstore(); err != nil {
		return err
	}
	order := store.NewOrder()

	if len(c.products) > 0 {
		for i, p := range c.products {
			prod, err := store.GetProduct(p)
			if err != nil {
				return err
			}
			if i < len(c.toppings) {
				prod.AddTopping(c.toppings[i], dawg.ToppingFull, 1.0)
			}
			order.AddProduct(prod)
		}
	} else if len(c.toppings) > 0 {
		return errors.New("cannot add just a toppings without products")
	}
	return saveOrder(orderName, order)
}

func (b *cliBuilder) newAddOrderCmd() cliCommand {
	c := &addOrderCmd{name: "", products: []string{}}
	c.basecmd = b.newBaseCommand(
		"add <new order name>",
		"Create a new order that will be stored in the cart.",
		c.run,
	)

	c.cmd.Flags().StringVarP(&c.name, "name", "n", c.name, "set the name of a new order")
	c.cmd.Flags().StringSliceVarP(&c.products, "products", "p", c.products, "product codes for the new order")
	c.cmd.Flags().StringSliceVarP(&c.toppings, "toppings", "t", c.toppings, "toppings for the products being added")
	return c
}

func getOrder(name string) (*dawg.Order, error) {
	raw, err := db.Get(orderPrefix + name)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("cannot find order %s", name)
	}
	order := &dawg.Order{}
	if err = json.Unmarshal(raw, order); err != nil {
		return nil, err
	}
	return order, nil
}

func saveOrder(name string, o *dawg.Order) error {
	raw, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return db.Put(orderPrefix+name, raw)
}
