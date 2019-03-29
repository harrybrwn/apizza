package cache

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	. "github.com/harrybrwn/apizza/pkg/tests"
)

func TestUtils(t *testing.T) {
	p := "/this/is/a/test/path.test"
	if filename(p) != "path" {
		t.Errorf("filename did not get the right name given: '%s'", p)
	}
	if filename("notapath.test") != "notapath" {
		t.Error("filename didn't work for just the name of a file")
	}
	if err := ensurePath(filepath.Join(os.TempDir(), "test_dir", "name")); err != nil {
		t.Error(err)
	}
}

func TestGetDB(t *testing.T) {
	db, err := GetDB(TempFile())
	if err != nil {
		t.Error(err)
	}
	if db == nil {
		t.Error("GetDB returned a 'nil' value DataBase")
	}
	if db.Path() != db.db.Path() {
		t.Error("the path tracked by the inner database was different than the wrapper.")
	}
	_, err = os.Stat(db.Path())
	if os.IsNotExist(err) {
		t.Error("path does not exist")
	}
	if err := db.Close(); err != nil {
		t.Error("didn't close db:", err)
	}
	if err = db.Destroy(); err != nil {
		t.Error("Error in deleting the database:", err)
	}
	db, err = GetDB(filepath.Join(TempFile(), "testdatabase"))
	if err != nil {
		t.Error(err)
	}
	if err = db.Destroy(); err != nil {
		t.Error(err)
	}
	db, err = GetDB("")
	if err == nil {
		t.Error("expected error")
	}
	if db != nil {
		t.Error("db should be nil")
	}
}

func TestDB_Put(t *testing.T) {
	dbfname := TempFile()
	fname := filename(dbfname)
	db, err := GetDB(dbfname)
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
	err = db.Put("yes", []byte("'yes' is a key that should exist."))
	if err != nil {
		t.Error(err)
	}
	if db.Exists("yes") == false {
		t.Error("should exist")
	}
	if db.Exists("no") == true {
		t.Error("shouldn't exist")
	}
	if err := db.Delete("yes"); err != nil {
		t.Error(err)
	}
	if db.Exists("yes") == true {
		t.Error("shouldn't exist")
	}
	all, err := db.Map()
	if err != nil {
		t.Error(err)
	}
	if all == nil {
		t.Error("got empty map")
	}
	if err := db.Close(); err != nil {
		t.Error("didn't close db:", err)
	}
	if err := db.Destroy(); err != nil {
		t.Error(err)
	}
}

func TestDB_Get(t *testing.T) {
	dbfname := TempFile()
	fname := filename(dbfname)
	db, err := GetDB(dbfname)
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
	if db.Exists("") {
		t.Error("should be false")
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
	if err := db.Destroy(); err != nil {
		t.Error(err)
	}
}

func TestBuckets(t *testing.T) {
	var err error
	tcases := []struct {
		buck string
		data []byte
	}{
		{buck: "bucket1", data: []byte("test data")},
		{buck: "testbuck2", data: []byte("just some stuff to use for tests")},
		{buck: "tb3", data: []byte("this is the part I hate about testing")},
	}
	key := "testdata"
	db, _ := GetDB(TempFile())
	boltdb := db.innerdb.db
	defbucket := "TestBucket"
	db.SetBucket(defbucket)
	for i, tc := range tcases {
		err = db.WithBucket(tc.buck).Put(key, tc.data)
		if err != nil {
			t.Error(err)
		}
		if string(db.defaultBucket) != defbucket {
			t.Error("bucket didn't reset")
		}
		err = boltdb.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(tc.buck))
			data := b.Get([]byte(key))
			if bytes.Compare(data, tc.data) != 0 {
				t.Errorf("didn't get the right data: test case %d", i)
			}
			return nil
		})
		if err = db.DeleteBucket(tc.buck); err != nil {
			t.Error(err)
		}
		if err != nil {
			t.Error(err)
		}
		err = boltdb.View(func(tx *bolt.Tx) error {
			if b := tx.Bucket([]byte(tc.buck)); b != nil {
				t.Error("bucket not deleted")
			}
			return err
		})
		if err != nil {
			t.Error(err)
		}
	}
	if err = db.Destroy(); err != nil {
		t.Error(err)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("should have paniced")
		}
	}()
	db.DeleteBucket(defbucket)
}
