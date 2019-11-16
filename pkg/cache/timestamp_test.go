package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestNewTimeChecker(t *testing.T) {}

func TestTimeStamp(t *testing.T) {
	db, err := GetDB(tests.TempFile())
	if err != nil || db == nil {
		t.Fatal("bad db creation")
	}
	stamp, err := db.TimeStamp("test")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second / 4)
	tdiff := time.Since(stamp)
	if time.Millisecond*240 > tdiff || tdiff > time.Millisecond*260 {
		t.Error("time stamp is not in the right range (240ms to 260ms); got", tdiff)
	}
	time.Sleep(time.Second / 4)
	stamp2, err := db.TimeStamp("test")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
	t1, t2 := time.Since(stamp), time.Since(stamp2)
	if t1 > t2 {
		t.Error("time one shouldn't be bigger than t1")
	}
	if t2-t1 > time.Second {
		t.Error("time difference is way too big")
	}
	if err := db.ResetTimeStamp("test"); err != nil {
		t.Error(err)
	}
	stampRe, err := db.TimeStamp("test")
	if err != nil {
		t.Error(err)
	}
	if time.Since(stampRe) > t2 {
		t.Error("timestamp didn't reset")
	}
	if err = db.Destroy(); err != nil {
		t.Error(err)
	}
}

func TestAutoTimeStamp(t *testing.T) {
	db, err := GetDB(tests.TempFile())
	if err != nil || db == nil {
		t.Fatal("bad db creation")
	}
	if err = db.UpdateTS("test", NewUpdater(time.Second/10, func() error { return nil }, func() error { return nil })); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second / 2)
	if err = db.UpdateTS("test", NewUpdater(time.Second/10, func() error { return nil }, func() error { return nil })); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second / 2)
	if err = db.UpdateTS("test", NewUpdater(time.Second/10, func() error { return errors.New("this error should be raised") }, func() error { return nil })); err == nil {
		t.Error("expected error from update func")
	}
	if err = db.UpdateTS("test", NewUpdater(time.Second*2, func() error { return nil }, func() error { return errors.New("this error should be raised") })); err == nil {
		t.Error("expected error from notUpdate func")
	}
	if err = db.UpdateTS("test", NewUpdater(time.Second*2, func() error { return nil }, func() error { return nil })); err != nil {
		t.Error("notUpdate passed as nil:", err)
	}
	if err = db.UpdateTS("test", NewUpdater(time.Second*2, func() error { return errors.New("update func shouldn't be run but was") }, func() error { return nil })); err != nil {
		t.Error(err)
	}
	if err = db.Destroy(); err != nil {
		t.Error(err)
	}
}
