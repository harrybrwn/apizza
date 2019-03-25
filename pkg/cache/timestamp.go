package cache

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

var (
	timestampSuffix = "_timestamp"
)

// TimeChecker is an object that supports named timestamps and runs a set of
// update functions based on those timestamps
type TimeChecker interface {
	Updater
	TimeStamper
	private()
}

// Updater defines an interface for objects used while updating cached timestamps.
type Updater interface {
	OnUpdate() error
	NotUpdate() error
	Decay() time.Duration
}

// NewTimeChecker creates a timestamper from a DB object and an Updater object.
func NewTimeChecker(database DB, updater Updater) TimeChecker {
	var inner *innerdb

	switch db := database.(type) {
	case *DataBase:
		inner = db.innerdb
	case *innerdb:
		inner = db
	default:
		if err := database.Close(); err != nil {
			panic(err)
		}
		boltdb, err := bolt.Open(database.Path(), 0777, nil)
		if err != nil {
			panic(err)
		}
		inner = &innerdb{
			db:            boltdb,
			DefaultBucket: "TimeStamps",
		}
	}

	return &tChecker{
		innerdb: inner,
		updater: updater,
	}
}

// NewTimeCheckerFromBolt creates a TestChecker from an Updater and a *bolt.DB.
func NewTimeCheckerFromBolt(db *bolt.DB, updater Updater) TimeChecker {
	return &tChecker{
		innerdb: &innerdb{
			db:            db,
			DefaultBucket: "TimeStamps",
		},
		updater: updater,
	}
}

type tChecker struct {
	*innerdb
	updater Updater
}

func (tc *tChecker) private() {}

func (tc *tChecker) TimeStamp(key string) (time.Time, error) {
	return timestamp(tc, key)
}

func (tc *tChecker) ResetTimeStamp(key string) error {
	return tc.Put(ts(key), unixNow())
}

func (tc *tChecker) OnUpdate() error {
	return tc.updater.OnUpdate()
}

func (tc *tChecker) NotUpdate() error {
	return tc.updater.NotUpdate()
}

func (tc *tChecker) Decay() time.Duration {
	return tc.updater.Decay()
}

func (tc *tChecker) Exists(key string) bool {
	return exists(tc.innerdb, key)
}

// TimeStamp gets the timestamp for a given key and will also create one if it does
// not already exist.
func (db *DataBase) TimeStamp(key string) (time.Time, error) {
	return timestamp(db, ts(key))
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

// AutoTimeStamp will check the timestamp at a given key everytime AutoTimeStamp
// is run. The function given as the 'update' argument will be run if the stored
// timestamp is past the decay argument or if the timestamp associated with that
// key does not exist. The 'notUpdate' argument is a function that will run if
// the timestamp has not expired.
//
// depricated.
func (db *DataBase) AutoTimeStamp(
	key string,
	decay time.Duration,
	update, notUpdate func() error,
) error {
	return check(db, key, newUpdater(decay, update, notUpdate))
}

func newUpdater(decay time.Duration, update, notUpdate func() error) Updater {
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

// Check will execute an Updater's methods in correspondence with the database's
// timestamp at the key given
func (db *DataBase) Check(key string, updater Updater) error {
	return check(db, ts(key), updater)
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
