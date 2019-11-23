package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTempFile(t *testing.T) {
	filenames := []string{TempFile(), NamedTempFile("test_prefix", "test_suffix")}
	for _, fname := range filenames {
		if fname == "" {
			t.Error("empty file name")
		}
		tmp := os.TempDir()
		if tmp[len(tmp)-1] == '/' {
			tmp = tmp[:len(tmp)-1]
		}
		if filepath.Dir(fname) != tmp {
			t.Errorf("wrong temp directory; got %s, want %s", filepath.Dir(fname), tmp)
		}
		_, err := os.Open(fname)
		if os.IsExist(err) {
			t.Error("temp file should not exist")
		}
	}
}

func TestTempDir(t *testing.T) {
	dir := TempDir()
	if dir == "" {
		t.Error("bad temp dir")
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error(dir, "should exist")
	}
	if err := os.Remove(dir); err != nil {
		t.Error(err)
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

func TestComparisons(t *testing.T) {
	tcases := []string{"testing", "testing this string"}
	for _, tc := range tcases {
		Compare(t, tc, tc)
		CompareV(t, tc, tc)
		CompareOutput(t, tc, func() { fmt.Print(tc) })
		CompareV(&testing.T{}, "going_to"+tc+"failcompairison", tc)
	}
}
