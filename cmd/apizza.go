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
	"apizza/dawg"
	"apizza/pkg/config"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
)

var (
	addr  *dawg.Address
	menu  *dawg.Menu
	store *dawg.Store
	db    *bolt.DB
)

type apizzaCmd struct {
	basecmd
	address string
	service string
	storeID string
	test    bool
}

func (c *apizzaCmd) run(cmd *cobra.Command, args []string) (err error) {
	if c.test {
		fmt.Println("config file:", config.File())
	} else {
		err = cmd.Usage()
	}
	return err
}

func newApizzaCmd() cliCommand {
	c := &apizzaCmd{address: "", service: cfg.Service}

	c.basecmd.cmd = &cobra.Command{
		Use:   "apizza",
		Short: "Dominos pizza from the command line.",
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return config.Save()
		},
	}
	c.cmd.RunE = c.run

	c.cmd.PersistentFlags().StringVar(&c.address, "address", c.address, "use a specific address")
	c.cmd.PersistentFlags().StringVar(&c.service, "service", c.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	c.cmd.Flags().BoolVarP(&c.test, "test", "t", false, "testing flag")
	c.cmd.Flags().MarkHidden("test")
	return c
}
