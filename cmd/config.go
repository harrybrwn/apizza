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

// var _ dawg.Address = (*address)(nil)

// type address struct {
// 	Street   string `config:"street"`
// 	CityName string `config:"cityname"`
// 	State    string `config:"state"`
// 	Zipcode  string `config:"zipcode"`
// }

// func (a *address) LineOne() string {
// 	return a.Street
// }

// func (a *address) StateCode() string {
// 	if strLen(a.State) == 2 {
// 		return strings.ToUpper(a.State)
// 	} else if len(a.State) == 0 {
// 		return ""
// 	}
// 	panic(fmt.Sprintf("bad statecode %s", a.State))
// }

// func (a *address) City() string {
// 	return a.CityName
// }

// func (a *address) Zip() string {
// 	if strings.Contains(a.Zipcode, " ") {
// 		panic(fmt.Sprintf("bad zipcode %s", a.Zipcode))
// 	}
// 	if strLen(a.Zipcode) == 5 {
// 		return a.Zipcode
// 	}
// 	panic(fmt.Sprintf("bad zipcode %s", a.Zipcode))
// }

// func addressStr(a dawg.Address) string {
// 	return addressStrIndent(a, 0)
// }

// func addressStrIndent(a dawg.Address, tablen int) string {
// 	var format string
// 	if strLen(a.StateCode()) == 0 {
// 		format = "%s\n%s%s, %s%s"
// 	} else {
// 		format = "%s\n%s%s, %s %s"
// 	}

// 	return fmt.Sprintf(format,
// 		a.LineOne(), spaces(tablen), a.City(), a.StateCode(), a.Zip())
// }

// func (a address) String() string {
// 	return addressStr(&a)
// }

type configCmd struct {
	*basecmd
	file       bool
	dir        bool
	resetCache bool
	getall     bool
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
	if c.resetCache {
		return os.Remove(filepath.Join(config.Folder(), "cache", "apizza.db"))
	}
	if c.getall {
		return cfg.printAll(c.output)
	}
	return c.cmd.Usage()
}

func newConfigCmd() cliCommand {
	c := &configCmd{file: false, dir: false, resetCache: false}
	c.basecmd = newVerboseBaseCommand("config", "Configure apizza", c.run)
	c.cmd.Long = `The 'config' command is used for accessing the .apizza config file
in your home directory. Feel free to edit the .apizza json file
by hand or use the 'config' command.

ex. 'apizza config get name' or 'apizza config set name=<your name>'`

	c.cmd.Flags().BoolVarP(&c.file, "file", "f", c.file, "show the path to the config.json file")
	c.cmd.Flags().BoolVarP(&c.dir, "dir", "d", c.dir, "show the apizza config directory path")
	c.cmd.Flags().BoolVarP(&c.resetCache, "reset-cache", "r", c.resetCache, "reset the database cache")
	c.cmd.Flags().BoolVar(&c.getall, "get-all", c.getall, "show all the contents of the config file")
	return c
}

type configSetCmd struct {
	*basecmd
}

func (c *configSetCmd) run(cmd *cobra.Command, args []string) error {
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
		return errors.New("no variable given")
	}

	for _, arg := range args {
		v := cfg.Get(arg)
		if v == nil {
			return fmt.Errorf("cannot find %s", arg)
		}
		fmt.Fprintln(c.output, v)
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
