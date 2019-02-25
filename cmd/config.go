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

	// "apizza/pkg/config"

	"github.com/spf13/cobra"
)

type configCmd struct {
	*basecmd
}

func newConfigCmd() cliCommand {
	c := &configCmd{}
	c.basecmd = &basecmd{cmd: &cobra.Command{
		Use:   "config",
		Short: "Configure apizza",
		Long: `The 'config' command is used for accessing the .apizza config file
	in your home directory. Feel free to edit the .apizza json file
	by hand or use the 'config' command.
	
	ex. 'apizza config get <variable>' or 'apizza config set name=<your name>'`,
		RunE: c.run,
	}}

	c.AddCmd(
		newConfigSet(),
		newConfigGet(),
	)
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
	c := &configCmd{}
	c.basecmd = newSilentBaseCommand(
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
	c.basecmd = newSilentBaseCommand(
		"get",
		"print the specified config variable to screen",
		c.run,
	)
	return c
}
