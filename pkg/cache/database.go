package cache

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// DataBase is a wrapper struct for boltdb key-value pair database.
type DataBase struct {
	Path          string
	DefaultBucket string
	db            *bolt.DB
}

// GetDB returns an initialized DataBase. Will either create a brand new boltdb
// or open existing one.
func GetDB(dbfile string) (db *DataBase, err error) {
	err = ensureDBPath(dbfile)
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
		Path:          dbfile,
		DefaultBucket: name,
		db:            boltdb,
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
func (db *DataBase) Get(key string) (raw []byte, err error) {
	err = db.view(func(b *bolt.Bucket) error {
		raw = b.Get([]byte(key))
		return nil
	})
	return raw, err
}

// Delete removes the data for a specific key.
func (db *DataBase) Delete(key string) error {
	return db.update(func(b *bolt.Bucket) error {
		return b.Delete([]byte(key))
	})
}

// Exists will return true if the key supplied has data associated with it.
func (db *DataBase) Exists(key string) (exists bool) {
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

// TimeStamp gets the timestamp for a given key and will also create one if it does
// not already exist.
func (db *DataBase) TimeStamp(key string) (time.Time, error) {
	key += "_timestamp"
	if !db.Exists(key) {
		return time.Now(), db.Put(key, timeStampNow())
	}

	rawstamp, err := db.Get(key)
	if err != nil {
		return time.Time{}, err
	}
	tStamp, err := strconv.ParseInt(string(rawstamp), 10, 64)
	return time.Unix(tStamp, 0), err
}

// AutoTimeStamp will check the timestamp at a given key everytime AutoTimeStamp
// is run. The function given as the 'update' argument will be run if the stored
// timestamp is past the decay argument or if the timestamp associated with that
// key does not exist. The 'notUpdate' argument is a function that will run if
// the timestamp has not expired.
func (db *DataBase) AutoTimeStamp(
	key string,
	decay time.Duration,
	update, notUpdate func() error,
) error {
	if !db.Exists(key + "_timestamp") {
		if _, err := db.TimeStamp(key); err != nil {
			return err
		}
		return update()
	}
	tstamp, err := db.TimeStamp(key)
	if err != nil {
		return err
	}

	if time.Since(tstamp) > decay {
		if err = update(); err != nil {
			return err
		}
		return db.ResetTimeStamp(key)
	}
	if notUpdate == nil {
		return nil
	}
	return notUpdate()
}

// ResetTimeStamp stores a new timestamp for the given key.
func (db *DataBase) ResetTimeStamp(key string) error {
	return db.Put(key+"_timestamp", timeStampNow())
}

// Close will close the DataBase's inner bolt.DB
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

// GetAll returns a map of all the key-value pairs in the database.
func (db *DataBase) GetAll() (all map[string][]byte, err error) {
	all = map[string][]byte{}
	err = db.view(func(b *bolt.Bucket) error {
		return b.ForEach(func(k, v []byte) error {
			all[string(k)] = v
			return nil
		})
	})
	return all, err
}

func (db *DataBase) view(fn func(*bolt.Bucket) error) error {
	return db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.DefaultBucket))
		return fn(bucket)
	})
}

func (db *DataBase) update(fn func(*bolt.Bucket) error) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db.DefaultBucket))
		return fn(bucket)
	})
}

func filename(file string) string {
	fname := filepath.Base(file)
	return strings.TrimSuffix(fname, filepath.Ext(fname))
}

func timeStampNow() []byte {
	newtime := strconv.FormatInt(time.Now().Unix(), 10)
	return []byte(newtime)
}

func ensureDBPath(path string) error {
	p := filepath.Dir(path)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return os.Mkdir(p, 0700)
	}
	return nil
}
