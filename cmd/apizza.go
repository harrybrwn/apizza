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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/pkg/config"
)

type apizzaCmd struct {
	*basecmd
	address    string
	service    string
	storeID    string
	test       bool
	clearCache bool
}

func (c *apizzaCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if test {
		all, err := db.Map()
		if err != nil {
			return err
		}
		for k := range all {
			c.Printf("%v\n", k)
		}
		return nil
	}
	if reset {
		if err = db.Destroy(); err != nil {
			return err
		}
		if err = os.Remove(filepath.Dir(db.Path())); err != nil {
			return err
		}
		if err = os.Remove(config.File()); err != nil {
			return err
		}
		if err = os.Remove(config.Folder()); err != nil {
			return err
		}
		return err
	}
	if c.clearCache {
		if err := db.Close(); err != nil {
			return err
		}
		c.Printf("removing %s\n", db.Path())
		return os.Remove(db.Path())
	}
	return cmd.Usage()
}

var test = false
var reset = false

func newApizzaCmd() base.CliCommand {
	c := &apizzaCmd{address: "", service: cfg.Service, clearCache: false}
	c.basecmd = newCommand("apizza", "Dominos pizza from the command line.", c)

	// c.cmd.PersistentFlags().StringVar(&c.address, "address", c.address, "use a specific address")
	c.Cmd().PersistentFlags().StringVar(&c.service, "service", c.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	c.Cmd().PersistentFlags().BoolVar(&test, "test", false, "testing flag (for development)")
	c.Cmd().PersistentFlags().BoolVar(&reset, "reset", false, "reset the program (for development)")
	c.Cmd().PersistentFlags().MarkHidden("test")
	c.Cmd().PersistentFlags().MarkHidden("reset")
	return c
}
