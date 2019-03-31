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
	"io"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"
)

type menuCmd struct {
	*basecmd
	all            bool
	toppings       bool
	preconfigured  bool
	showCategories bool
	item           string
	category       string
}

func (c *menuCmd) Run(cmd *cobra.Command, args []string) error {
	if err := db.UpdateTS("menu", c); err != nil {
		return err
	}

	if c.item != "" {
		prod, err := c.findProduct(c.item)
		if err != nil {
			return err
		}
		iteminfo(prod, cmd.OutOrStdout())
		return nil
	}

	if c.toppings {
		c.printToppings()
		return nil
	}

	return c.printMenu(c.category) // still works with an empty string
}

func (b *cliBuilder) newMenuCmd() base.CliCommand {
	c := &menuCmd{
		all: false, toppings: false,
		preconfigured: false, showCategories: false}
	c.basecmd = b.newCommand("menu", "Get the Dominos menu.", c)

	c.Flags().StringVarP(&c.item, "item", "i", "", "show info on the menu item given")
	c.Flags().StringVarP(&c.category, "category", "c", "", "show one category on the menu")

	c.Flags().BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	c.Flags().BoolVarP(&c.toppings, "toppings", "t", c.toppings, "print out the toppings on the menu")
	c.Flags().BoolVarP(&c.preconfigured, "preconfigured",
		"p", c.preconfigured, "show the pre-configured products on the dominos menu")
	c.Flags().BoolVar(&c.showCategories, "show-categories", c.showCategories, "print categories")
	return c
}

func (c *menuCmd) printMenu(categoryName string) error {
	var printfunc func(dawg.MenuCategory, int) error

	printfunc = func(cat dawg.MenuCategory, depth int) error {
		if cat.IsEmpty() {
			return nil
		}
		c.Printf("%s%s\n", strings.Repeat("  ", depth), cat.Name)

		if cat.HasItems() {
			for _, p := range cat.Products {
				product, err := c.findProduct(p)
				if err != nil {
					// continue
					panic(err)
				}
				c.printMenuItem(product, strings.Repeat("  ", depth))
			}
		} else {
			for _, category := range cat.Categories {
				printfunc(category, depth+1)
			}
		}
		return nil
	}

	if len(categoryName) > 0 {
		for _, cat := range c.menu.Categorization.Food.Categories {
			if categoryName == cat.Name || categoryName == cat.Code {
				return printfunc(cat, 0)
			}
		}
		return fmt.Errorf("could not find %s", categoryName)
	} else if c.showCategories {
		for _, cat := range c.menu.Categorization.Food.Categories {
			if cat.Name != "" {
				fmt.Println(cat.Name)
			}
		}
		return nil
	}

	if c.preconfigured {
		return printfunc(c.menu.Categorization.Preconfigured, 0)
	} else if c.all {
		printfunc(c.menu.Categorization.Preconfigured, 0)
	}
	return printfunc(c.menu.Categorization.Food, 0)
}

func (c *menuCmd) printMenuItem(product map[string]interface{}, spacer string) {
	// if product has varients, print them
	if varients, ok := product["Variants"].([]interface{}); ok {
		c.Printf("%s  \"%s\" [%s]\n", spacer, product["Name"], product["Code"])
		max := maxStrLen(varients)

		for _, v := range varients {
			c.Println(spaces(strLen(spacer)+3), "-", v,
				spaces(max-strLen(v.(string))),
				c.menu.Variants[v.(string)].(map[string]interface{})["Name"])
		}
	} else {
		// if product has no varients, it is a preconfigured product
		c.Printf("%s  \"%s\"\n%s - %s", spacer, product["Name"], spacer+"    ", product["Code"])
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
		c.Println("  ", key)
		for k, v := range val.(map[string]interface{}) {
			spacer := strings.Repeat(" ", 3-strLen(k))
			c.Println(indent, k, spacer, v.(map[string]interface{})["Name"])
		}
		c.Println("")
	}
}

func (c *basecmd) findProduct(key string) (map[string]interface{}, error) {
	if c.menu == nil {
		if err := db.UpdateTS("menu", c); err != nil {
			return nil, err
		}
	}
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

func (c *basecmd) product(code string) (*dawg.Product, error) {
	if c.menu == nil {
		if err := db.UpdateTS("menu", c); err != nil {
			return nil, err
		}
	}
	return c.menu.GetProduct(code)
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
