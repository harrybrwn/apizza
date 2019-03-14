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
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"apizza/dawg"
)

type menuCmd struct {
	*basecmd
	menu          *dawg.Menu
	all           bool
	toppings      bool
	preconfigured bool
	item          string
	output        io.Writer
}

func (c *menuCmd) run(cmd *cobra.Command, args []string) (err error) {
	if store == nil {
		if store, err = dawg.NearestStore(c.addr, cfg.Service); err != nil {
			return err
		}
	}

	if err := c.initMenu(); err != nil {
		return err
	}

	if len(c.item) > 0 {
		prod, err := c.findProduct(c.item)
		if err != nil {
			return err
		}
		iteminfo(prod, c.output)
		return nil
	}

	if c.toppings {
		c.printToppings()
		return nil
	}
	c.printMenu()
	return nil
}

func (b *cliBuilder) newMenuCmd() cliCommand {
	c := &menuCmd{output: os.Stdout, all: false, toppings: false, preconfigured: false}
	c.basecmd = b.newBaseCommand("menu", "Get the Dominos menu.", c.run)

	c.cmd.Flags().BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	c.cmd.Flags().BoolVarP(&c.toppings, "toppings", "t", c.toppings, "print out the toppings on the menu")
	c.cmd.Flags().BoolVarP(&c.preconfigured, "preconfigured",
		"p", c.preconfigured, "show the pre-configured products on the dominos menu")
	c.cmd.Flags().StringVarP(&c.item, "item", "i", "", "show info on the menu item given")
	return c
}

func (c *menuCmd) initMenu() error {
	cachedMenu, err := db.Get("menu")
	if err != nil {
		return err
	}
	cacheTime, err := db.TimeStamp("menu")
	if err != nil {
		return err
	}

	if cachedMenu != nil && time.Since(cacheTime) < 12*time.Hour {
		c.menu = &dawg.Menu{}
		if err = json.Unmarshal(cachedMenu, c.menu); err != nil {
			return err
		}
	} else {
		if err = db.ResetTimeStamp("menu"); err != nil {
			return err
		}
		if err = c.cacheNewMenu(); err != nil {
			return err
		}
	}
	return nil
}

func (c *menuCmd) printMenu() {
	var f func(map[string]interface{}, string)

	f = func(m map[string]interface{}, spacer string) {
		cats := m["Categories"].([]interface{})
		prods := m["Products"].([]interface{})

		// if there is nothing in that category, dont print the code name
		if len(cats) != 0 || len(prods) != 0 {
			fmt.Fprint(c.output, spacer, m["Code"], "\n")
		}
		if len(cats) > 0 { // the recursive part
			for _, c := range cats {
				f(c.(map[string]interface{}), spacer+"  ")
			}
		} else if len(prods) > 0 { // the printing part
			for _, p := range prods {
				product, err := c.findProduct(p.(string))
				if err != nil {
					continue
				}
				c.printMenuItem(product, spacer)
			}
			fmt.Fprint(c.output, "\n")
		}
	}
	keys := []string{"Food"}
	if c.preconfigured {
		keys = []string{"PreconfiguredProducts"}
	}
	if c.all {
		keys = []string{"PreconfiguredProducts", "Food"}
	}

	for _, key := range keys {
		f(c.menu.Categorization[key].(map[string]interface{}), "")
	}
}

func (c *menuCmd) printMenuItem(product map[string]interface{}, spacer string) {
	// if product has varients, print them
	if varients, ok := product["Variants"].([]interface{}); ok {
		fmt.Fprintf(c.output, "%s  \"%s\" [%s]\n", spacer, product["Name"], product["Code"])
		max := maxStrLen(varients)

		for _, v := range varients {
			fmt.Fprintln(
				c.output, spaces(strLen(spacer)+3), "-", v, spaces(max-strLen(v.(string))), c.menu.Variants[v.(string)].(map[string]interface{})["Name"])
		}
	} else {
		// if product has no varients, it is a preconfigured product
		fmt.Fprintf(c.output, "%s  \"%s\"\n%s - %s", spacer, product["Name"], spacer+"    ", product["Code"])
	}
}

func iteminfo(prod map[string]interface{}, output io.Writer) {
	fmt.Fprintf(output, "%s [%s]\n", prod["Name"], prod["Code"])
	fmt.Fprintf(output, "    %s: %v\n", "DefaultToppings", prod["Tags"].(map[string]interface{})["DefaultToppings"])

	for k, v := range prod {
		if k != "Name" && k != "Tags" && k != "ImageCode" {
			fmt.Fprintf(output, "    %s: %v\n", k, v)
		}
	}
}

func (c *menuCmd) printToppings() {
	indent := strings.Repeat(" ", 4)
	for key, val := range c.menu.Toppings {
		fmt.Fprint(c.output, "  ", key, "\n")
		for k, v := range val.(map[string]interface{}) {
			spacer := strings.Repeat(" ", 3-strLen(k))
			fmt.Fprintln(
				c.output, indent, k, spacer, v.(map[string]interface{})["Name"])
		}
		fmt.Fprint(c.output, "\n")
	}
}

func (c *menuCmd) findProduct(key string) (map[string]interface{}, error) {
	var product map[string]interface{}
	if prod, ok := c.menu.Products[key]; ok {
		product = prod.(map[string]interface{})
	} else if prod, ok := c.menu.Variants[key]; ok {
		product = prod.(map[string]interface{})
	} else if prod, ok := c.menu.Preconfigured[key]; ok {
		product = prod.(map[string]interface{})
	} else {
		return nil, fmt.Errorf("could not find %s", key)
	}
	return product, nil
}

func (c *menuCmd) cacheNewMenu() (err error) {
	if store == nil {
		store, err = dawg.NearestStore(c.addr, cfg.Service)
		if err != nil {
			return err
		}
	}

	c.menu, err = store.Menu()
	if err != nil {
		return err
	}
	rawMenu, err := json.Marshal(c.menu)
	if err != nil {
		return err
	}
	return db.Put("menu", rawMenu)
}

func maxStrLen(list []interface{}) int {
	max := 0
	for _, s := range list {
		length := strLen(s.(string))
		if length > max {
			max = length
		}
	}

	return max
}

var strLen = utf8.RuneCountInString // this is a function

func spaces(i int) string {
	return strings.Repeat(" ", i)
}
