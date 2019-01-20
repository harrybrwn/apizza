package config

import (
	"encoding/json"
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
)

// Folder returns the path to the folder that was set in SetConfig
func Folder() string {
	return cfgFolder
}

// SetConfig sets the config file and also runs through the configuration
// setup process.
func SetConfig(foldername string, cfg interface{}) error {
	cfgFolder = getfile(foldername)
	cfgFile = filepath.Join(cfgFolder, "config.json")

	if !Exists() {
		os.Mkdir(cfgFolder, os.ModeDir)
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
		return write(cfgFile, raw)
	}

	reset = func() {
		os.Remove(cfgFile)
		setup(cfgFile, cfg)
	}

	return nil
}

// SaveConfig saves the config file
func SaveConfig() error {
	return save()
}

// ResetConfig deletes the config file and runs through the setup process
func ResetConfig() error {
	reset()
	return save()
}

func Exists() bool {
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
	autogen := []byte(emptyConfig(t.Type(), 0))
	f.Write(autogen)
	return nil
}

func emptyConfig(t reflect.Type, level int) string {
	spacer := "    "
	empty := "{\n"

	for i := 0; i < t.NumField(); i++ {
		comma := ",\n"
		if i == t.NumField()-1 {
			comma = "\n"
		}
		f := t.Field(i)
		for i := 0; i < level; i++ {
			empty += spacer
		}
		empty += fmt.Sprintf("%s\"%s\": ", spacer, f.Name)

		if deflt := f.Tag.Get("default"); deflt != "" {
			empty += deflt + comma
			continue
		}
		switch f.Type.Kind() {
		case reflect.Struct:
			empty += emptyConfig(f.Type, level+1)
		case reflect.Int:
			empty += "0" + comma
		case reflect.String:
			empty += "\"\"" + comma
		default:
			empty += "null" + comma
		}
	}

	for i := 0; i < level; i++ {
		empty += spacer
	}
	if level > 0 {
		return empty + "},\n"
	}
	return empty + "}"
}

func getfile(fname string) string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, fname)
}

func write(file string, b []byte) error {
	return ioutil.WriteFile(file, b, 0644)
}
