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

// CompareV compairs strings verbosely.
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

var currentTest *struct {
	name string
	t    *testing.T
} = nil

// PrintErrType is a switch for the Check method, so that it prints
// the error type on failure.
var PrintErrType bool = false

func nilcheck() {
	if currentTest == nil {
		panic("No testing.T registered; must call errs.InitHelpers(t) at test function start")
	}
}

// InitHelpers will set the err package testing.T variable for tests
func InitHelpers(t *testing.T) {
	initHelpers(t)
}

// ResetHelpers will set the current test to nil and make sure that
// no callers after it can use the testing.T object.
func ResetHelpers() {
	currentTest = nil
}

// Check will check to see that an error is nil, and cause an error if not
func Check(err error) {
	nilcheck()
	currentTest.t.Helper()
	if err != nil {
		if PrintErrType {
			currentTest.t.Errorf("%T %v\n", err, err)
		} else {
			currentTest.t.Errorf("%v\n", err)
		}
	}
}

// Exp will fail the test if the error is nil
func Exp(err error, vs ...interface{}) {
	nilcheck()
	currentTest.t.Helper()
	if err == nil {
		if len(vs) > 0 {
			msg := []interface{}{"expected an error; "}
			msg = append(msg, vs...)
			currentTest.t.Error(msg...)
		} else {
			currentTest.t.Error("expected an error; got <nil>")
		}
	}
}

// Fatal will fail and exit the test if the error is not nil.
func Fatal(err error) {
	nilcheck()
	currentTest.t.Helper()
	if err != nil {
		currentTest.t.Fatal(err)
	}
}

// StrEq will show an error message if a is not equal to b.
func StrEq(a, b string, fmt string, vs ...interface{}) {
	nilcheck()
	currentTest.t.Helper()
	if a != b {
		currentTest.t.Errorf(fmt+"\n", vs...)
	}
}

// NotNil is an assertion that the argument given is not nil.
// If the argument v is nil it will stop the test with t.Fatal().
func NotNil(v interface{}) {
	nilcheck()
	currentTest.t.Helper()
	if _, ok := v.(error); ok {
		currentTest.t.Log("tests warning: NotNil should not be used to check errors, it will call t.Fatal()")
	}
	if v == nil {
		currentTest.t.Fatalf("%T should not be nil\n", v)
	}
}
