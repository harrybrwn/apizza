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
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
)

var db *cache.DataBase

// Execute runs the root command
func Execute() {
	var err error
	if err = config.SetConfig(".apizza", cfg); err != nil {
		handle(err, "Internal Error", 1)
	}

	builder := newBuilder()

	if db, err = cache.GetDB(builder.dbPath()); err != nil {
		handle(err, "Internal Error", 1)
	}

	defer func() {
		if err = db.Close(); err != nil {
			handle(err, "Internal Error", 1)
		}
		if err = config.Save(); err != nil {
			handle(err, "Internal Error", 1)
		}
	}()

	if _, err = builder.exec(); err != nil {
		handle(err, "Error", 0)
	}
}

func handle(e error, msg string, code int) { fmt.Fprintf(os.Stderr, "%s: %s\n", msg, e); os.Exit(code) }

type cliCommand interface {
	command() *cobra.Command
	addcmd(...cliCommand) cliCommand
	run(*cobra.Command, []string) error
	setOutput(io.Writer)
}

type basecmd struct {
	cmd    *cobra.Command
	addr   *address
	menu   *dawg.Menu
	output io.Writer
}

func (c *basecmd) command() *cobra.Command {
	return c.cmd
}

func (c *basecmd) addcmd(cmds ...cliCommand) cliCommand {
	for _, cmd := range cmds {
		c.cmd.AddCommand(cmd.command())
	}
	return c
}

func (c *basecmd) run(cmd *cobra.Command, args []string) error {
	return c.cmd.Usage()
}

func (c *basecmd) setOutput(output io.Writer) {
	c.output = output
	c.cmd.SetOutput(output)
}

func (c *basecmd) getstore() (err error) {
	if store == nil {
		if store, err = dawg.NearestStore(c.addr, cfg.Service); err != nil {
			return err
		}
	}
	return nil
}

func (c *basecmd) Flags() *pflag.FlagSet {
	return c.cmd.Flags()
}

func (c *basecmd) cacheNewMenu() (err error) {
	if err = c.getstore(); err != nil {
		return err
	}

	c.menu, err = store.Menu()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(c.menu)
	if err != nil {
		return err
	}
	return db.Put("menu", raw)
}

func (c *basecmd) getCachedMenu() error {
	if c.menu == nil {
		raw, err := db.Get("menu")
		if err != nil {
			return err
		}
		c.menu = &dawg.Menu{}
		return json.Unmarshal(raw, c.menu)
	}
	return nil
}

type runFunc func(*cobra.Command, []string) error

func newVerboseBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{
		cmd: &cobra.Command{
			Use:   use,
			Short: short,
			RunE:  f,
		},
		output: os.Stdout,
	}

	return base
}

func newBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{
		cmd: &cobra.Command{
			Use:           use,
			Short:         short,
			RunE:          f,
			SilenceErrors: true,
			SilenceUsage:  true,
		},
		output: os.Stdout,
	}

	return base
}

type commandBuilder interface {
	exec()
}

type cliBuilder struct {
	root cliCommand
	addr *address
}

func newBuilder() *cliBuilder {
	b := &cliBuilder{root: newApizzaCmd()}

	addrStr := b.root.(*apizzaCmd).address
	if addrStr == "" {
		b.addr = &cfg.Address
	} else {
		b.addr = nil
	}
	return b
}

func (b *cliBuilder) exec() (*cobra.Command, error) {
	b.root.addcmd(
		b.newCartCmd().addcmd(
			b.newAddOrderCmd(),
		),
		newConfigCmd().addcmd(
			newConfigSet(),
			newConfigGet(),
		),
		b.newMenuCmd(),
	)
	return b.root.command().ExecuteC()
}

// this is here for future plans
func (b *cliBuilder) newBaseCommand(use, short string, f runFunc) *basecmd {
	base := newBaseCommand(use, short, f)
	base.addr = b.addr
	return base
}

func (b *cliBuilder) dbPath() string {
	return filepath.Join(config.Folder(), "cache", "apizza.db")
}
