package cli

import (
	"errors"

	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/pkg/config"
)

// Config is the configuration struct
type Config struct {
	Name               string      `config:"name" json:"name"`
	Email              string      `config:"email" json:"email"`
	Phone              string      `config:"phone" json:"phone"`
	Address            obj.Address `config:"address" json:"address"`
	DefaultAddressName string      `config:"default-address-name" json:"default-address-name"`
	Card               struct {
		Number     string `config:"number" json:"number"`
		Expiration string `config:"expiration" json:"expiration"`
	} `config:"card" json:"card"`
	Service string `config:"service" default:"Delivery" json:"service"`
}

// Get a config variable
func (c *Config) Get(key string) interface{} {
	return config.GetField(c, key)
}

// Set a config variable
func (c *Config) Set(key string, val interface{}) error {
	if config.FieldName(c, key) == "Service" {
		if val != "Delivery" && val != "Carryout" {
			return errors.New("service must be either 'Delivery' or 'Carryout'")
		}
	}
	return config.SetField(c, key, val)
}
