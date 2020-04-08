package dawg

import (
	"fmt"
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func testAddress() *StreetAddr {
	return &StreetAddr{
		Street:   "1600 Pennsylvania Ave NW",
		CityName: "Washington",
		State:    "DC",
		Zipcode:  "20500",
		AddrType: "House",
	}
}

func TestNewStore(t *testing.T) {
	id := "4339"
	service := Carryout
	s, err := NewStore(id, service, nil)
	if err != nil {
		t.Error(err)
	}
	if s.ID != id {
		t.Errorf("wrong store id: got %s, wanted: %s", s.ID, id)
	}
	if s.userService != service {
		t.Error("userService should not have been changed in NewStore")
	}
	if s.userAddress != nil {
		t.Error("userAddress should be nil here")
	}
	from, to := s.WaitTime()
	if from == 0 || to == 0 {
		t.Error("WaitTime should not be from 0 to 0")
		fmt.Printf("WaitTime(): from: %d, to: %d\n", from, to)
	}
}

func TestNearestStore_Err(t *testing.T) {
	_, err := NearestStore(&StreetAddr{}, Delivery)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetAllNearbyStores(t *testing.T) {
	tests.InitHelpers(t)
	addr := testAddress()
	validation, err := findNearbyStores(orderClient, addr, "Delivery")
	if err != nil {
		t.Error(err)
	}
	stores, err := GetNearbyStores(addr, Delivery)
	tests.Check(err)
	for i, s := range stores {
		if s == nil {
			t.Error("should not have nil store")
		}
		if s.userAddress == nil {
			t.Fatal("nil store.userAddress")
		}
		if s.cli == nil {
			t.Error("should not have nil client when returned from GetNearbyStores")
		}
		tests.StrEq(s.ID, validation.Stores[i].ID, "ids are not the same %s %s", stores[i].ID, validation.Stores[i].ID)
		tests.StrEq(s.Phone, validation.Stores[i].Phone, "wrong phone")
		tests.StrEq(s.userService, Delivery, "wrong service method")
		tests.StrEq(s.userAddress.City(), addr.City(), "wrong city")
		tests.StrEq(s.userAddress.LineOne(), addr.LineOne(), "wrong line one")
		tests.StrEq(s.userAddress.StateCode(), addr.StateCode(), "wrong state code")
		tests.StrEq(s.userAddress.Zip(), addr.Zip(), "wrong zip co de")
	}
	_, err = GetNearbyStores(&StreetAddr{}, Delivery)
	tests.Exp(err)
	_, err = GetNearbyStores(testAddress(), "")
	tests.Exp(err)
}

func TestInitStore(t *testing.T) {
	id := "4336"
	m := map[string]interface{}{}
	if err := InitStore(id, &m); err != nil {
		t.Error(err)
	}
	if m["StoreID"] != id {
		t.Error("wrong store id")
	}
	test := &struct {
		Status  int
		StoreID string
	}{}
	if err := InitStore(id, test); err != nil {
		t.Error(err)
	}
	if test.StoreID != id {
		t.Error("error in InitStore for custom struct")
	}
	if test.Status != 0 {
		t.Error("bad status")
	}
	sMap := map[string]interface{}{}
	err := InitStore("", &sMap)
	if err == nil {
		t.Error("expected error")
	}
	if err = InitStore("1234", &sMap); err == nil {
		t.Error("expected error")
	}
}

func TestInitStore_Err(t *testing.T) {
	ids := []string{"", "0000", "999999999999", "-7765"}
	for _, id := range ids {
		s := new(Store)
		err := initStore(orderClient, id, s)
		if err == nil {
			t.Error("expected error from a ridiculous store id")
		}
	}
}

func TestGetNearestStore(t *testing.T) {
	a := testAddress()
	for _, service := range []string{Delivery, Carryout} {
		s, err := getNearestStore(orderClient, a, service)
		if err != nil {
			t.Error(err)
		}
		if s == nil {
			t.Error("store is nil")
		}
	}
}
