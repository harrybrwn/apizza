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

// TimeStamper defines objects that support named timestamps.
type TimeStamper interface {
	TimeStamp(string) (time.Time, error)
}

type internalDB interface {
	internal
	DB
}

type internal interface {
	view(func(*bolt.Bucket) error) error
	update(func(*bolt.Bucket) error) error
}
