package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
)

func TestUtils(t *testing.T) {
	p := "/this/is/a/test/path.test"
	if filename(p) != "path" {
		t.Errorf("filename did not get the right name given: '%s'", p)
	}
	p = `C:\Testing\a\windows\path.wintest`
	if filename(p) != "path" {
		t.Errorf("filename did not get the right name given: '%s'", p)
	}
	if filename("notapath.test") != "notapath" {
		t.Error("filename didn't work for just the name of a file")
	}
}

func TestGetDB(t *testing.T) {
	db, err := GetDB(temp())
	if err != nil {
		t.Error(err)
	}
	if db == nil {
		t.Error("GetDB returned a 'nil' value DataBase")
	}
	if db.Path != db.db.Path() {
		t.Error("the path tracked by the inner database was different than the wrapper.")
	}

	if err := db.Close(); err != nil {
		t.Error("didn't close db:", err)
	}
	err = db.Destroy()
	if err != nil {
		t.Error("Error in deleting the database:", err)
	}
}

func TestGetDB_ExpectedErr(t *testing.T) {
	_, err := GetDB("", "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDB_Put(t *testing.T) {
	dir, fname := temp()
	db, err := GetDB(dir, fname)
	if err != nil || db == nil {
		t.Fatal("bad db creation:", err)
	}
	expected := []byte("this is a test")
	err = db.Put("test_val", expected)
	if err != nil {
		t.Error("problem in Put")
	}
	badkey := "nothere"
	var badkeyExists bool
	err = db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(fname))
		content := b.Get([]byte("test_val"))
		if string(content) != string(expected) {
			t.Error("the contents was not found in the inner database")
		}

		if b.Get([]byte(badkey)) == nil {
			badkeyExists = false
		} else {
			badkeyExists = true
		}
		return nil
	})
	if err != nil {
		t.Error("error on boltDB's end")
	}

	if db.Exists(badkey) != badkeyExists {
		t.Error("db.Exists does not match reality")
	}
	if !db.Exists("test_val") {
		t.Error("the 'test_val' key should exist")
	}

	if err := db.Close(); err != nil {
		t.Error("didn't close db:", err)
	}
}

func TestDB_Get(t *testing.T) {
	dir, fname := temp()
	db, err := GetDB(dir, fname)
	if err != nil || db == nil {
		t.Fatal("bad db creation")
	}
	testval := []byte("testing value")

	err = db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(fname))
		return b.Put([]byte("test"), testval)
	})
	if err != nil {
		t.Error("error on boltDB's end")
	}

	val, err := db.Get("test")
	if err != nil {
		t.Error("returned error:", err)
	}
	if string(val) != string(testval) {
		t.Error("returned wrong value")
	}
	if err := db.Close(); err != nil {
		t.Error("didn't close db:", err)
	}
}

func temp() (string, string) {
	// f, err := ioutil.TempFile("", "apizza-")
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	name := f.Name()
	return filepath.Dir(name), filepath.Base(name)
}
