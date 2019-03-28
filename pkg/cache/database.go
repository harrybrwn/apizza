package cache

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

// DB is an interface that represents a database wrapper.
type DB interface {
	Get(string) ([]byte, error)
	Put(string, []byte) error
	Delete(string) error
	Path() string
	Close() error
}

// DataBase is a wrapper struct for boltdb key-value pair database.
type DataBase struct {
	*innerdb
}

// GetDB returns an initialized DataBase. Will either create a brand new boltdb
// or open existing one.
func GetDB(dbfile string) (db *DataBase, err error) {
	err = ensurePath(dbfile)
	if err != nil {
		return nil, err
	}
	boltdb, err := bolt.Open(dbfile, 0777, nil)
	if err != nil {
		return nil, err
	}

	name := []byte(filename(dbfile))
	err = boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		return err
	})
	db = &DataBase{
		innerdb: &innerdb{
			defaultBucket: name,
			db:            boltdb,
			path:          dbfile,
			bucketHEAD:    name,
		},
	}
	return db, err
}

type innerdb struct {
	db            *bolt.DB
	defaultBucket []byte
	bucketHEAD    []byte
	path          string
}

// Put stores bytes the database
func (idb *innerdb) Put(key string, val []byte) error {
	return idb.update(func(b *bolt.Bucket) error {
		return b.Put([]byte(key), val)
	})
}

// Get will retrieve the value given a key
func (idb *innerdb) Get(key string) (raw []byte, err error) {
	err = idb.view(func(b *bolt.Bucket) error {
		raw = b.Get([]byte(key))
		return nil
	})
	return raw, err
}

// Delete removes the data for a specific key.
func (idb *innerdb) Delete(key string) error {
	return idb.update(func(b *bolt.Bucket) error {
		return b.Delete([]byte(key))
	})
}

// Exists will return true if the key supplied has data associated with it.
func (db *DataBase) Exists(key string) bool {
	return exists(db, key)
}

func exists(db internal, key string) (exists bool) {
	if err := db.view(func(b *bolt.Bucket) error {
		data := b.Get([]byte(key))

		if data == nil {
			exists = false
		} else {
			exists = true
		}
		return nil
	}); err != nil {
		return false
	}
	return exists
}

func (idb *innerdb) Path() string {
	return idb.path
}

// Close will close the DataBase's inner bolt.DB
func (idb *innerdb) Close() error {
	return idb.db.Close()
}

// Destroy will close the database and completely delete the database file.
func (db *DataBase) Destroy() error {
	err := db.Close()
	if err != nil {
		return err
	}
	return os.Remove(db.Path())
}

// Map returns a map of all the key-value pairs in the database.
func (db *DataBase) Map() (all map[string][]byte, err error) {
	all = map[string][]byte{}
	return all, db.view(func(b *bolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			all[string(k)] = v
			return nil
		})
	})
}

// WithBucket temporaraly sets the bucket to the string given and returns the
// database with the new bucket.
//
// The default bucket will be reset when the database calls Put, Get, Exists,
// Map, TimeStamp, and UpdateTS (any method that calls view or update internaly).
func (db *DataBase) WithBucket(bucket string) *DataBase {
	db.bucketHEAD = []byte(bucket)
	db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(db.bucketHEAD))
		return err
	})
	return db
}

// SetBucket will set the bucket used in all database transactions.
func (db *DataBase) SetBucket(name string) {
	db.defaultBucket = []byte(name)
	db.bucketHEAD = []byte(name)
}

// DeleteBucket will delete the bucket given.
//
// Will panic if the name argument is the same as the database's default bucket.
func (db *DataBase) DeleteBucket(name string) error {
	if name == string(db.defaultBucket) {
		panic("cannot delete default bucket")
	}
	return db.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})
}

func (idb *innerdb) view(fn func(*bolt.Bucket) error) error {
	return idb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(idb.bucketHEAD))
		defer idb.resetHEAD()
		return fn(bucket)
	})
}

func (idb *innerdb) update(fn func(*bolt.Bucket) error) error {
	return idb.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(idb.bucketHEAD))
		defer idb.resetHEAD()
		return fn(bucket)
	})
}

func (idb *innerdb) resetHEAD() {
	if bytes.Compare(idb.bucketHEAD, idb.defaultBucket) != 0 {
		idb.bucketHEAD = idb.defaultBucket
	}
}

func filename(file string) string {
	name := filepath.Base(file)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func ensurePath(path string) (err error) {
	p := filepath.Dir(path)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return os.Mkdir(p, 0777)
	}
	return err
}
