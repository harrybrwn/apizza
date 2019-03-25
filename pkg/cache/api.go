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
