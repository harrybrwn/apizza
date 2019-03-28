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
	"github.com/harrybrwn/apizza/cmd/internal/cmds"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
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

	builder := &cliBuilder{
		root: newApizzaCmd(),
		addr: configAddr(),
	}

	dbPath := filepath.Join(config.Folder(), "cache", "apizza.db")
	if db, err = cache.GetDB(dbPath); err != nil {
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

	if err = builder.exec(); err != nil {
		handle(err, "Error", 0)
	}
}

func handle(e error, msg string, code int) { fmt.Fprintf(os.Stderr, "%s: %s\n", msg, e); os.Exit(code) }

var _ base.CliCommand = (*basecmd)(nil)

type basecmd struct {
	*base.Command
	menu    *dawg.Menu
	dstore  *dawg.Store
	tsDecay time.Duration
	addr    *obj.Address
}

func (c *basecmd) store() *dawg.Store {
	var (
		s   *dawg.Store
		err error
	)

	if c.dstore == nil {
		if s, err = dawg.NearestStore(c.addr, config.GetString("service")); err != nil {
			handle(err, "Internal Error", 1) // will exit
			return nil
		}
		c.dstore = s
		return s
	}
	return c.dstore
}

func (c *basecmd) cacheNewMenu() (err error) {
	c.menu, err = c.store().Menu()
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
		c.menu = new(dawg.Menu)
		raw, err := db.Get("menu")
		if err != nil {
			return err
		}
		if err = json.Unmarshal(raw, c.menu); err != nil {
			return err
		}
		if c.menu.ID != c.store().ID {
			return c.cacheNewMenu()
		}
	}
	return nil
}

var _ cache.Updater = (*basecmd)(nil)

func (c *basecmd) OnUpdate() error {
	return c.cacheNewMenu()
}

func (c *basecmd) NotUpdate() error {
	return c.getCachedMenu()
}

func (c *basecmd) Decay() time.Duration {
	return c.tsDecay
}

func newCommand(use, short string, c base.Runner) *basecmd {
	return &basecmd{
		Command: base.NewCommand(use, short, c.Run),
		tsDecay: 12 * time.Hour,
	}
}

type cliBuilder struct {
	root base.CliCommand
	addr *obj.Address
}

func newBuilder() *cliBuilder {
	return &cliBuilder{
		root: newApizzaCmd(),
		addr: &cfg.Address,
	}
}

func (b *cliBuilder) exec() error {
	b.root.Addcmd(
		b.newCartCmd().Addcmd(
			b.newAddOrderCmd(),
		),
		newConfigCmd().Addcmd(
			newConfigSet(),
			newConfigGet(),
		),
		b.newMenuCmd(),
	)
	b.root.AddCobraCmd(cmds.OrderCmd)
	return b.root.Cmd().Execute()
}

func (b *cliBuilder) newCommand(use, short string, c base.Runner) *basecmd {
	base := newCommand(use, short, c)
	base.addr = b.addr
	return base
}

func configAddr() *obj.Address {
	a := config.Get("address").(obj.Address)
	return &a
}
