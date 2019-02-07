package cache

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

// DataBase is a wrapper struct for boltdb key-value pair database.
type DataBase struct {
	Path         string
	Bucketname   string
	bucketExists bool
	db           *bolt.DB
}

// GetDB returns an initialized DataBase with either a brand new boltdb or and
// existing one.
func GetDB(dbdir, dbname string) (*DataBase, error) {
	name := filename(dbname)
	boltdb, err := bolt.Open(filepath.Join(dbdir, dbname), 0600, nil)
	if err != nil {
		return nil, err
	}

	err = boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	db := &DataBase{
		Path:       filepath.Join(dbdir, dbname),
		Bucketname: name,
		db:         boltdb,
	}
	return db, err
}

// Put stores bytes the database
func (db *DataBase) Put(key string, val []byte) error {
	return db.update(func(b *bolt.Bucket) error {
		return b.Put([]byte(key), val)
	})
}

// Get will retrieve the value given a key
func (db *DataBase) Get(key string) ([]byte, error) {
	var err error
	var raw []byte

	err = db.view(func(b *bolt.Bucket) error {
		raw = b.Get([]byte(key))
		return nil
	})
	return raw, err
}

// Exists will return true if the key supplied has data associated with it.
func (db *DataBase) Exists(key string) bool {
	var exists bool
	db.view(func(b *bolt.Bucket) error {
		data := b.Get([]byte(key))

		if data == nil {
			exists = false
		} else {
			exists = true
		}
		return nil
	})
	return exists
}

// Close will close the DataBases inner bolt.DB
func (db *DataBase) Close() error {
	return db.db.Close()
}

// Destroy will close the database and completly delete the database file.
func (db *DataBase) Destroy() error {
	err := db.Close()
	if err != nil {
		return err
	}
	return os.Remove(db.Path)
}

func (db *DataBase) view(fn func(*bolt.Bucket) error) error {
	return db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.Bucketname))
		return fn(bucket)
	})
}

func (db *DataBase) update(fn func(*bolt.Bucket) error) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.Bucketname))
		return fn(bucket)
	})
}

func filename(file string) string {
	fname := filepath.Base(file)
	return strings.TrimSuffix(fname, filepath.Ext(fname))
}
