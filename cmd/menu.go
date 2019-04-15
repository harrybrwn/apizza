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
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/harrybrwn/apizza/cmd/internal/out"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"
)

type menuCmd struct {
	*basecmd
	all            bool
	verbose        bool
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

	var item dawg.Item

	if len(args) == 1 {
		if c.item == "" {
			item = c.menu.FindItem(args[0])
		}

		if item == nil && c.category == "" {
			c.category = strings.ToLower(args[0])
		} else {
			c.iteminfo(item, c.Output())
			return nil
		}
	}

	if c.item != "" {
		prod := c.menu.FindItem(c.item)
		if prod == nil {
			return fmt.Errorf("cannot find %s", c.item)
		}
		c.iteminfo(prod, cmd.OutOrStdout())
		return nil
	}

	if c.toppings {
		c.printToppings()
		return nil
	}

	// print menu handles most of the menu command's flags
	return c.printMenu(strings.ToLower(c.category)) // still works with an empty string
}

func (b *cliBuilder) newMenuCmd() base.CliCommand {
	c := &menuCmd{
		all: false, toppings: false,
		preconfigured: false, showCategories: false}
	c.basecmd = b.newCommand("menu <item>", "View the Dominos menu.", c)
	c.Cmd().Long = `This command will show the dominos menu.

To show a subdivition of the menu, give an item or
category to the --category and --item flags or give them
as an argument to the command itself.`

	c.Flags().BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	c.Flags().BoolVarP(&c.verbose, "verbose", "v", false, "print the menu verbosly")

	c.Flags().StringVarP(&c.item, "item", "i", "", "show info on the menu item given")
	c.Flags().StringVarP(&c.category, "category", "c", "", "show one category on the menu")

	c.Flags().BoolVarP(&c.toppings, "toppings", "t", c.toppings, "print out the toppings on the menu")
	c.Flags().BoolVarP(&c.preconfigured, "preconfigured",
		"p", c.preconfigured, "show the pre-configured products on the dominos menu")
	c.Flags().BoolVar(&c.showCategories, "show-categories", c.showCategories, "print categories")
	return c
}

func (c *menuCmd) printMenu(name string) error {
	var printfunc func(dawg.MenuCategory, int) error

	printfunc = func(cat dawg.MenuCategory, depth int) error {
		if cat.IsEmpty() {
			return nil
		}
		c.Printf("\n%s%s%s%s\n", spaces(depth*2), strings.Repeat("-", 8),
			cat.Name, strings.Repeat("-", 60-len(cat.Name)-(depth*2)))

		if cat.HasItems() {
			for _, p := range cat.Products {
				c.printCategory(p, depth+1)
			}
		} else {
			for _, category := range cat.Categories {
				printfunc(category, depth+1)
			}
		}
		return nil
	}

	var allCategories = c.menu.Categorization.Food.Categories

	if c.preconfigured {
		allCategories = c.menu.Categorization.Preconfigured.Categories
	} else if c.all {
		allCategories = append(allCategories, c.menu.Categorization.Preconfigured.Categories...)
	}

	if len(name) > 0 {
		for _, cat := range allCategories {
			if name == strings.ToLower(cat.Name) || name == strings.ToLower(cat.Code) {
				return printfunc(cat, 0)
			}
		}
		return fmt.Errorf("could not find %s", name)
	} else if c.showCategories {
		for _, cat := range allCategories {
			if cat.Name != "" {
				c.Println(strings.ToLower(cat.Name))
			}
		}
		return nil
	}

	for _, c := range allCategories {
		printfunc(c, 0)
	}
	return nil
}

func (c *menuCmd) printCategory(code string, indentLen int) {
	item := c.menu.FindItem(code)

	switch product := item.(type) {
	case *dawg.Product:
		if len(product.Variants) == 1 {
			p, err := c.menu.GetVariant(product.Variants[0])
			if err != nil {
				panic(err)
			}

			c.Printf("%s%s  %s\n", spaces(indentLen*2), p.Code, p.Name)
			return
		}

		c.Printf("%s%s [%s]\n", spaces(indentLen*2), item.ItemName(), item.ItemCode())
		n := maxStrLen(product.Variants)
		for _, variant := range product.Variants {
			v, err := c.menu.GetVariant(variant)
			if err != nil {
				continue
			}
			c.Printf("%s%s %s %s\n",
				spaces(2*(indentLen+1)), variant,
				spaces(n-strLen(variant)), v.Name)
		}

	case *dawg.PreConfiguredProduct:
		c.Printf("%s%s   %s\n", spaces(indentLen*2),
			item.ItemCode(), item.ItemName())
	default:
		panic("dawg.Product and dawg.PreConfiguredProduct are the only catagories to be printed")
	}
}

func (c *menuCmd) iteminfo(prod dawg.Item, w io.Writer) {
	o := &bytes.Buffer{}
	out.SetOutput(o)
	out.ItemInfo(prod, c.menu)

	switch p := prod.(type) {
	case *dawg.Variant:
		fmt.Fprintf(o, "  Price: %s\n", p.Price)
		parent := p.GetProduct()
		if parent == nil {
			break
		}
		fmt.Fprintf(o, "  Parent Product: '%s' [%s]\n", parent.ItemName(), parent.ItemCode())

	case *dawg.PreConfiguredProduct:
		fmt.Fprintf(o, "  Description: '%s'\n", out.FormatLineIndent(p.Description, 70, 16))
		fmt.Fprintf(o, "  Size: %s\n", p.Size)

	case *dawg.Product:
		out.PrintProduct(p)
	}
	out.ResetOutput()
	w.Write(o.Bytes())
}

func (c *menuCmd) printToppings() {
	var tops = c.menu.Toppings

	if c.category != "" {
		category := strings.Title(c.category)
		printToppingCategory(category, tops[category], c.Output())
		return
	}

	if c.showCategories {
		for cat := range tops {
			c.Println(strings.ToLower(cat))
		}
		return
	}

	for typ, toppings := range tops {
		printToppingCategory(typ, toppings, c.Output())
	}
}

func printToppingCategory(name string, toppings map[string]dawg.Topping, w io.Writer) {
	fmt.Fprintln(w, "  ", name)
	indent := strings.Repeat(" ", 4)
	for k, v := range toppings {
		fmt.Fprintln(w, indent, k, strings.Repeat(" ", 3-strLen(k)), v.Name)
	}
	fmt.Fprintln(w, "")
}

func maxStrLen(list []string) int {
	max := 0
	for _, s := range list {
		max = getmax(s, max)
	}

	return max
}

func getmax(s string, i int) int {
	if strLen(s) > i {
		return len(s)
	}
	return i
}

var strLen = utf8.RuneCountInString // this is a function

func spaces(i int) string {
	return strings.Repeat(" ", i)
}
