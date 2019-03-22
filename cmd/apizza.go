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

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/dawg"
)

var (
	addr  *dawg.Address
	store *dawg.Store
)

type apizzaCmd struct {
	*basecmd
	address    string
	service    string
	storeID    string
	test       bool
	clearCache bool
}

func (c *apizzaCmd) run(cmd *cobra.Command, args []string) (err error) {
	if test {
		all, err := db.GetAll()
		if err != nil {
			return err
		}
		for k := range all {
			fmt.Fprintf(c.output, "%v\n", k)
		}
		return nil
	}
	if c.clearCache {
		if err := db.Close(); err != nil {
			return err
		}
		fmt.Fprintln(c.output, "removing", db.Path)
		return os.Remove(db.Path)
	}
	return cmd.Usage()
}

var test bool

func newApizzaCmd() cliCommand {
	c := &apizzaCmd{address: "", service: cfg.Service, clearCache: false}
	c.basecmd = newBaseCommand("apizza", "Dominos pizza from the command line.", c.run)

	// c.cmd.PersistentFlags().StringVar(&c.address, "address", c.address, "use a specific address")
	c.cmd.PersistentFlags().StringVar(&c.service, "service", c.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	c.cmd.PersistentFlags().BoolVar(&test, "test", false, "testing flag (for development)")
	c.cmd.PersistentFlags().MarkHidden("test")

	c.cmd.Flags().BoolVar(&c.clearCache, "clear-cache", c.clearCache, "delete the database used for caching")
	return c
}
