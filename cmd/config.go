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
	"strings"

	"github.com/spf13/cobra"

	"apizza/pkg/config"
)

var cfg = &Config{}

// Config is the configuration struct
type Config struct {
	Name    string `config:"name"`
	Email   string `config:"email"`
	Address struct {
		Street string `config:"street"`
		// Streetname   string `config:"streetname"`
		// Streetnumber string `config:"streetnumber"`
		City  string `config:"city"`
		State string `config:"state"`
		Zip   string `config:"zip"`
	} `config:"address"`
	Card struct {
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
	return config.Set(c, key, val)
}

func (c *Config) validateAddress() error {
	splitStreet := strings.Split(c.Address.Street, " ")
	fmt.Println(splitStreet)
	return nil
}

type configCmd struct {
	*basecmd
	file bool
	dir  bool
}

func (c *configCmd) run(cmd *cobra.Command, args []string) error {
	if c.file {
		fmt.Println(config.File())
		return nil
	}
	if c.dir {
		fmt.Println(config.Folder())
		return nil
	}
	return c.cmd.Usage()
}

func newConfigCmd() cliCommand {
	c := &configCmd{file: false}
	c.basecmd = newVerboseBaseCommand("config", "Configure apizza", c.run)
	c.cmd.Long = `The 'config' command is used for accessing the .apizza config file
in your home directory. Feel free to edit the .apizza json file
by hand or use the 'config' command.

ex. 'apizza config get name' or 'apizza config set name=<your name>'`

	c.cmd.Flags().BoolVarP(&c.file, "file", "f", c.file, "show the path to the config.json file")
	c.cmd.Flags().BoolVarP(&c.dir, "dir", "d", c.dir, "show the apizza config directory path")
	c.cmd.PersistentFlags().MarkHidden("address")
	c.cmd.PersistentFlags().MarkHidden("service")
	c.cmd.InheritedFlags().MarkHidden("address")
	return c
}

type configSetCmd struct {
	*basecmd
}

func (c *configSetCmd) run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Error: no variable given")
	}

	for _, arg := range args {
		keys := strings.Split(arg, "=")
		err := cfg.Set(keys[0], keys[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func newConfigSet() cliCommand {
	c := &configSetCmd{}
	c.basecmd = newBaseCommand(
		"set",
		"change variables in the config file",
		c.run,
	)
	return c
}

type configGetCmd struct {
	*basecmd
}

func (c *configGetCmd) run(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("Error: no variable given")
	}

	// add a flag '--all' that prints the contents of the config file
	for _, arg := range args {
		v := cfg.Get(arg)
		if v == nil {
			return fmt.Errorf("Error: cannot find %s", arg)
		}
		fmt.Println(v)
	}
	return nil
}

func newConfigGet() cliCommand {
	c := &configGetCmd{}
	c.basecmd = newBaseCommand(
		"get",
		"print the specified config variable to screen",
		c.run,
	)
	return c
}
