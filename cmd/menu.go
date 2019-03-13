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
}

func (c *menuCmd) run(cmd *cobra.Command, args []string) (err error) {
	if store == nil {
		if store, err = dawg.NearestStore(c.addr, cfg.Service); err != nil {
			return err
		}
	}

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

	if c.toppings {
		c.printToppings()
		return nil
	}
	c.printMenu()
	return nil
}

func (b *cliBuilder) newMenuCmd() cliCommand {
	c := &menuCmd{all: false, toppings: false}
	c.basecmd = b.newBaseCommand("menu", "Get the Dominos menu.", c.run)

	c.cmd.Flags().BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	c.cmd.Flags().BoolVarP(&c.toppings, "toppings", "t", c.toppings, "print out the toppings on the menu")
	return c
}

func (c *menuCmd) printMenu() {
	var f func(map[string]interface{}, string)

	f = func(m map[string]interface{}, spacer string) {
		cats := m["Categories"].([]interface{})
		prods := m["Products"].([]interface{})

		// if there is nothing in that category, dont print the code name
		if len(cats) != 0 || len(prods) != 0 {
			fmt.Print(spacer, m["Code"], "\n")
		}
		if len(cats) > 0 { // the recursive part
			for _, c := range cats {
				f(c.(map[string]interface{}), spacer+"  ")
			}
		} else if len(prods) > 0 { // the printing part
			var product map[string]interface{}
			for _, p := range prods {
				key := p.(string)

				if prod, ok := c.menu.Products[key]; ok {
					product = prod.(map[string]interface{})
				} else if prod, ok := c.menu.Variants[key]; ok {
					product = prod.(map[string]interface{})
				} else if prod, ok := c.menu.Preconfigured[key]; ok {
					product = prod.(map[string]interface{})
				} else {
					continue
				}
				c.printMenuVarient(product, spacer)
			}
			print("\n")
		}
	}
	f(c.menu.Categorization["Food"].(map[string]interface{}), "")
	f(c.menu.Categorization["PreconfiguredProducts"].(map[string]interface{}), "")
}

func (c *menuCmd) printMenuVarient(product map[string]interface{}, spacer string) {
	if varients, ok := product["Variants"].([]interface{}); ok {
		fmt.Printf("%s  \"%s\" (%s)\n", spacer, product["Name"], product["Code"])
		// max := maxStrLen(varients)
		for _, v := range varients {
			fmt.Println(spaces(8), "-", v, spaces(10-strLen(v.(string))), c.menu.Variants[v.(string)].(map[string]interface{})["Name"])
		}
	} else {
		// fmt.Println("no varients")
		fmt.Printf("%s  \"%s\"\n%s - %s", spacer, product["Name"], spacer+"    ", product["Code"])
	}
}

func (c *menuCmd) printToppings() {
	indent := strings.Repeat(" ", 4)
	for key, val := range c.menu.Toppings {
		fmt.Print("  ", key, "\n")
		for k, v := range val.(map[string]interface{}) {
			spacer := strings.Repeat(" ", 3-strLen(k))
			fmt.Print(
				indent, k, spacer, v.(map[string]interface{})["Name"], "\n")
		}
		print("\n")
	}
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
