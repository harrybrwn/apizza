package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

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
		return newError(err)
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
	autogen := []byte(emptyConfig(t.Type(), 0))
	f.Write(autogen)
	return nil
}

func emptyConfig(t reflect.Type, level int) string {
	spacer := "    "
	// spacer := "\t"
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
			rawcnfg += emptyConfig(f.Type, level+1) + end
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

func write(file string, b []byte) error {
	return ioutil.WriteFile(file, b, 0644)
}

// Error is an error object that convays more information about errors
// raised during configuration.
type Error struct {
	inner error
	fun   string
	line  int
	file  string
}

func newError(inner error) Error {
	fpcs := make([]uintptr, 2)
	runtime.Callers(2, fpcs)
	fun := runtime.FuncForPC(fpcs[0]).Name() + "()"
	_, file, line, _ := runtime.Caller(2)
	return Error{inner: inner, fun: fun, file: file, line: line}
}

func (ce Error) Error() string {
	return fmt.Sprintf("%s:%d %s\n%s", ce.file, ce.line, ce.fun,
		ce.inner.Error())
}
