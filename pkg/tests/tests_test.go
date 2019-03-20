package tests

import (
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
