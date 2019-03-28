package cache

import (
	"time"

	"github.com/boltdb/bolt"
)

// Getter is an object that gets.
type Getter interface {
	Get(string) ([]byte, error)
}

// Putter is an object that puts.
type Putter interface {
	Put(string, []byte) error
}

// Storage is a combination of Putter and Getter.
type Storage interface {
	Getter
	Putter
}

// TimeStamper defines objects that support named timestamps.
type TimeStamper interface {
	TimeStamp(string) (time.Time, error)
	ResetTimeStamp(string) error
}

type internalDB interface {
	internal
	DB
}

type internal interface {
	view(func(*bolt.Bucket) error) error
	update(func(*bolt.Bucket) error) error
}

// Im not sure if i'll need these but i'll leave them here just in case.

// TimerDB is an interface that defines a database that can store timestamps.
type TimerDB interface {
	DB
	TimeStamper
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
