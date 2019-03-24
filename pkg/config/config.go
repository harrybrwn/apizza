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
	save      func() error
	reset     func()
	cfgFolder string
	cfgFile   string

	// Object of the config object that os passes to SetConfig
	Object Config

	conf Config
)

// SetConfig sets the config file and also runs through the configuration
// setup process.
func SetConfig(foldername string, cfg Config) error {
	if cfgFile != "" {
		return errors.New("cannot set multiple config files")
	}
	cfgFolder = getdir(foldername)
	cfgFile = filepath.Join(cfgFolder, "config.json")

	if !exists() {
		os.Mkdir(cfgFolder, 0700)
		fmt.Printf("setting up config file at %s\n", cfgFile)
		setup(cfgFile, cfg)
	}
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		return err
	}

	save = func() error {
		raw, err := json.MarshalIndent(cfg, "", "    ")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(cfgFile, raw, 0644)
	}

	reset = func() {
		os.Remove(cfgFile)
		setup(cfgFile, cfg)
	}
	conf = cfg
	return nil
}

// Get returns the value at a key for the config struct passes into SetConfig
func Get(key string) interface{} {
	return GetField(conf, key)
}

// GetString returns the config key value as a string.
func GetString(key string) string {
	return GetField(conf, key).(string)
}

// GetInt returns the config key value as an integer.
func GetInt(key string) int {
	return GetField(conf, key).(int)
}

// Set will set a value at a key for the config struct passed to SetConfig
func Set(key string, val interface{}) error {
	return SetField(conf, key, val)
}

// Folder returns the path to the folder that was set in SetConfig
func Folder() string {
	return cfgFolder
}

// File returns the config file
func File() string {
	return cfgFile
}

// Save saves the config file
func Save() error {
	return save()
}

// Reset deletes the config file and runs through the setup process
func Reset() error {
	reset()
	return save()
}

func exists() bool {
	_, err := os.Stat(cfgFolder)
	return !os.IsNotExist(err)
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

		if deflt, ok := f.Tag.Lookup("default"); ok {
			if f.Type.Kind() == reflect.String {
				deflt = fmt.Sprintf("\"%s\"", deflt)
			}
			rawcnfg += deflt + comma
			continue
		}

		switch f.Type.Kind() {
		case reflect.Struct:
			rawcnfg += emptyJSONConfig(f.Type, level+1) + end
		case reflect.Int:
			rawcnfg += "0" + comma
		case reflect.String:
			rawcnfg += "\"\"" + comma
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
