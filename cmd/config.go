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

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure apizza",
	Long: `The 'config' command is used for accessing the .apizza config file
in your home directory. Feel free to edit the .apizza json file
by hand or use the 'config' command.

ex. 'apizza config get <variable>' or 'apizza config set name=<your name>'`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <config var>",
	Short: "change variables in the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Error: no variable given")
		}

		for _, a := range args {
			keys := strings.Split(a, "=")
			err := cfg.Set(keys[0], keys[1])
			if err != nil {
				return err
			}
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

var configGetCmd = &cobra.Command{
	Use:   "get <config var>",
	Short: "print the specified config variable to screen",
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	configCmd.AddCommand(configSetCmd)

	configGetCmd.Flags().BoolP("test", "t", false, "testing stuff")
	configGetCmd.Flags().MarkHidden("test")

	configCmd.AddCommand(configGetCmd)

	rootCmd.AddCommand(configCmd)
}
