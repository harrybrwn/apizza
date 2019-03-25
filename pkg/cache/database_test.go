package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestUtils(t *testing.T) {
	p := "/this/is/a/test/path.test"
	if filename(p) != "path" {
		t.Errorf("filename did not get the right name given: '%s'", p)
	}
	if filename("notapath.test") != "notapath" {
		t.Error("filename didn't work for just the name of a file")
	}
}

func TestGetDB(t *testing.T) {
	db, err := GetDB(tests.TempFile())
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
	_, err := GetDB("")
	if err == nil {
		t.Error("expected error")
	}
}
func TestDB_Put(t *testing.T) {
	dbfname := tests.TempFile()
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
}

func TestDB_Get(t *testing.T) {
	dbfname := tests.TempFile()
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
}

func TestTimeStamp(t *testing.T) {
	db, err := GetDB(tests.TempFile())
	if err != nil || db == nil {
		t.Fatal("bad db creation")
	}

	stamp, err := db.TimeStamp("test")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second / 4)
	tdiff := time.Since(stamp)
	if time.Millisecond*240 > tdiff || tdiff > time.Millisecond*260 {
		t.Error("time stamp is not in the right range")
	}

	time.Sleep(time.Second / 4)
	stamp2, err := db.TimeStamp("test")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
	t1, t2 := time.Since(stamp), time.Since(stamp2)
	if t1 > t2 {
		t.Error("time one shouldn't be bigger than t1")
	}
	if t2-t1 > time.Second {
		t.Error("time difference is way too big")
	}
	if err := db.ResetTimeStamp("test"); err != nil {
		t.Error(err)
	}
	stampRe, err := db.TimeStamp("test")
	if err != nil {
		t.Error(err)
	}
	if time.Since(stampRe) > t2 {
		t.Error("timestamp didn't reset")
	}
}

func TestAutoTimeStamp(t *testing.T) {
	db, err := GetDB(tests.TempFile())
	if err != nil || db == nil {
		t.Fatal("bad db creation")
	}
	if err = db.AutoTimeStamp("test", time.Second/10,
		func() error { return nil },
		func() error { return nil },
	); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second / 2)

	if err = db.AutoTimeStamp("test", time.Second/10,
		func() error { return nil },
		func() error { return nil },
	); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second / 2)

	if err = db.AutoTimeStamp("test", time.Second/10,
		func() error { return errors.New("this error should be raised") },
		func() error { return nil },
	); err == nil {
		t.Error("expected error from update func")
	}

	if err = db.AutoTimeStamp("test", time.Second*2,
		func() error { return nil },
		func() error { return errors.New("this error should be raised") },
	); err == nil {
		t.Error("expected error from notUpdate func")
	}

	if err = db.AutoTimeStamp("test", time.Second*2,
		func() error { return nil }, nil,
	); err != nil {
		t.Error("notUpdate passed as nil:", err)
	}

	if err = db.AutoTimeStamp("test", time.Second*2,
		func() error { return errors.New("update func shouldn't be run but was") },
		func() error { return nil },
	); err != nil {
		t.Error(err)
	}
}
