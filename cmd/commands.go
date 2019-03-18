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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
)

var db *cache.DataBase

// Execute runs the root command
func Execute() {
	var err error
	if err = config.SetConfig(".apizza", cfg); err != nil {
		handle(err, "", 1)
	}

	builder := newBuilder()

	if db, err = cache.GetDB(builder.dbPath()); err != nil {
		handle(err, "", 1)
	}

	defer func() {
		if err = db.Close(); err != nil {
			handle(err, "", 1)
		}
		if err = config.Save(); err != nil {
			handle(err, "", 1)
		}
	}()

	if _, err = builder.exec(); err != nil {
		handle(err, "Error", 0)
	}
}

func handle(e error, msg string, code int) { fmt.Printf("%s: %s\n", msg, e); os.Exit(code) }

type cliCommand interface {
	command() *cobra.Command
	AddCmd(...cliCommand) cliCommand
	run(*cobra.Command, []string) error
	setOutput(io.Writer)
}

type basecmd struct {
	cmd    *cobra.Command
	addr   *address
	output io.Writer
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

func (bc *basecmd) setOutput(output io.Writer) {
	bc.output = output
	bc.cmd.SetOutput(output)
}

func (bc *basecmd) getstore() (err error) {
	if store == nil {
		if store, err = dawg.NearestStore(bc.addr, cfg.Service); err != nil {
			return err
		}
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
	if f == nil {
		base.cmd.RunE = base.run
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
	addr *address
}

func newBuilder() *cliBuilder {
	b := &cliBuilder{root: newApizzaCmd()}
	b.addr = b.getAddress()
	return b
}

func (b *cliBuilder) exec() (*cobra.Command, error) {
	b.root.AddCmd(
		b.newOrderCommand().AddCmd(
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

func (b *cliBuilder) getAddress() *address {
	addrStr := b.root.(*apizzaCmd).address
	if addrStr == "" {
		return cfg.Address
	}
	// return dawg.ParseAddress(addrStr)
	return nil
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
