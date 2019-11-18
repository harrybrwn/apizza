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
	"os"
	"path/filepath"
	"time"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
)

var db *cache.DataBase

// Execute runs the root command
func Execute() {
	var err error
	if err = config.SetConfig(".apizza", cfg); err != nil {
		handle(err, "Internal Error", 1)
	}

	builder := &cliBuilder{
		root: newApizzaCmd(),
		addr: &cfg.Address,
	}

	dbPath := filepath.Join(config.Folder(), "cache", "apizza.db")
	if db, err = cache.GetDB(dbPath); err != nil {
		handle(err, "Internal Error", 1)
	}

	defer func() {
		err = errs.Pair(db.Close(), config.Save())
		if err != nil {
			handle(err, "Internal Error", 1)
		}
	}()

	if err = builder.exec(); err != nil {
		handle(err, "Error", 1)
	}
}

func handle(e error, msg string, code int) { fmt.Fprintf(os.Stderr, "%s: %s\n", msg, e); os.Exit(code) }

var _ base.CliCommand = (*basecmd)(nil)

type basecmd struct {
	*base.Command
	cache.Updater

	menu *dawg.Menu

	// don't access this field directly, use store() to get the store
	dstore *dawg.Store
	addr   *obj.Address
	db     *cache.DataBase
}

func (c *basecmd) store() *dawg.Store {
	if c.dstore == nil {
		var err error
		var address = c.addr
		if address == nil {
			address = &cfg.Address
		}
		if c.dstore, err = dawg.NearestStore(address, config.GetString("service")); err != nil {
			handle(err, "Internal Error", 1) // will exit
		}
	}
	return c.dstore
}

func (c *basecmd) cacheNewMenu() error {
	var e1, e2 error
	var raw []byte
	c.menu, e1 = c.store().Menu()
	raw, e2 = json.Marshal(c.menu)
	return errs.Append(e1, e2, db.Put("menu", raw))
}

func (c *basecmd) getCachedMenu() error {
	if c.menu == nil {
		c.menu = new(dawg.Menu)
		raw, err := db.Get("menu")
		if raw == nil {
			return c.cacheNewMenu()
		}
		err = errs.Pair(err, json.Unmarshal(raw, c.menu))
		if err != nil {
			return err
		}
		if c.menu.ID != c.store().ID {
			return c.cacheNewMenu()
		}
	}
	return nil
}

func (c *basecmd) init() *basecmd {
	c.Updater = cache.NewUpdater(menuUpdateTime, c.cacheNewMenu, c.getCachedMenu)
	return c
}

func newCommand(use, short string, r base.Runner) *basecmd {
	bc := &basecmd{Command: base.NewCommand(use, short, r.Run)}
	return bc.init()
}

type cliBuilder struct {
	root base.CliCommand
	addr *obj.Address
	db   *cache.DataBase
}

func (b *cliBuilder) Build(use, short string, r base.Runner) *base.Command {
	return base.NewCommand(use, short, r.Run)
}

func (b *cliBuilder) DB() *cache.DataBase {
	return b.db
}

func (b *cliBuilder) Output() io.Writer {
	return b.root.Output()
}

func (b *cliBuilder) Config() config.Config {
	return config.Object()
}

func newBuilder() *cliBuilder {
	return &cliBuilder{
		root: newApizzaCmd(),
		addr: &cfg.Address,
	}
}

// NewBuilder creates a cliBuilder that has a database.
func newCliBuilder(root base.CliCommand, database *cache.DataBase, out io.Writer) *cliBuilder {
	return &cliBuilder{
		root: root,
		addr: &cfg.Address,
		db:   database,
	}
}

// builds the command tree...
//
// Some of the newCommand functions are members of
// the cliBuilder because they need to access the cliBuilder's address field by
// useing the cliBuilder's 'newCommand' function instead of the basic newCommand
// function.
func (b *cliBuilder) exec() error {
	b.root.Addcmd(
		newCartCmd(b).Addcmd(
			newAddOrderCmd(b),
		),
		newConfigCmd().Addcmd(
			newConfigSet(),
			newConfigGet(),
		),
		b.newMenuCmd(),
		newOrderCmd(),
		newDumpCmd(newCliBuilder(b.root, db, b.root.Output())),
	)
	return b.root.Cmd().Execute()
}

func (b *cliBuilder) newCommand(use, short string, c base.Runner) *basecmd {
	base := newCommand(use, short, c)
	base.addr = b.addr
	return base
}
