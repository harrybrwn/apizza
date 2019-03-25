package cache

import (
	"os"
	"path/filepath"
	"strings"
	"time"

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

// TimerDB is an interface that defines a database that can store timestamps.
type TimerDB interface {
	DB
	TimeStamper
}

// AutoTimerDB is a TimerDB with the AutoTimeStamp convenience function.
type AutoTimerDB interface {
	TimerDB
	AutoTimeStamp(string, time.Duration, func() error, func() error) error
}

// MapDB defines a database that can be converted to a map.
type MapDB interface {
	DB
	Map() (map[string][]byte, error)
}

// FullDB defines a fully featured database.
type FullDB interface {
	MapDB
	Exists(string) bool
	Destroy() error
}

// FullTimerDB defines an interface for fully featured timer database.
type FullTimerDB interface {
	AutoTimerDB
	Exists(string) bool
	Destroy() error
}

var _ FullTimerDB = (*DataBase)(nil)

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

	name := filename(dbfile)
	err = boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	db = &DataBase{
		innerdb: &innerdb{
			DefaultBucket: name,
			db:            boltdb,
			path:          dbfile,
		},
	}
	return db, err
}

type innerdb struct {
	db            *bolt.DB
	DefaultBucket string
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

// ResetTimeStamp stores a new timestamp for the given key.
func (db *DataBase) ResetTimeStamp(key string) error {
	return db.Put(key+"_timestamp", unixNow())
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
	err = db.view(func(b *bolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			all[string(k)] = v
			return nil
		})
	})
	return all, err
}

func (idb *innerdb) view(fn func(*bolt.Bucket) error) error {
	return idb.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(idb.DefaultBucket))
		return fn(bucket)
	})
}

func (idb *innerdb) update(fn func(*bolt.Bucket) error) error {
	return idb.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(idb.DefaultBucket))
		return fn(bucket)
	})
}

func filename(file string) string {
	name := filepath.Base(file)
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func ensurePath(path string) error {
	p := filepath.Dir(path)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return os.Mkdir(p, 0777)
	}
	return nil
}
