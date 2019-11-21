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
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/pkg/errs"
)

type apizzaCmd struct {
	*basecmd
	address    string
	service    string
	storeID    string
	clearCache bool
	resetMenu  bool

	storeLocation bool
}

func (c *apizzaCmd) Run(cmd *cobra.Command, args []string) (err error) {
	// panic("get rid of this command")
	if c.clearCache {
		err = db.Close()
		c.Printf("removing %s\n", db.Path())
		return errs.Pair(err, os.Remove(db.Path()))
	}
	if c.storeLocation {
		c.Println(c.store().Address)
		c.Printf("\n")
		c.Println("Store id:", c.store().ID)
		c.Printf("Coordinates: %s, %s\n",
			c.store().StoreCoords["StoreLatitude"],
			c.store().StoreCoords["StoreLongitude"],
		)
		return nil
	}
	return cmd.Usage()
}

var test = false
var reset = false

func newApizzaCmd() base.CliCommand {
	c := &apizzaCmd{address: "", service: "", clearCache: false}
	c.basecmd = newCommand("apizza", "Dominos pizza from the command line.", c)

	// c.Cmd().PersistentPreRunE = c.preRun

	// c.Flags().BoolVar(&c.clearCache, "clear-cache", false, "delete the database")
	// c.Cmd().PersistentFlags().BoolVar(&c.resetMenu, "delete-menu", false, "delete the menu stored in cache")

	// c.Cmd().PersistentFlags().StringVar(&c.address, "address", c.address, "use a specific address")
	// c.Cmd().PersistentFlags().StringVar(&c.service, "service", c.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	// c.Cmd().PersistentFlags().BoolVar(&test, "test", false, "testing flag (for development)")
	// c.Cmd().PersistentFlags().BoolVar(&reset, "reset", false, "reset the program (for development)")
	// c.Cmd().PersistentFlags().MarkHidden("test")
	// c.Cmd().PersistentFlags().MarkHidden("reset")

	// c.Flags().BoolVarP(&c.storeLocation, "store-location", "L", false, "show the location of the nearest store")
	return c
}

func (c *apizzaCmd) preRun(cmd *cobra.Command, args []string) (err error) {
	if c.resetMenu {
		err = db.Delete("menu")
	}
	return
}

func yesOrNo(msg string) bool {
	var in string
	fmt.Printf("%s ", msg)
	fmt.Scan(&in)

	switch strings.ToLower(in) {
	case "y", "yes", "si":
		return true
	}
	return false
}
