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
	"text/template"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"
)

type menuCmd struct {
	*basecmd
	all           bool
	toppings      bool
	preconfigured bool
	categories    bool
	item          string
}

func (c *menuCmd) Run(cmd *cobra.Command, args []string) error {
	if err := db.UpdateTS("menu", c); err != nil {
		return err
	}

	if len(c.item) > 0 {
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
	c.printMenu()
	return nil
}

func (b *cliBuilder) newMenuCmd() base.CliCommand {
	c := &menuCmd{all: false, toppings: false, preconfigured: false, categories: false}
	c.basecmd = b.newCommand("menu", "Get the Dominos menu.", c)

	c.Flags().BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	c.Flags().BoolVarP(&c.toppings, "toppings", "t", c.toppings, "print out the toppings on the menu")
	c.Flags().BoolVarP(&c.preconfigured, "preconfigured",
		"p", c.preconfigured, "show the pre-configured products on the dominos menu")
	c.Flags().StringVarP(&c.item, "item", "i", "", "show info on the menu item given")
	c.Flags().BoolVarP(&c.categories, "categories", "c", c.categories, "print categories")
	return c
}

func (c *menuCmd) printMenu() {
	var f func(map[string]interface{}, string)

	f = func(m map[string]interface{}, spacer string) {
		cats := m["Categories"].([]interface{})
		prods := m["Products"].([]interface{})

		// if there is nothing in that category, dont print the code name
		if len(cats) != 0 || len(prods) != 0 {
			c.Printf("%s%v\n", spacer, m["Code"])
		}
		if len(cats) > 0 { // the recursive part
			for _, c := range cats {
				f(c.(map[string]interface{}), spacer+"  ")
			}
		} else if len(prods) > 0 && !c.categories { // the printing part
			for _, p := range prods {
				product, err := c.findProduct(p.(string))
				if err != nil {
					continue
				}
				c.printMenuItem(product, spacer)
			}
			c.Printf("%s", "\n")
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

func tmpl(w io.Writer, templt string, a interface{}) error {
	t := template.New("apizza")
	template.Must(t.Parse(templt))
	return t.Execute(w, a)
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
