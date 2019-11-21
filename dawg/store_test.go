package dawg

import (
	"fmt"
	"testing"
	"time"
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
	if err != errBadService {
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
	m, err1 := s.Menu()
	m2, err2 := s.Menu()
	err = errpair(err1, err2)
	if err != nil {
		t.Error(err)
	}
	if m != m2 {
		t.Errorf("should be the same for two calls to Store.Menu: %p %p", m, m2)
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
	addr := testAddress()
	for _, service := range []string{Delivery, Carryout} {
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
			if s.userService != service {
				t.Error("wrong service method")
			}
			if s.userAddress.City() != addr.City() {
				t.Error("wrong city")
			}
			if s.userAddress.LineOne() != addr.LineOne() {
				t.Error("wrong line one")
			}
			if s.userAddress.StateCode() != addr.StateCode() {
				t.Error("wrong state code")
			}
			if s.userAddress.Zip() != addr.Zip() {
				t.Error("wrong zip co de")
			}
		}
	}
}

func TestGetAllNearbyStores_Async_Err(t *testing.T) {
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
	addr := testAddress()
	serv := Delivery
	tm := time.Now()
	storesInOrder, err := GetNearbyStores(addr, serv)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("old:", time.Now().Sub(tm))

	tm = time.Now()
	stores, err := asyncNearbyStores(orderClient, addr, serv)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("asyncInOrder:", time.Now().Sub(tm))

	n := len(storesInOrder)
	if n != len(stores) {
		t.Fatal("the results did not return lists of the same length")
	}
	for i := 0; i < n; i++ {
		if storesInOrder[i].ID != stores[i].ID {
			t.Error("wrong id")
		}
		if storesInOrder[i].Phone != stores[i].Phone {
			t.Error("stores have different phone numbers")
		}
		if storesInOrder[i].Address != stores[i].Address {
			t.Error("stores have different addresses")
		}
	}
}
