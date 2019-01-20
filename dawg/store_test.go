package dawg

import (
	"fmt"
	"testing"
)

func TestFindNearbyStores(t *testing.T) {
	var addr = &Address{
		StreetNum: "1600",
		Street:    "Pennsylvania Ave NW",
		City:      "Washington",
		State:     "DC",
		Zip:       "20500",
		AddrType:  "House",
	}
	for _, service := range []string{"Delivery", "Carryout"} {
		_, err := findNearbyStores(addr, service)
		if err != nil {
			t.Error("\n TestNearbyStore:", err, "\n")
		}
	}
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

func TestNearbyStore(t *testing.T) {
	var addr = &Address{
		StreetNum: "1600",
		Street:    "Pennsylvania Ave NW",
		City:      "Washington",
		State:     "DC",
		Zip:       "20500",
		AddrType:  "House",
	}
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
		fmt.Println(err)
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
	service = "Carryout"
	stores, err := GetAllNearbyStores(addr, service)
	if err != nil {
		t.Error(err)
	}
	for i, s := range stores {
		if s.ID != nearbyStores.Stores[i].ID {
			t.Error("ids are not the same", stores[i].ID, nearbyStores.Stores[i].ID)
		}
	}
}

func TestInitStore(t *testing.T) {
	id := "4336"
	s := &Store{}
	if err := InitStore(id, s); err != nil {
		t.Error(err)
	}

	m := &map[string]interface{}{}
	if err := InitStore(id, m); err != nil {
		t.Error(err)
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
}
