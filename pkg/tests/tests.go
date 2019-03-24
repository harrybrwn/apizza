// Package tests is a package for internal testing in the apizza program.
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
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

// WithTempFile is a test wrapper that accepts a function with the file
// and testing.T as arguments.
func WithTempFile(test func(string, *testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		test(TempFile(), t)
	}
}

// TempDir returns a temproary directory.
func TempDir() string {
	dir := randFile(os.TempDir(), "", "")
	if err := os.Mkdir(dir, 0777); err != nil {
		return ""
	}
	return dir
}

// Compare two strings and fail the test with an error message if they are not
// the same.
func Compare(t *testing.T, got, expected string) {
	got = strings.Replace(got, " ", "_", -1)
	expected = strings.Replace(expected, " ", "_", -1)
	msg := fmt.Sprintf("wrong output:\ngot:\n'%s'\nexpected:\n'%s'\n", got, expected)
	if got != expected {
		t.Errorf(msg)
	}
	if len(got) != len(expected) {
		t.Error("they are different lengths too", len(got), len(expected))
	}
}

// CompareV compairs strings verbosly.
func CompareV(t *testing.T, got, expected string) {
	Compare(t, got, expected)
	var min int
	if len(got) > len(expected) {
		min = len(expected)
	} else {
		min = len(got)
	}

	for i := 0; i < min; i++ {
		// fmt.Sprintf("'%s' == '%s'\n", string(got[i]), string(expected[i]))
		if got[i] != expected[i] {
			t.Errorf("char %d: '%s' == '%s'\n", i, string(got[i]), string(expected[i]))
		}
	}
}

// Parts of this function came from the Go standard library io/ioutil/tempfile.go
func randFile(dir string, prefix, suffix string) (fname string) {
	for i := 0; i < 1000; i++ {
		fname = filepath.Join(dir, prefix+nextRandom()+suffix)
		if _, err := os.Stat(fname); os.IsExist(err) {
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
