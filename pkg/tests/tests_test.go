package tests

import (
	"bytes"
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

	r := &Runner{
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
	expected := `setting up tests...
t1
t2
closing tests.
`
	if string(output.Bytes()) != expected {
		t.Error("output should be as expected")
	}
	output.Reset()

	r = &Runner{
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
