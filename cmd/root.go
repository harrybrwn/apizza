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

	"apizza/dawg"
	"apizza/pkg/config"

	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "apizza",
	Short: "Dominos pizza from the command line.",
	Run: func(cmd *cobra.Command, args []string) {
		if test, err := cmd.Flags().GetBool("test"); test && err == nil {
			// print("nothing is being tested\n")
			dbTest()
			fmt.Println(db.Path())
			// fmt.Printf("%+v\n", db.Info())
			// fmt.Printf("%+v\n", db.Stats())
		} else {
			cmd.Usage()
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return config.Save()
	},
}

var (
	cfg = &Config{}

	addr  *dawg.Address
	menu  *dawg.Menu
	store *dawg.Store
	db    *bolt.DB
)

// Execute runs the root command
func Execute() {
	err := initDatabase()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Config is the configuration struct
type Config struct {
	Name    string `config:"name"`
	Email   string `config:"email"`
	Address struct {
		Street       string `config:"street"`
		Streetname   string `config:"streetname"`
		Streetnumber string `config:"streetnumber"`
		City         string `config:"city"`
		State        string `config:"state"`
		Zip          string `config:"zip"`
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

func init() {
	var err error
	err = config.SetConfig(".apizza", cfg)
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().String("address", "", "use a specific address")
	rootCmd.PersistentFlags().String("service", "Delivery", "select a Dominos service, either 'Delivery' or 'Carryout'")
	rootCmd.PersistentFlags().String("store-id", "", "store id used for all endpoint calls")

	rootCmd.Flags().BoolP("test", "t", false, "testing flag")
	rootCmd.Flags().MarkHidden("test")

	address, err := rootCmd.PersistentFlags().GetString("address")
	if address == "" {
		addr = &dawg.Address{
			Street: cfg.Address.Street,
			City:   cfg.Address.City,
			State:  cfg.Address.State,
			Zip:    cfg.Address.Zip,
		}
	} else if address != "" && err != nil {
		addr = dawg.ParseAddress(address)
	}
}
