package config

import (
	"fmt"
	"reflect"
	"strings"
)

type configuration interface {
	Get(string) interface{}
	Set(string, interface{}) error
}

// Get is a helper funtion that finds the field of the config object that
// matches the key and then returns the value.
func Get(config configuration, key string) interface{} {
	value := reflect.ValueOf(config).Elem()
	v := find(key, value)
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

// Set is a helper funtion that binds a variable to the fiels that
// matches the key argument
func Set(config configuration, key string, val interface{}) error {
	v := reflect.ValueOf(config).Elem()
    field := find(key, v)
	if !field.IsValid() {
		return fmt.Errorf("cannot find %s", key)
	}
    switch reflect.TypeOf(val).Kind() {
    case reflect.String:
        field.SetString(val.(string))
    case reflect.Int:
        field.SetInt(val.(int64))
    case reflect.Struct:
		// fmt.Println("structs haven't been figured out yet in config.Set")
        return fmt.Errorf("structs haven't been figured out yet in config.Set")
    }
	return nil
}

func find(key string, val reflect.Value) reflect.Value {
    keys := strings.Split(key, ".")
    t := val.Type()
    for i := 0; i < val.NumField(); i++ {
		tfield := t.Field(i)
        if tfield.Name == keys[0] || tfield.Tag.Get("config") == keys[0] {
            rtVal := val.Field(i)
            if rtVal.Kind() == reflect.Struct {
				if len(keys) > 1 {
					return find(keys[1], rtVal)
				} else if len(keys) == 1 {
					return rtVal
				}
            }
            return rtVal
        }
    }
    return reflect.ValueOf(nil)
}
