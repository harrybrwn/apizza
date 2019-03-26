package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	homedir "github.com/mitchellh/go-homedir"
)

var (
	cfg configfile
)

// SetConfig sets the config file and also runs through the configuration
// setup process.
func SetConfig(foldername string, c Config) error {
	if cfg.file != "" {
		return errors.New("cannot set multiple config files")
	}
	dir := getdir(foldername)

	cfg = configfile{
		conf: c,
		dir:  dir,
		file: filepath.Join(dir, "config.json"),
	}

	if !cfg.exists() {
		if err := os.Mkdir(cfg.dir, 0700); err != nil {
			return err
		}
		fmt.Printf("setting up config file at %s\n", cfg.file)
		cfg.setup()
	}
	return cfg.init()
}

// SetNonFileConfig sets a configuration struct without creating a file.
func SetNonFileConfig(c Config) error {
	cfg = configfile{
		conf: c,
		dir:  "",
		file: "",
	}
	t := reflect.ValueOf(c).Elem()
	autogen := emptyJSONConfig(t.Type(), 0)
	return json.Unmarshal([]byte(autogen), c)
}

type configfile struct {
	conf Config
	file string
	dir  string
}

func (c *configfile) save() error {
	raw, err := json.MarshalIndent(c.conf, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.file, raw, 0644)
}

func (c *configfile) reset() error {
	err := os.Remove(c.file)
	if err != nil {
		return err
	}
	return setup(c.file, c.conf)
}

func (c *configfile) setup() error {
	return setup(c.file, c.conf)
}

func (c *configfile) init() error {
	b, err := ioutil.ReadFile(c.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, c.conf)
}

func (c *configfile) exists() bool {
	_, dirErr := os.Stat(c.dir)
	_, fileErr := os.Stat(c.file)
	return !os.IsNotExist(dirErr) && !os.IsNotExist(fileErr)
}

// Object returns the configuration struct passes to SetConfig.
func Object() Config {
	return cfg.conf
}

// Get returns the value at a key for the config struct passes into SetConfig
func Get(key string) interface{} {
	return GetField(cfg.conf, key)
}

// GetString returns the config key value as a string.
func GetString(key string) string {
	return GetField(cfg.conf, key).(string)
}

// GetInt returns the config key value as an integer.
func GetInt(key string) int {
	return GetField(cfg.conf, key).(int)
}

// GetFloat returns the config key value as a float.
func GetFloat(key string) float64 {
	return GetField(cfg.conf, key).(float64)
}

// Set will set a value at a key for the config struct passed to SetConfig
func Set(key string, val interface{}) error {
	return SetField(cfg.conf, key, val)
}

// Folder returns the path to the folder that was set in SetConfig
func Folder() string {
	return cfg.dir
}

// File returns the config file
func File() string {
	return cfg.file
}

// Save saves the config file
func Save() error {
	return cfg.save()
}

// Reset deletes the config file and runs through the setup process
func Reset() error {
	return cfg.reset()
}

func setup(fname string, obj interface{}) error {
	f, err := os.Create(fname)
	defer f.Close()
	if err != nil {
		return err
	}
	t := reflect.ValueOf(obj).Elem()
	autogen := emptyJSONConfig(t.Type(), 0)
	_, err = f.Write([]byte(autogen))
	return err
}

func emptyJSONConfig(t reflect.Type, level int) string {
	spacer := "    "
	rawcnfg := "{\n"

	for i := 0; i < t.NumField(); i++ {
		comma := ",\n"
		end := "},\n"
		if i == t.NumField()-1 {
			comma = "\n"
			end = "}\n"
		}

		f := t.Field(i)
		for i := 0; i < level; i++ {
			rawcnfg += spacer
		}
		rawcnfg += fmt.Sprintf("%s\"%s\": ", spacer, f.Name)

		// look for a default value
		if deflt, ok := f.Tag.Lookup("default"); ok {
			if f.Type.Kind() == reflect.String {
				deflt = fmt.Sprintf("\"%s\"", deflt)
			}
			rawcnfg += deflt + comma
			continue
		}

		// add an empty value
		switch f.Type.Kind() {
		case reflect.Struct:
			rawcnfg += emptyJSONConfig(f.Type, level+1) + end
		case reflect.Int:
			rawcnfg += "0" + comma
		case reflect.Float64:
			rawcnfg += "0.0" + comma
		case reflect.String:
			rawcnfg += fmt.Sprintf("\"%s\"%s", "", comma)
		default:
			rawcnfg += "null" + comma
		}
	}
	for k := 0; k < level; k++ {
		rawcnfg += spacer
	}

	if level == 0 {
		return rawcnfg + "}"
	}
	return rawcnfg
}

func getdir(fname string) string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, fname)
}

func rightLable(key string, field reflect.StructField) bool {
	if key == field.Name || key == field.Tag.Get("config") {
		return true
	}
	return false
}
