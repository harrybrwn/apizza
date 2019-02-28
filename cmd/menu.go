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
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"apizza/dawg"
)

type menuCmd struct {
	*basecmd
	all, food, toppings bool
}

func (c *menuCmd) run(cmd *cobra.Command, args []string) (err error) {
	if store == nil {
		store, err = dawg.NearestStore(c.addr, cfg.Service)
		if err != nil {
			return err
		}
	}

	err = menuManagment()
	if err != nil {
		return err
	}

	if c.toppings {
		printToppings()
	} else if c.food {
		printMenu()
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

func printMenu() {
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
				_, ok := menu.Products[p.(string)]
				if ok {
					prod = menu.Products[p.(string)].(map[string]interface{})
				} else {
					continue
				}
				space := strings.Repeat(" ", max-strLen(p.(string)))
				fmt.Print(spacer+"  ", p, space, prod["Name"], "\n")
			}
			print("\n")
		}
	}

	f(menu.Categorization["Food"].(map[string]interface{}), "")
}

func printToppings() {
	indent := strings.Repeat(" ", 4)
	for key, val := range menu.Toppings {
		fmt.Print("  ", key, "\n")
		for k, v := range val.(map[string]interface{}) {
			spacer := strings.Repeat(" ", 3-strLen(k))
			fmt.Print(
				indent, k, spacer, v.(map[string]interface{})["Name"], "\n")
		}
		print("\n")
	}
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
