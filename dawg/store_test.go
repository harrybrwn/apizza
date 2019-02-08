package dawg

import (
	"fmt"
	"testing"
)

func testAddress() *Address {
	return &Address{
		StreetNum: "1600",
		Street:    "Pennsylvania Ave NW",
		City:      "Washington",
		State:     "DC",
		Zip:       "20500",
		AddrType:  "House",
	}
}

func TestFindNearbyStores(t *testing.T) {
	addr := testAddress()
	for _, service := range []string{"Delivery", "Carryout"} {
		_, err := findNearbyStores(addr, service)
		if err != nil {
			t.Error("\n TestNearbyStore:", err, "\n")
		}
	}
}

func TestFindNearbyStores_Err(t *testing.T) {
	addr := testAddress()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Should panic")
			} else if r != "service must be either 'Delivery' or 'Carryout'" {
				t.Errorf("caught wrong panic msg: %v\n", r)
			}
		}()
		_, _ = findNearbyStores(addr, "invalid service")
	}()

	_, err := findNearbyStores(&Address{}, "Delivery")
	if err == nil {
		t.Error("should return error")
	}
}

func TestNewStore(t *testing.T) {
	id := "4339"
	service := "Carryout"
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
	service := "Delivery"

	s, err := NearestStore(addr, service)
	if err != nil {
		t.Error(err)
	}
	if s.userService != service {
		t.Error("userService was changed in NearestStore")
	}
	nearbyStores, err := findNearbyStores(addr, service)
	if err != nil {
		t.Error(err)
	}
	if nearbyStores.Stores[0].ID != s.ID {
		t.Error("NearestStore and findNearbyStores found different stores for the same address")
	}
	for _, s := range nearbyStores.Stores {
		if s.userAddress != addr {
			t.Error("should have same address")
		}
		if s.userService != service {
			t.Error("should have same service")
		}
	}
}

func TestNearestStore_Err(t *testing.T) {
	_, err := NearestStore(&Address{}, "Delivery")
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetAllNearbyStores(t *testing.T) {
	addr := testAddress()

	validationStores, err := findNearbyStores(addr, "Delivery")
	if err != nil {
		t.Error(err)
	}

	stores, err := GetAllNearbyStores(addr, "Delivery")
	if err != nil {
		t.Error(err)
	}
	for i, s := range stores {
		if s.ID != validationStores.Stores[i].ID {
			t.Error("ids are not the same", stores[i].ID, validationStores.Stores[i].ID)
		}
	}
}

func TestGetAllNearbyStores_Err(t *testing.T) {
	_, err := GetAllNearbyStores(&Address{}, "Delivery")
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

	type Test struct {
		Status  int
		StoreID string
	}
	test := &Test{}
	if err := InitStore(id, test); err != nil {
		t.Error(err)
	}
	if test.StoreID != id {
		t.Error("error in InitStore for custom struct")
	}
	if test.Status != 0 {
		t.Error("bad status")
	}
}

func TestInitStore_Err(t *testing.T) {
	sMap := map[string]interface{}{}
	err := InitStore("", &sMap)
	if err == nil {
		t.Error("expected error")
	}
}
