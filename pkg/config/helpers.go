package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

var (
	errWrongType = errors.New("wrong type")
)

// Config represents a config object that can act like a series of
// key-value pairs by using Get and Set methods.
type Config interface {
	Get(string) interface{}
	Set(string, interface{}) error
}

// GetField is a helper function that finds the field of the config object that
// matches the key and then returns the value. This function is meant to be
// wrapped by a config object. GetField will return nil if it does not find the key.
//
// Example:
// 	type MyConfig struct {}
//
// 	func (c *MyConfig) Get(key string) interface{} { return config.GetField(c, key) }
//
// note: this will only work if the struct implements the Config interface.
func GetField(config Config, key string) interface{} {
	value := reflect.ValueOf(config).Elem()
	_, _, val := find(value, strings.Split(key, "."))
	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int:
		return val.Int()
	case reflect.Float64:
		return val.Float()
	case reflect.Float32:
		return val.Float()
	case reflect.Struct:
		return val
	default:
		return nil
	}
}

// SetField is a helper function that binds a variable to the field that
// matches the key argument. This function is meant to be
// wrapped by a config object.
//
// Example:
// 	type MyConfig struct {}
//
// 	func (c *MyConfig) Get(key string, val interface{}) error { return config.SetField(c, key, val) }
//
// note: this will only work if the struct implements the Config interface.
func SetField(config Config, key string, val interface{}) error {
	v := reflect.ValueOf(config).Elem()
	_, _, field := find(v, strings.Split(key, "."))
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
	case float32, float64:
		if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
			field.SetFloat(val.(float64))
		} else {
			return fmt.Errorf("config.Set: wrong type")
		}
	default:
		return fmt.Errorf("structs haven't been figured out yet for config.Set")
	}
	return nil
}

// IsField will return true is the Config argumetn has either a field or a
// config tag that coressponds with the key given.
func IsField(c Config, key string) bool {
	_, _, val := find(reflect.ValueOf(c).Elem(), strings.Split(key, "."))
	return val != reflect.ValueOf(nil)
}

// FieldName will return the name of the struct field based on a config key.
func FieldName(c Config, key string) string {
	name, _, _ := find(reflect.ValueOf(c).Elem(), strings.Split(key, "."))
	return name
}

// PrintAll prints out the config struct.
func PrintAll(config interface{}) error {
	return FprintAll(os.Stdout, config)
}

// FprintAll prints the config to an io.Writer interface.
func FprintAll(w io.Writer, config interface{}) error {
	_, err := fmt.Fprint(
		w, visitAll(reflect.ValueOf(config).Elem(), 0, DefaultFormatter),
	)
	return err
}

func find(val reflect.Value, keys []string) (string, *reflect.StructField, reflect.Value) {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {

		if rightLable(keys[0], typ.Field(i)) {
			typFld := typ.Field(i)

			if len(keys) > 1 {
				name, sf, v := find(val.Field(i), keys[1:])
				return fmt.Sprintf("%s.%s", typFld.Name, name), sf, v
			} else if len(keys) == 1 {
				return typFld.Name, &typFld, val.Field(i)
			}
			return typFld.Name, &typFld, val.Field(i)
		}
	}
	return "", nil, reflect.ValueOf(nil)
}

// Formatter is a struct holding a collection of formatting functions.
type Formatter struct {
	// KeyVal handles key-value pairs
	KeyValFormat func(string, string) string

	// StructFormat handles key-value pairs where the value is a struct.
	StructFormat func(string, string) string

	// ValueHandler handles reflection values given the correct field number.
	ValueHandler func(reflect.Value, int) string

	// TabSize is the length of tab used for formatting.
	TabSize int

	indentLength int
}

func visitAll(val reflect.Value, depth int, fmtr Formatter) string {
	var (
		display string
		name    string
		ok      bool
	)

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		name, ok = typ.Field(i).Tag.Lookup("config")
		if !ok {
			name = typ.Field(i).Name
		}

		fieldVal := val.Field(i)
		if fieldVal.Kind() == reflect.Struct {
			display += fmtr.StructFormat(name, visitAll(fieldVal, depth+1, fmtr))
		} else if fieldVal.Kind() == reflect.Interface && fieldVal.IsNil() {
			display += fmtr.KeyValFormat(name, "null")
		} else {
			display += fmt.Sprintf("%s%s",
				strings.Repeat(" ", depth*fmtr.TabSize), fmtr.KeyValFormat(name, fmtr.ValueHandler(val, i)))
		}
	}
	return display
}

// DefaultFormatter is the default formatter object.
var DefaultFormatter = Formatter{
	KeyValFormat: func(k, v string) string { return fmt.Sprintf("%s: %s\n", k, v) },
	StructFormat: func(k, v string) string { return fmt.Sprintf("%s:\n%s", k, v) },
	ValueHandler: func(v reflect.Value, i int) string {
		field := v.Field(i)
		var val string

		switch field.Kind() {
		case reflect.String:
			val = fmt.Sprintf("\"%s\"", field.String())
		default:
			val = fmt.Sprintf("%v", field)
		}
		return val
	},
	TabSize: 2,
}
