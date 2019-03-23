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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/pkg/config"
)

var cfg = &Config{}

// Config is the configuration struct
type Config struct {
	Name    string      `config:"name"`
	Email   string      `config:"email"`
	Address obj.Address `config:"address"`
	Card    struct {
		Number     string `config:"number"`
		Expiration string `config:"expiration"`
		CVV        string `config:"cvv"`
	} `config:"card"`
	Service  string `config:"service" default:"\"Delivery\""`
	MyOrders []struct {
		Name string `config:"name"`
	} `config:"myorders"`
}

// Get a config variable
func (c *Config) Get(key string) interface{} {
	return config.Get(c, key)
}

// Set a config variable
func (c *Config) Set(key string, val interface{}) error {
	if config.FieldName(c, key) == "Service" {
		if val != "Delivery" && val != "Carryout" {
			return errors.New("service must be either 'Delivery' or 'Carryout'")
		}
	}
	return config.Set(c, key, val)
}

func (c *Config) printAll(output io.Writer) error {
	m := map[string]interface{}{}
	err := mapstructure.Decode(c, &m)
	if err != nil {
		return err
	}

	for k, v := range m {
		fmt.Fprintln(output, k, v)
	}
	return nil
}

type configCmd struct {
	*basecmd
	file       bool
	dir        bool
	resetCache bool
	getall     bool
}

func (c *configCmd) Run(cmd *cobra.Command, args []string) error {
	if c.file {
		c.Printf("%s\n", config.File())
		return nil
	}
	if c.dir {
		c.Printf("%s\n", config.Folder())
		return nil
	}
	if c.resetCache {
		return os.Remove(filepath.Join(config.Folder(), "cache", "apizza.db"))
	}
	if c.getall {
		return cfg.printAll(cmd.OutOrStdout())
	}
	return cmd.Usage()
}

func newConfigCmd() base.CliCommand {
	c := &configCmd{file: false, dir: false, resetCache: false}
	c.basecmd = newVerboseBaseCommand("config", "Configure apizza", c.Run)
	c.Cmd().Long = `The 'config' command is used for accessing the .apizza config file
in your home directory. Feel free to edit the .apizza json file
by hand or use the 'config' command.

ex. 'apizza config get name' or 'apizza config set name=<your name>'`

	c.Flags().BoolVarP(&c.file, "file", "f", c.file, "show the path to the config.json file")
	c.Flags().BoolVarP(&c.dir, "dir", "d", c.dir, "show the apizza config directory path")
	c.Flags().BoolVarP(&c.resetCache, "reset-cache", "r", c.resetCache, "reset the database cache")
	c.Flags().BoolVar(&c.getall, "get-all", c.getall, "show all the contents of the config file")
	return c
}

type configSetCmd struct {
	*basecmd
}

func (c *configSetCmd) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("no variable given")
	}

	for _, arg := range args {
		keys := strings.Split(arg, "=")
		if len(keys) < 2 || keys[0] == "" || keys[1] == "" {
			return errors.New(`use '<key>=<value>' format (no spaces)`)
		}
		err := cfg.Set(keys[0], keys[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func newConfigSet() base.CliCommand {
	c := &configSetCmd{}
	c.basecmd = newBaseCommand(
		"set",
		"change variables in the config file",
		c.Run,
	)
	return c
}

type configGetCmd struct {
	*basecmd
}

func (c *configGetCmd) Run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("no variable given")
	}

	for _, arg := range args {
		v := cfg.Get(arg)
		if v == nil {
			return fmt.Errorf("cannot find %s", arg)
		}
		c.Printf("%v", v)
	}
	return nil
}

func newConfigGet() base.CliCommand {
	c := &configGetCmd{}
	c.basecmd = newBaseCommand(
		"get",
		"print the specified config variable to screen",
		c.Run,
	)
	return c
}
