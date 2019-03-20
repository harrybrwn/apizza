package tests

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func init() {
	// panic("haven't started writing this package")
}

var rand uint32
var mu sync.Mutex

// TempFile returns the path to a temporary file that does not exits.
func TempFile() string {
	return randFile(os.TempDir())
}

// Parts of this function came from the Go standard library io/ioutil/tempfile.go
func randFile(dir string) (fname string) {
	for i := 0; i < 1000; i++ {
		fname = filepath.Join(dir, nextRandom())
		_, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			continue
		}
		break
	}
	return fname
}

// This function came from the Go standard library io/ioutil/tempfile.go
func nextRandom() string {
	mu.Lock()
	r := rand
	if r == 0 {
		r = uint32(time.Now().UnixNano() + int64(os.Getpid()))
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	mu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}
