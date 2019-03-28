package cache

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const timestampSuffix = "_timestamp"

// Updater defines an interface for objects used while updating cached timestamps.
type Updater interface {
	OnUpdate() error
	NotUpdate() error
	Decay() time.Duration
}

// TimeStamp gets the timestamp for a given key and will also create one if it does
// not already exist.
func (db *DataBase) TimeStamp(key string) (time.Time, error) {
	key = ts(key)
	stamp, err := timestampE(db, key)
	if err == nil {
		return stamp, nil
	} else if isTimeStampNotFound(err) {
		return time.Now(), db.Put(key, unixNow())
	}
	return time.Time{}, err
}

// ResetTimeStamp stores a new timestamp for the given key.
func (db *DataBase) ResetTimeStamp(key string) error {
	return db.Put(ts(key), unixNow())
}

// UpdateTS or "UpdateTimeStamp" will execute an Updater's methods in correspondence with the
// database's timestamp at the key given
func (db *DataBase) UpdateTS(key string, updater Updater) error {
	return check(db, ts(key), updater)
}

// isTimeStampNotFound returns true when the error given was raised because a
// timestamp was not found.
func isTimeStampNotFound(e error) bool {
	return e == errTimeStampNotFound
}

var errTimeStampNotFound = errors.New("could not find timestamp")

type tsGetter interface {
	TimeStamper
	Getter
}

func timestampE(db Getter, key string) (time.Time, error) {
	rawstamp, err := db.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	if rawstamp == nil {
		return time.Time{}, errTimeStampNotFound
	}
	tStamp, err := strconv.ParseInt(string(rawstamp), 10, 64)
	return time.Unix(tStamp, 0), err
}

// NewUpdater returns an updater from a decay time and two functions.
func NewUpdater(decay time.Duration, update, notUpdate func() error) Updater {
	return &updater{
		decay:     decay,
		update:    update,
		notUpdate: notUpdate,
	}
}

type updater struct {
	decay     time.Duration
	update    func() error
	notUpdate func() error
}

func (t *updater) OnUpdate() error {
	return t.update()
}

func (t *updater) NotUpdate() error {
	return t.notUpdate()
}

func (t *updater) Decay() time.Duration {
	return t.decay
}

func check(db Storage, key string, updater Updater) error {
	stamp, err := timestampE(db, key)

	if err != nil {
		if isTimeStampNotFound(err) {
			if err = db.Put(key, unixNow()); err == nil {
				return updater.OnUpdate()
			}
		}
		return err
	}

	if time.Since(stamp) > updater.Decay() {
		if err = updater.OnUpdate(); err != nil {
			return err
		}
		return db.Put(key, unixNow())
	}
	return updater.NotUpdate()
}

func unixNow() []byte {
	newtime := strconv.FormatInt(time.Now().Unix(), 10)
	return []byte(newtime)
}

func ts(key string) string {
	return fmt.Sprintf("%s%s", key, timestampSuffix)
}
