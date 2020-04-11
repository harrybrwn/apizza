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
	e, ok := err.(*DominosError)
	if !ok {
		t.Error("should return dominos error")
	}
	if e.Status != -1 {
		t.Error("should be a dominos failure")
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
		tests.StrEq(s.userAddress.Zip(), addr.Zip(), "wrong zip code")
	}
	_, err = GetNearbyStores(&StreetAddr{}, Delivery)
	tests.Exp(err)
	_, err = GetNearbyStores(testAddress(), "")
	tests.Exp(err)
}

func TestInitStore(t *testing.T) {
	tests.InitHelpers(t)
	m := map[string]interface{}{}
	check := func(k string, exp interface{}) {
		if res, ok := m[k]; !ok {
			t.Errorf("key: %s not in store map\n", k)
		} else if res != exp {
			t.Errorf("store map: got %s; want %s\n", res, exp)
		}
	}
	id := "4336"
	tests.Check(InitStore(id, &m))
	check("StoreID", id)
	check("City", "Washington")
	check("Region", "DC")
	check("PreferredCurrency", "USD")
	check("PostalCode", "20005")
	ks := []string{"AddressDescription", "Phone", "AcceptablePaymentTypes", "IsOpen", "IsOnlineNow"}
	for _, k := range ks {
		if _, ok := m[k]; !ok {
			t.Errorf("did not find key %s\n", k)
		}
	}

	test := &struct {
		Status  int
		StoreID string
	}{}
	tests.Check(InitStore(id, test))
	tests.StrEq(test.StoreID, id, "error in InitStore for custom struct")
	if test.Status != 0 {
		t.Error("bad status")
	}
	sMap := map[string]interface{}{}
	tests.Exp(InitStore("", &sMap))
	tests.Exp(InitStore("1234", &sMap))
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
