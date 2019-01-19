package cache

import (
	"os"
	"path"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

func SetCache(filename string) error {
	cacheFolder := path.Join
	if _, err := os.Stat(cfgFolder); os.IsNotExist(err) {
		// make folder
	}
}

func getfile(fname string) string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, fname)
}
