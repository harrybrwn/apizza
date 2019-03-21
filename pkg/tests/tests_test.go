package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTempFile(t *testing.T) {
	filenames := []string{
		TempFile(),
		NamedTempFile("test_prefix", "test_suffix"),
	}

	for _, fname := range filenames {
		if fname == "" {
			t.Error("empty file name")
		}
		if filepath.Dir(fname) != os.TempDir() {
			t.Error("wrong temp directory")
		}
		_, err := os.Open(fname)
		if os.IsExist(err) {
			t.Error("temp file should not exist")
		}
	}
}

func TestRunner(t *testing.T) {
	output := &bytes.Buffer{}
	expected := `setting up tests...
t1
t2
closing tests.
`
	r := &Runner{
		T: t,
		Setup: func() {
			fmt.Fprintln(output, "setting up tests...")
		},
		Teardown: func() {
			fmt.Fprintln(output, "closing tests.")
		},
	}
	r.AddTest(
		func(t *testing.T) { fmt.Fprintln(output, "t1") },
		func(t *testing.T) { fmt.Fprintln(output, "t2") },
	)
	r.Run()
	if string(output.Bytes()) != expected {
		t.Error("output should be as expected")
	}
}

func TestNewRunner(t *testing.T) {
	var (
		setupRan    = false
		middleRan   = false
		teardownRan = false
	)

	r := NewRunner(t, func() { setupRan = true }, func() { teardownRan = true })
	if r.T == nil {
		t.Error("bad vars")
	}
	r.AddTest(func(*testing.T) { middleRan = true })
	r.Run()

	if !setupRan {
		t.Error("setup did not run")
	}
	if !middleRan {
		t.Error("test did not run")
	}
	if !teardownRan {
		t.Error("teardown did not run")
	}
}

func TestMatchString(t *testing.T) {
	m := matchString(matchStr)
	if ok, err := m.MatchString("", ""); err != nil {
		t.Error(err)
	} else if !ok {
		t.Error("should be true")
	}

	if err := m.StartCPUProfile(&bytes.Buffer{}); err == nil {
		t.Error("expected error")
	}
	if err := m.WriteProfileTo("", &bytes.Buffer{}, 0); err == nil {
		t.Error("expected error")
	}
	if err := m.StopTestLog(); err == nil {
		t.Error("expected error")
	}

	if m.ImportPath() != "" {
		t.Error("wrong path")
	}

	m.StartTestLog(&bytes.Buffer{})
	m.StopCPUProfile()
}

func TestTempDir(t *testing.T) {
	dir := TempDir()
	if dir == "" {
		t.Error("bad temp dir")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error(dir, "should exist")
	}
}
func TestWithTempFile(t *testing.T) {
	f := func(file string, innerT *testing.T) {
		if _, err := os.Stat(file); os.IsExist(err) {
			t.Error(file, "should not exist")
		}
	}

	t.Run("inner_test_WithTempFile", WithTempFile(f))
}

func TestWrap(t *testing.T) {
	r := &Runner{
		T:        t,
		Setup:    func() {},
		Teardown: func() {},
	}

	nilErrF := func() error { return nil }
	errF := func() error { return errors.New("test err") }

	r.AddTest(
		func(*testing.T) {},
		r.Wrap(func(*testing.T) {}, nilErrF, nilErrF),
	)
	r.Run()

	r.Reset()
	if len(r.tests) > 0 {
		t.Error("tests were not reset")
	}

	r.AddTest(
		func(*testing.T) {},
		r.Wrap(func(t *testing.T) { t.Skip("should be skipped") }, nilErrF, errF),
	)
	r.Run()
}
