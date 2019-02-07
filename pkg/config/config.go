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
)

// SetConfig sets the config file and also runs through the configuration
// setup process.
func SetConfig(foldername string, cfg interface{}) error {
	if cfgFile != "" {
		return errors.New("cannot set multiple config files")
	}
	cfgFolder = getdir(foldername)
	cfgFile = filepath.Join(cfgFolder, "config.json")

	if !exists() {
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
		return ioutil.WriteFile(cfgFile, raw, 0644)
	}

	reset = func() {
		os.Remove(cfgFile)
		setup(cfgFile, cfg)
	}

	return nil
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

	nfields := t.NumField()
	for i := 0; i < nfields; i++ {
		comma := ",\n"
		end := "},\n"
		if i == nfields-1 {
			comma = "\n"
			end = "}\n"
		}

		f := t.Field(i)
		for i := 0; i < level; i++ {
			rawcnfg += spacer
		}
		rawcnfg += fmt.Sprintf("%s\"%s\": ", spacer, f.Name)

		if deflt := f.Tag.Get("default"); deflt != "" {
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
