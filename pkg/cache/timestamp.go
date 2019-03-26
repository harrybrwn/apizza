package cache

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	timestampSuffix = "_timestamp"
)

// Updater defines an interface for objects used while updating cached timestamps.
type Updater interface {
	OnUpdate() error
	NotUpdate() error
	Decay() time.Duration
}

// TimeStamp gets the timestamp for a given key and will also create one if it does
// not already exist.
func (db *DataBase) TimeStamp(key string) (time.Time, error) {
	return timestamp(db, ts(key))
}

// UpdateTS or "UpdateTimeStamp" will execute an Updater's methods in correspondence with the
// database's timestamp at the key given
func (db *DataBase) UpdateTS(key string, updater Updater) error {
	return check(db, ts(key), updater)
}

// IsTimeStampNotFound returns true when the error given was raised because a
// timestamp was not found.
func IsTimeStampNotFound(e error) bool {
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

func timestamp(db Storage, key string) (time.Time, error) {
	stamp, err := timestampE(db, key)
	if err == nil {
		return stamp, nil
	} else if IsTimeStampNotFound(err) {
		return time.Now(), db.Put(key, unixNow())
	}
	return time.Time{}, err
}

// AutoTimeStamp (Deprecated) will check the timestamp at a given key everytime AutoTimeStamp
// is run. The function given as the 'update' argument will be run if the stored
// timestamp is past the decay argument or if the timestamp associated with that
// key does not exist. The 'notUpdate' argument is a function that will run if
// the timestamp has not expired.
//
// deprecated.
func (db *DataBase) AutoTimeStamp(
	key string,
	decay time.Duration,
	update, notUpdate func() error,
) error {
	fmt.Fprintln(os.Stderr, "Developer Warning: AutoTimeStamp is deprecated.")
	return check(db, ts(key), NewUpdater(decay, update, notUpdate))
	// return errors.New("should not be using DataBase.AutoTimeStamp")
}

// NewUpdater returns an updater from a decay time and two functions.
func NewUpdater(decay time.Duration, update, notUpdate func() error) Updater {
	return &tempUpdater{
		decay:     decay,
		update:    update,
		notUpdate: notUpdate,
	}
}

type tempUpdater struct {
	decay     time.Duration
	update    func() error
	notUpdate func() error
}

func (t *tempUpdater) OnUpdate() error {
	return t.update()
}

func (t *tempUpdater) NotUpdate() error {
	return t.notUpdate()
}

func (t *tempUpdater) Decay() time.Duration {
	return t.decay
}

func check(db Storage, key string, updater Updater) error {
	stamp, err := timestampE(db, key)

	if err != nil {
		if IsTimeStampNotFound(err) {
			if err = db.Put(key, unixNow()); err == nil {
				err = updater.OnUpdate()
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
