// Package tests is a package for internal testing in the apizza program.
package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/pkg/errs"
)

// Compare two strings and fail the test with an error message if they are not
// the same.
func Compare(t *testing.T, got, expected string) {
	CompareCallDepth(t, got, expected, 2)
}

// CompareV compairs strings verbosly.
func CompareV(t *testing.T, got, expected string) {
	CompareCallDepth(t, got, expected, 2)
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

// CompareOutput will redirect stdout and compair it to the expected string.
func CompareOutput(t *testing.T, expected string, f func()) {
	stdout, stderr := os.Stdout, os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stdout = w
	os.Stderr = w

	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout, os.Stderr = stdout, stderr

	CompareCallDepth(t, buf.String(), expected, 2)
}

// CaptureOutput from a function
func CaptureOutput(f func()) (*bytes.Buffer, error) {
	stdout, stderr := os.Stdout, os.Stderr
	r, w, err := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	f()
	err = errs.Pair(err, w.Close())
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	if err != nil {
		return nil, err
	}
	os.Stdout, os.Stderr = stdout, stderr
	return buf, nil
}

// CompareCallDepth compares two strings. depth is the function call depth
func CompareCallDepth(t *testing.T, got, exp string, depth int) {
	got = strings.Replace(got, " ", "_", -1)
	exp = strings.Replace(exp, " ", "_", -1)
	msg := fmt.Sprintf("wrong output!\n\ngot:\n'%s'\nexpected:\n'%s'\n", got, exp)
	if got != exp {
		_, file, line, ok := runtime.Caller(depth)
		if ok && depth > 0 {
			msg = fmt.Sprintf("\nComparison Failure - %s:%d\n\n%s", file, line, msg)
		}
		if len(got) != len(exp) {
			msg += fmt.Sprintf("\nthey are different lengths too: %d %d", len(got), len(exp))
		}
		t.Errorf(msg)
	}
}
