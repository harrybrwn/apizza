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
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/errs"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"
)

var menuUpdateTime = 12 * time.Hour

type menuCmd struct {
	base.CliCommand
	data.MenuCacher
	client.StoreFinder

	db *cache.DataBase

	addr dawg.Address

	all            bool
	page           bool
	verbose        bool
	toppings       bool
	preconfigured  bool
	showCategories bool
	item           string
	category       string
}

func (c *menuCmd) Run(cmd *cobra.Command, args []string) error {
	if err := c.db.UpdateTS("menu", c); err != nil {
		return err
	}
	out.SetOutput(c.Output())
	defer out.ResetOutput()

	var item dawg.Item

	if len(args) == 1 {
		if c.item == "" {
			item = c.Menu().FindItem(args[0])
		}

		if item == nil && c.category == "" {
			c.category = strings.ToLower(args[0])
		} else {
			return out.ItemInfo(item, c.Menu())
		}
	}

	if c.item != "" {
		prod := c.Menu().FindItem(c.item)
		if prod == nil {
			return fmt.Errorf("cannot find %s", c.item)
		}
		return out.ItemInfo(prod, c.Menu())
	}

	if c.toppings {
		c.printToppings()
		return nil
	}

	// print menu handles most of the menu command's flags
	if c.page {
		return c.pageMenu(strings.ToLower(c.category))
	}
	return c.printMenu(c.Output(), strings.ToLower(c.category)) // still works with an empty string
}

// NewMenuCmd creates a new menu command.
func NewMenuCmd(b base.Builder) base.CliCommand {
	c := &menuCmd{
		db:             b.DB(),
		all:            false,
		toppings:       false,
		preconfigured:  false,
		showCategories: false,
	}
	// TODO: this will not work with a global service or address flag
	if app, ok := b.(*App); ok {
		c.StoreFinder = app
	} else {
		c.StoreFinder = client.NewStoreGetter(b)
	}

	c.CliCommand = b.Build("menu <item>", "View the Dominos menu.", c)
	c.MenuCacher = data.NewMenuCacher(menuUpdateTime, b.DB(), c.Store)
	c.SetOutput(b.Output())

	c.Cmd().Long = `This command will show the dominos menu.

To show a subdivition of the menu, give an item or
category to the --category and --item flags or give them
as an argument to the command itself.`

	flags := c.Flags()
	flags.BoolVarP(&c.all, "all", "a", c.all, "show the entire menu")
	flags.BoolVarP(&c.verbose, "verbose", "v", false, "print the menu verbosly")
	flags.BoolVar(&c.page, "page", false, "pipe the menu to a pager")

	flags.StringVarP(&c.item, "item", "i", "", "show info on the menu item given")
	flags.StringVarP(&c.category, "category", "c", "", "show one category on the menu")

	flags.BoolVarP(&c.toppings, "toppings", "t", c.toppings, "print out the toppings on the menu")
	flags.BoolVarP(&c.preconfigured, "preconfigured",
		"p", c.preconfigured, "show the pre-configured products on the dominos menu")
	flags.BoolVar(&c.showCategories, "show-categories", c.showCategories, "print categories")
	return c
}

func (c *menuCmd) printMenu(w io.Writer, name string) error {
	out.SetOutput(w)
	defer out.ResetOutput()
	menu := c.Menu()
	var allCategories = menu.Categorization.Food.Categories
	if c.preconfigured {
		allCategories = menu.Categorization.Preconfigured.Categories
	} else if c.all {
		allCategories = append(allCategories, menu.Categorization.Preconfigured.Categories...)
	}

	if len(name) > 0 {
		for _, cat := range allCategories {
			if name == strings.ToLower(cat.Name) || name == strings.ToLower(cat.Code) {
				return out.PrintMenu(cat, 0, menu)
			}
		}
		return fmt.Errorf("could not find %s", name)
	} else if c.showCategories {
		for _, cat := range allCategories {
			if cat.Name != "" {
				fmt.Fprintln(w, strings.ToLower(cat.Name))
			}
		}
		return nil
	}

	for _, cat := range allCategories {
		out.PrintMenu(cat, 0, menu)
	}
	return nil
}

func (c *menuCmd) printToppings() {
	var tops = c.Menu().Toppings

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

func (c *menuCmd) pageMenu(category string) error {
	less := exec.Command("less")
	less.Stdout = c.Output()
	stdin, err := less.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		err = c.printMenu(stdin, strings.ToLower(category)) // still works with an empty string
		errs.Handle(err, "io Error", 1)
	}()

	return less.Run()
}

func printToppingCategory(name string, toppings map[string]dawg.Topping, w io.Writer) {
	fmt.Fprintln(w, "  ", name)
	indent := strings.Repeat(" ", 4)
	for k, v := range toppings {
		fmt.Fprintln(w, indent, k, strings.Repeat(" ", 3-strLen(k)), v.Name)
	}
	fmt.Fprintln(w, "")
}

var strLen = utf8.RuneCountInString // this is a function

func spaces(i int) string {
	return strings.Repeat(" ", i)
}
