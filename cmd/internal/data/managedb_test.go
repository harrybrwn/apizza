package data

import (
	"bytes"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestDBManagment(t *testing.T) {
	db, err := cache.GetDB(tests.TempFile())
	if err != nil {
		t.Error(err)
	}
	db.Get("name")
	a := &obj.Address{
		Street:   "1600 Pennsylvania Ave NW",
		CityName: "Washington",
		State:    "DC",
		Zipcode:  "20500",
	}
	s, _ := dawg.NearestStore(a, "Delivery")
	o := s.NewOrder()
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
