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
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
)

var (
	db             *cache.DataBase
	menuUpdateTime = 12 * time.Hour
)

// Execute runs the root command
func Execute() {
	var (
		err, logErr error
	)
	if err = config.SetConfig(".apizza", cfg); err != nil {
		handle(err, "Internal Error", 1)
	}

	if db, err = data.NewDatabase(); err != nil {
		handle(err, "Internal Error", 1)
	}

	app := newapp(db, cfg, os.Stdout)

	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join(config.Folder(), "logs", "dev.log"),
		MaxSize:    25,  // megabytes
		MaxBackups: 10,  // number of spare files
		MaxAge:     365, //days
		Compress:   false,
	})

	defer func() {
		err = errs.Append(db.Close(), config.Save()) //, logs.Close())
		if err != nil {
			handle(err, "Internal Error", 1)
		}
	}()

	if err = errs.Pair(logErr, app.exec()); err != nil {
		handle(err, "Error", 1)
	}
}

func handle(e error, msg string, code int) {
	log.Printf("(Failure) %s: %s\n", msg, e)
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, e)
	os.Exit(code)
}

var _ base.CliCommand = (*basecmd)(nil)

type basecmd struct {
	*base.Command
	cache.Updater
	storefinder
	menu *dawg.Menu

	addr *obj.Address
	db   *cache.DataBase
}

func (c *basecmd) cacheNewMenu() error {
	var e1, e2 error
	var raw []byte
	c.menu, e1 = c.store().Menu()
	log.Println("caching another menu")
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
	c.storefinder = newStoreGetter(serviceGetter, func() dawg.Address {
		if c.addr == nil {
			return &cfg.Address
		}
		return c.addr
	})
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
		newConfigCmd(nil).Addcmd(
			newConfigSet(),
			newConfigGet(),
		),
		b.newMenuCmd(),
		newOrderCmd(),
	)
	return b.root.Cmd().Execute()
}

func (b *cliBuilder) newCommand(use, short string, c base.Runner) *basecmd {
	base := newCommand(use, short, c)
	base.addr = b.addr
	return base
}

type storefinder interface {
	store() *dawg.Store
}

// storegetter is meant to be a mixin for any struct that needs to be able to
// get a store.
type storegetter struct {
	getaddr   func() dawg.Address
	getmethod func() string
	dstore    *dawg.Store
}

var serviceGetter = func() string {
	return config.GetString("service")
}

func newStoreGetter(service func() string, addr func() dawg.Address) storefinder {
	return &storegetter{
		getmethod: service,
		getaddr:   addr,
		dstore:    nil,
	}
}

func (s *storegetter) store() *dawg.Store {
	if s.dstore == nil {
		var err error
		var address = s.getaddr()
		s.dstore, err = dawg.NearestStore(address, s.getmethod())
		if err != nil {
			handle(err, "Store Find Error", 1) // will exit
		}
	}
	return s.dstore
}

type menuupdater interface {
}

type menuUpdater struct {
}
