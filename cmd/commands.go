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
	"path/filepath"

	"github.com/spf13/cobra"

	"apizza/dawg"
	"apizza/pkg/cache"
	"apizza/pkg/config"
)

var db *cache.DataBase

// Execute runs the root command
func Execute() {
	err := config.SetConfig(".apizza", cfg)
	if err != nil {
		handle(err)
	}

	builder := newBuilder()

	db, err = cache.GetDB(builder.dbPath())
	if err != nil {
		handle(err)
	}

	defer func() {
		err = db.Close()
		if err != nil {
			handle(err)
		}
		err = config.Save()
		if err != nil {
			handle(err)
		}
	}()

	_, err = builder.exec()
	if err != nil {
		handle(err)
	}
}

func handle(e error) { fmt.Println(e); os.Exit(1) }

type cliCommand interface {
	command() *cobra.Command
	AddCmd(...cliCommand) cliCommand
	run(*cobra.Command, []string) error
}

type basecmd struct {
	cmd  *cobra.Command
	addr *dawg.Address
}

func (bc *basecmd) command() *cobra.Command {
	return bc.cmd
}

func (bc *basecmd) AddCmd(cmds ...cliCommand) cliCommand {
	for _, cmd := range cmds {
		bc.cmd.AddCommand(cmd.command())
	}
	return bc
}

func (bc *basecmd) run(cmd *cobra.Command, args []string) error {
	return bc.cmd.Usage()
}

type runFunc func(*cobra.Command, []string) error

func newVerboseBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{cmd: &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  f,
	}}
	if f == nil {
		base.cmd.RunE = base.run
	}
	return base
}

func newBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{cmd: &cobra.Command{
		Use:           use,
		Short:         short,
		RunE:          f,
		SilenceErrors: true,
		SilenceUsage:  true,
	}}
	if f == nil {
		base.cmd.RunE = base.run
	}

	return base
}

type commandBuilder interface {
	exec()
}

type cliBuilder struct {
	root cliCommand
	addr *dawg.Address
}

func newBuilder() *cliBuilder {
	b := &cliBuilder{root: newApizzaCmd()}
	b.addr = b.getAddress()
	return b
}

func (b *cliBuilder) exec() (*cobra.Command, error) {
	b.root.AddCmd(
		newOrderCommand().AddCmd(
			b.newNewOrderCmd(),
		),
		newConfigCmd().AddCmd(
			newConfigSet(),
			newConfigGet(),
		),
		b.newMenuCmd(),
	)
	return b.root.command().ExecuteC()
}

func (b *cliBuilder) getAddress() *dawg.Address {
	addrStr := b.root.(*apizzaCmd).address
	if addrStr == "" {
		return &dawg.Address{
			Street: cfg.Address.Street,
			City:   cfg.Address.City,
			State:  cfg.Address.State,
			Zip:    cfg.Address.Zip,
		}
	}
	return dawg.ParseAddress(addrStr)
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
