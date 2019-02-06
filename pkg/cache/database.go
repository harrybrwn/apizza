package cache

import (
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

// DataBase is a wrapper struct for boltdb key-value pair database.
type DataBase struct {
	Path       string
	Bucketname string
	db         *bolt.DB
}

var (
	bucketname = "apizza"
)

// GetDB returns an initialized DataBase with either a brand new boltdb or and
// existing one.
func GetDB(path, name string) *DataBase {
	dir := filepath.Join(path, "cache", name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModeDir)
	}
	boltdb, err := bolt.Open(dir, 0600, nil)
	if err != nil {
		panic(err)
	}
	return &DataBase{Path: dir, db: boltdb, Bucketname: bucketname}
}

// Put stores bytes the database
func (db *DataBase) Put(key string, val []byte) error {
	return nil
}

// Get will retrieve the value given a key
func (db *DataBase) Get(key string) ([]byte, error) {
	var err error
	var raw []byte

	err = db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Bucketname))
		raw = b.Get([]byte(key))
		return nil
	})
	return raw, err
}

func (db *DataBase) view(f func(*bolt.Tx) error) error {
	return db.db.View(f)
}

// Close will close the DataBases inner bolt.DB
func (db *DataBase) Close() error {
	return db.db.Close()
}
