package data

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/tests"
)

var testStore *dawg.Store

func init() {
	a := &obj.Address{
		Street:   "1600 Pennsylvania Ave NW",
		CityName: "Washington",
		State:    "DC",
		Zipcode:  "20500",
	}
	testStore, _ = dawg.NearestStore(a, "Delivery")
}

func TestDBManagment(t *testing.T) {
	db := cmdtest.TempDB()
	defer db.Destroy()

	var err error
	o := testStore.NewOrder()
	o.SetName("test_order")
	buf := &bytes.Buffer{}

	if err = PrintOrders(db, buf, false); err != nil {
		t.Error(err)
	}
	tests.Compare(t, buf.String(), "No orders saved.\n")
	buf.Reset()

	if err = SaveOrder(o, buf, db); err != nil {
		t.Error(err)
	}
	tests.Compare(t, buf.String(), "order successfully updated.\n")
	buf.Reset()

	if err = PrintOrders(db, buf, false); err != nil {
		t.Error(err)
	}
	tests.Compare(t, buf.String(), "Your Orders:\n  test_order\n")
	buf.Reset()

	if _, err := GetOrder("badorder", db); err == nil {
		t.Error("expected error")
	}
	newO, err := GetOrder("test_order", db)
	if err != nil {
		t.Error(err)
	}
	if newO.Name() != o.Name() {
		t.Error("wrong order")
	}
	if newO.Address.LineOne() != o.Address.LineOne() {
		t.Error("wrong address saved")
	}
	if newO.Address.City() != o.Address.City() {
		t.Error("wrong address saved")
	}
	if err = db.Destroy(); err != nil {
		t.Error(err)
	}
}

func TestPrintOrders(t *testing.T) {
	var err error
	o := testStore.NewOrder()
	p, err := testStore.GetVariant("10SCREEN")
	if err != nil {
		t.Error(err)
	}
	if p == nil {
		t.Fatal("got nil product")
	}
	if err = o.AddProductQty(p, 10); err != nil {
		t.Error(err)
	}
	db := cmdtest.TempDB()
	defer db.Destroy()

	o.SetName("test_order")
	buf := new(bytes.Buffer)
	if err = SaveOrder(o, buf, db); err != nil {
		t.Error(err)
	}
	buf.Reset()
	if err = PrintOrders(db, buf, true); err != nil {
		t.Error(err)
	}
	exp := "Your Orders:\n  test_order -  10SCREEN, \n"
	tests.Compare(t, buf.String(), exp)
}

func TestMenuCacherJSON(t *testing.T) {
	t.Skip()
	var err error
	db := cmdtest.TempDB()
	defer db.Destroy()

	cacher := NewMenuCacher(time.Second, db, func() *dawg.Store { return testStore })
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)

	c := cacher.(*generalMenuCacher)
	if c.m != nil {
		t.Error("cacher should not have a menu yet")
	}
	if cacher.Menu() != nil {
		t.Error("cacher should not have a menu yet")
	}

	if err = db.UpdateTS("menu", cacher); err != nil {
		t.Error(err)
	}
	if c.m == nil {
		t.Error("cacher should have a menu now")
	}
	if cacher.Menu() == nil {
		t.Error("cacher should have a menu now")
	}
	data, err := db.Get("menu")
	if err != nil {
		t.Error(err)
	}
	if len(data) == 0 {
		t.Error("should have stored a menu")
	}

	if buf.String() != "caching another menu\n" {
		t.Error("should log with logging package about a new menu")
	}
	buf.Reset()

	if err = db.UpdateTS("menu", cacher); err != nil {
		t.Error(err)
	}
	c.m = nil
	if err = db.UpdateTS("menu", c); err != nil {
		t.Error(err)
	}
}
