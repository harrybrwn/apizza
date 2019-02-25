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
	"apizza/dawg"
	"apizza/pkg/config"

	"github.com/boltdb/bolt"
)

var (
	cfg = &Config{}

	addr  *dawg.Address
	menu  *dawg.Menu
	store *dawg.Store
	db    *bolt.DB
)

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
