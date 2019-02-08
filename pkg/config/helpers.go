package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	errWrongType = errors.New("wrong type")
)

// Configuration represents a config object that can act like a series of
// key-value pairs by using Get and Set methods.
type Configuration interface {
	Get(string) interface{}
	Set(string, interface{}) error
}

// Get is a helper funtion that finds the field of the config object that
// matches the key and then returns the value. This function is meant to be
// wrapped by a config object.
//
// Example.
// type Config struct {}
// func (c *Config) Get(key string) interface{} {
//     return config.Get(c, key)
// }
// note: this will only work if the struct impliments the Configuration type.
func Get(config Configuration, key string) interface{} {
	value := reflect.ValueOf(config).Elem()
	v := find(value, key)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int:
		return v.Int()
	case reflect.Struct:
		return v
	default:
		return nil
	}
}

// Set is a helper funtion that binds a variable to the field that
// matches the key argument. This function is meant to be
// wrapped by a config object.
//
// Example.
// type Config struct {}
// func (c *Config) Get(key string, val interface{}) error {
//     return config.Set(c, key, val)
// }
// note: this will only work if the struct impliments the Configuration type.
func Set(config Configuration, key string, val interface{}) error {
	v := reflect.ValueOf(config).Elem()
	field := find(v, key)
	if !field.IsValid() {
		return fmt.Errorf("cannot find '%s'", key)
	}

	switch val.(type) {
	case string:
		if field.Kind() == reflect.String {
			field.SetString(val.(string))
		} else {
			return fmt.Errorf("config.Set: wrong type")
		}
	case int, int64, int32:
		if field.Kind() == reflect.Int {
			field.SetInt(val.(int64))
		} else {
			return fmt.Errorf("config.Set: wrong type")
		}
	default:
		return fmt.Errorf("structs haven't been figured out yet for config.Set")
	}
	return nil
}

func find(val reflect.Value, key string) reflect.Value {
	keys := strings.Split(key, ".")
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		tfield := t.Field(i)
		if tfield.Name == keys[0] || tfield.Tag.Get("config") == keys[0] {
			rtVal := val.Field(i)
			if rtVal.Kind() == reflect.Struct {
				if len(keys) > 1 {
					return find(rtVal, keys[1])
				} else if len(keys) == 1 {
					return rtVal
				}
			}
			return rtVal
		}
	}
	return reflect.ValueOf(nil)
}
