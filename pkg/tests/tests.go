package tests

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var rand uint32
var mu sync.Mutex

// TempFile returns the path to a temporary file that does not exits.
// Tempfile is essentially a random filename generator.
func TempFile() string {
	return randFile(os.TempDir(), "", "")
}

// NamedTempFile gives the opton to give a named temporary file with
//
// will return "/path/to/temp_dir/<prefix>random_filename<suffix>"
func NamedTempFile(prefix, suffix string) string {
	return randFile(os.TempDir(), prefix, suffix)
}

// Parts of this function came from the Go standard library io/ioutil/tempfile.go
func randFile(dir string, prefix, suffix string) (fname string) {
	var (
		f   *os.File
		err error
	)
	for i := 0; i < 1000; i++ {
		fname = filepath.Join(dir, prefix+nextRandom()+suffix)
		f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			continue
		}
		break
	}
	if f.Close() != nil || os.Remove(fname) != nil {
		panic("could not remove temp file after creation")
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
