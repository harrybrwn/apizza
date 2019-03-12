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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"apizza/dawg"
)

type orderCommand struct {
	*basecmd
	showAll   bool
	showPrice bool
	delete    bool
}

func (c *orderCommand) run(cmd *cobra.Command, args []string) (err error) {
	if c.showAll {
		all, err := db.GetAll()
		if err != nil {
			return err
		}
		for k := range all {
			if strings.Contains(k, "user_order_") {
				fmt.Println(getOrderName(k))
			}
		}
		return nil
	}

	if len(args) < 1 {
		return errors.New("Error: no orders were given")
	}

	if c.delete {
		if err = db.Delete("user_order_" + args[0]); err != nil {
			return err
		}
		fmt.Println(args[0], "successfully deleted.")
		return nil
	}

	order, err := getOrder(args[0])
	if err != nil {
		return err
	}

	if c.showPrice {
		price, err := order.Price()
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", args[0])
		fmt.Printf("  Price: %f\n", price)
	}

	return nil
}

func newOrderCommand() cliCommand {
	c := &orderCommand{showAll: false, showPrice: false, delete: false}
	c.basecmd = newBaseCommand("order [order name]", "Order pizza from dominos", c.run)

	c.cmd.Flags().BoolVarP(&c.showAll, "show-all", "a", c.showAll, "show the previously cached and saved orders")
	c.cmd.Flags().BoolVarP(&c.showPrice, "show-price", "p", c.showPrice, "show to price of an order")
	c.cmd.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "delete the order from the database")
	return c
}

type newOrderCmd struct {
	*basecmd
	name, product string
}

func (c *newOrderCmd) run(cmd *cobra.Command, args []string) (err error) {
	if store == nil {
		store, err = dawg.NearestStore(c.addr, cfg.Service)
		if err != nil {
			return err
		}
	}
	order := store.NewOrder()

	if c.name == "" {
		return errors.New("Error: No order name... use '--name=<order name>'")
	}

	if c.product != "" {
		p, err := store.GetProduct(c.product)
		if err != nil {
			return err
		}
		order.AddProduct(p)
	}

	raw, err := json.Marshal(&order)
	if err != nil {
		return err
	}
	fmt.Print(c.name, ": ")
	fmt.Println(string(raw))

	price, err := order.Price()
	if err != nil {
		return err
	}
	fmt.Println("Price:", price)
	err = db.Put("user_order_"+c.name, raw)
	return nil
}

func (b *cliBuilder) newNewOrderCmd() cliCommand {
	c := &newOrderCmd{name: "", product: ""}
	c.basecmd = b.newBaseCommand(
		"new",
		"Create a new order that will be stored in the cache.",
		c.run,
	)

	c.cmd.Flags().StringVarP(&c.name, "name", "n", c.name, "set the name of a new order")
	c.cmd.Flags().StringVarP(&c.product, "product", "p", c.product, "the product code for the new order")
	return c
}

func getOrderName(name string) string {
	return strings.Replace(name, "user_order_", "", -1)
}

func getOrder(name string) (*dawg.Order, error) {
	raw, err := db.Get("user_order_" + name)
	if err != nil {
		return nil, err
	}
	order := &dawg.Order{}
	if err = json.Unmarshal(raw, order); err != nil {
		return nil, err
	}
	return order, nil
}
