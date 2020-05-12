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

func TestDBManagement(t *testing.T) {
	tests.InitHelpers(t)
	db := cmdtest.TempDB()
	var err error
	o := testStore.NewOrder()
	o.SetName("test_order")
	buf := &bytes.Buffer{}

	tests.Check(PrintOrders(db, buf, false, false))
	tests.Compare(t, buf.String(), "No orders saved.\n")
	buf.Reset()

	tests.Check(SaveOrder(o, buf, db))
	tests.Compare(t, buf.String(), "order successfully updated.\n")
	buf.Reset()

	tests.Check(PrintOrders(db, buf, false, false))
	tests.Compare(t, buf.String(), "Your Orders:\n  test_order\n")
	buf.Reset()

	_, err = GetOrder("badorder", db)
	tests.Exp(err)
	newO, err := GetOrder("test_order", db)
	tests.Check(err)
	tests.StrEq(newO.Name(), o.Name(), "wrong order")
	tests.StrEq(newO.Address.LineOne(), o.Address.LineOne(), "wrong address saved")
	tests.StrEq(newO.Address.City(), o.Address.City(), "wrong address saved")
	tests.Check(db.Destroy())
}

func TestPrintOrders(t *testing.T) {
	tests.InitHelpers(t)
	var err error
	o := testStore.NewOrder()
	p, err := testStore.GetVariant("10SCREEN")
	tests.Check(err)
	if p == nil {
		t.Fatal("got nil product")
	}
	tests.Check(o.AddProductQty(p, 10))
	db := cmdtest.TempDB()
	defer func() { tests.Check(db.Destroy()) }()

	o.SetName("test_order")
	buf := new(bytes.Buffer)
	tests.Check(SaveOrder(o, buf, db))
	buf.Reset()
	tests.Check(PrintOrders(db, buf, true, false))
	tests.Compare(t, buf.String(), "Your Orders:\n  test_order -  10SCREEN, \n")
}

func TestMenuCacherJSON(t *testing.T) {
	t.Skip()
	tests.InitHelpers(t)
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

	tests.Check(db.UpdateTS("menu", cacher))
	if c.m == nil {
		t.Error("cacher should have a menu now")
	}
	if cacher.Menu() == nil {
		t.Error("cacher should have a menu now")
	}
	data, err := db.Get("menu")
	tests.Check(err)
	if len(data) == 0 {
		t.Error("should have stored a menu")
	}

	if buf.String() != "caching another menu\n" {
		t.Error("should log with logging package about a new menu")
	}
	buf.Reset()

	tests.Check(db.UpdateTS("menu", cacher))
	c.m = nil
	tests.Check(db.UpdateTS("menu", c))
}
