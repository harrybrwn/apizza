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
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"apizza/dawg"
)

type menuCmd struct {
	*basecmd
	menu     *dawg.Menu
	all      bool
	food     bool
	toppings bool
}

func (c *menuCmd) run(cmd *cobra.Command, args []string) (err error) {
	if store == nil {
		store, err = dawg.NearestStore(c.addr, cfg.Service)
		if err != nil {
			return err
		}
	}

	cachedMenu, err := db.Get("menu")
	if err != nil {
		return err
	}
	if !db.Exists("menu_t") {
		newtime := strconv.Itoa(time.Now().Second())
		err = db.Put("menu_t", []byte(newtime))
	}

	menuCache, err := db.TimeStamp("menu")
	if err != nil {
		return err
	}

	if cachedMenu != nil && time.Since(menuCache) < 30*time.Minute {
		c.menu = &dawg.Menu{}
		err = json.Unmarshal(cachedMenu, c.menu)
		if err != nil {
			return err
		}
	} else {
		err = db.ResetTimeStamp("menu")
		if err != nil {
			return err
		}
		err = c.cacheNewMenu()
		if err != nil {
			return err
		}
	}

	if c.toppings {
		c.printToppings()
	} else if c.food {
		c.printMenu()
	}
	return nil
}

func (b *cliBuilder) newMenuCmd() cliCommand {
	c := &menuCmd{all: false, food: true, toppings: false}
	c.basecmd = b.newBaseCommand("menu", "Get the Dominos menu.", c.run)

	c.cmd.Flags().BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	c.cmd.Flags().BoolVarP(&c.food, "food", "f", c.food, "print out the food items on the menu")
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
			var prod map[string]interface{}
			max := maxStrLen(prods) + 2
			for _, p := range prods {
				_, ok := c.menu.Products[p.(string)]
				if ok {
					prod = c.menu.Products[p.(string)].(map[string]interface{})
				} else {
					continue
				}
				space := strings.Repeat(" ", max-strLen(p.(string)))
				fmt.Print(spacer+"  ", p, space, prod["Name"], "\n")
			}
			print("\n")
		}
	}

	f(c.menu.Categorization["Food"].(map[string]interface{}), "")
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
