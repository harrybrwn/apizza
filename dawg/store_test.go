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

func TestFindNearbyStores(t *testing.T) {
	for _, service := range []string{Delivery, Carryout} {
		_, err := findNearbyStores(orderClient, testAddress(), service)
		if err != nil {
			t.Error("\n TestNearbyStore:", err, "\n")
		}
	}
}

func TestFindNearbyStores_Err(t *testing.T) {
	_, err := findNearbyStores(orderClient, testAddress(), "invalid service")
	if err == nil {
		t.Error("expected an error")
	}
	if err != ErrBadService {
		t.Error("findNearbyStores gave the wrong error for an invalid service")
	}
	_, err = findNearbyStores(orderClient, &StreetAddr{}, Delivery)
	if err == nil {
		t.Error("should return error")
	}
	if _, err := NearestStore(nil, Delivery); err == nil {
		t.Error("expected error")
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

func TestNearestStore(t *testing.T) {
	if testing.Short() {
		return
	}
	addr := testAddress()
	service := Delivery
	s, err := NearestStore(addr, service)
	if err != nil {
		t.Error(err)
	}
	if s.cli == nil {
		t.Error("NearestStore should not return a store with a nil client")
	}
	if s.userService != service {
		t.Error("userService was changed in NearestStore")
	}
	nearbyStores, err := findNearbyStores(orderClient, addr, service)
	if err != nil {
		t.Error(err)
	}
	if nearbyStores.Stores[0].ID != s.ID {
		t.Error("NearestStore and findNearbyStores found different stores for the same address")
	}
	for _, s := range nearbyStores.Stores {
		if s.userAddress != nil {
			t.Error("userAddress should not have been initialized")
			continue
		}
		if s.userService != "" {
			t.Error("the userService should not have been initialized")
			continue
		}
	}
	if !testing.Short() {
		m, err1 := s.Menu()
		m2, err2 := s.Menu()
		err = errpair(err1, err2)
		if err != nil {
			t.Error(err)
		}
		if m != m2 {
			t.Errorf("should be the same for two calls to Store.Menu: %p %p", m, m2)
		}
	}
	id := s.ID
	s.ID = ""
	_, err = s.Menu()
	if err == nil {
		t.Error("expected an error for a store with no id")
	}
	_, err = s.GetProduct("14SCREEN")
	if err == nil {
		t.Error("expected error from a store with no id")
	}
	_, err = s.GetVariant("14SCREEN")
	if err == nil {
		t.Error("expected error from a store with no id")
	}
	s.ID = id
}

func TestGetAllNearbyStores_Async(t *testing.T) {
	tests.InitHelpers(t)
	addr := testAddress()
	var services []string
	if testing.Short() {
		services = []string{Delivery}
	} else {
		services = []string{Delivery, Carryout}
	}

	for _, service := range services {
		stores, err := GetNearbyStores(addr, service)
		if err != nil {
			t.Error(err)
		}
		for _, s := range stores {
			if s == nil {
				t.Error("should not have nil store")
			}
			if s.userAddress == nil {
				t.Fatal("nil store.userAddress")
			}
			tests.StrEq(s.userService, service, "wrong service method")
			tests.StrEq(s.userAddress.City(), addr.City(), "wrong city")
			tests.StrEq(s.userAddress.LineOne(), addr.LineOne(), "wrong line one")
			tests.StrEq(s.userAddress.StateCode(), addr.StateCode(), "wrong state code")
			tests.StrEq(s.userAddress.Zip(), addr.Zip(), "wrong zip co de")
		}
	}
}

func TestGetAllNearbyStores_Async_Err(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	_, err := GetNearbyStores(&StreetAddr{}, Delivery)
	if err == nil {
		t.Error("expected error")
	}
	_, err = GetNearbyStores(testAddress(), "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestNearestStore_Err(t *testing.T) {
	_, err := NearestStore(&StreetAddr{}, Delivery)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetAllNearbyStores(t *testing.T) {
	validationStores, err := findNearbyStores(orderClient, testAddress(), "Delivery")
	if err != nil {
		t.Error(err)
	}
	stores, err := GetNearbyStores(testAddress(), "Delivery")
	if err != nil {
		t.Error(err)
	}
	for i, s := range stores {
		if s.ID != validationStores.Stores[i].ID {
			t.Error("ids are not the same", stores[i].ID, validationStores.Stores[i].ID)
		}
		if s.cli == nil {
			t.Error("should not have nil client when returned from GetNearbyStores")
		}
	}
	_, err = GetNearbyStores(&StreetAddr{}, "Delivery")
	if err == nil {
		t.Error("expected error")
	}
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

func Test_initStoreErr(t *testing.T) {
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

func TestAsyncInOrder(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	tests.InitHelpers(t)
	addr := testAddress()
	serv := Delivery
	storesInOrder, err := GetNearbyStores(addr, serv)
	tests.Check(err)
	stores, err := asyncNearbyStores(orderClient, addr, serv)
	tests.Check(err)

	n := len(storesInOrder)
	if n != len(stores) {
		t.Fatal("the results did not return lists of the same length")
	}
	for i := 0; i < n; i++ {
		tests.StrEq(storesInOrder[i].ID, stores[i].ID, "wrong id")
		tests.StrEq(storesInOrder[i].Phone, stores[i].Phone, "stores have different phone numbers")
		tests.StrEq(storesInOrder[i].Address, stores[i].Address, "stores have different addresses")
	}
}
