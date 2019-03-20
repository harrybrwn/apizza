package dawg

import (
	"bytes"
	"encoding/json"
	"errors"
)

// NearestStore gets the dominos location closest to the given address.
//
// The addr argument should be the address to deliver to not the address of the
// store itself. The service should be either "Carryout" or "Delivery", this will
// deturmine wether the final order will be for pickup or delivery.
func NearestStore(addr Address, service string) (*Store, error) {
	if addr == nil {
		return nil, errors.New("empty address")
	}
	allnearby, err := findNearbyStores(addr, service)
	if err != nil {
		return nil, err
	}
	var store *Store

	for i := range allnearby.Stores {
		if allnearby.Stores[i].IsOnlineNow {
			store = &allnearby.Stores[i]
			break
		}
	}
	store.userAddress, store.userService = addr, service
	return store, InitStore(store.ID, store)
}

// GetAllNearbyStores is a way of getting all the nearby stores
// except they will by full initialized.
func GetAllNearbyStores(addr Address, service string) ([]Store, error) {
	all, err := findNearbyStores(addr, service)
	if err != nil {
		return nil, err
	}
	for i := range all.Stores {
		err = InitStore(all.Stores[i].ID, &all.Stores[i])
	}
	return all.Stores, err
}

// NewStore returns the default Store object given a store id.
//
// The service and address arguments are for store functions that require those
// variables. If the store does not need to use those functions, service can be
// an empty string, and add can be 'nil'.
//
// The addr argument should be the address to deliver to not the address of the
// store itself.
func NewStore(id string, service string, addr Address) (*Store, error) {
	store := &Store{userService: service, userAddress: addr}
	err := InitStore(id, store)
	return store, err
}

// InitStore allows for the creation of arbitrary store objects. The main
// use-case is when data id needed that is not in the default Store object.
// The obj argument is anything that would be used to decode json data.
//
// Use type '*map[string]interface{}' in the object argument for all the
// store data
func InitStore(id string, obj interface{}) error {
	path := format("/power/store/%s/profile", id)
	b, err := get(path, nil)
	if err != nil {
		return err
	}
	if bytes.HasPrefix(b, []byte("<!DOCTYPE html>")) {
		return errors.New("invalid 'id' argument")
	}
	return json.Unmarshal(b, obj)
}

// The Store object represents a physical dominos location.
type Store struct {
	ID                   string            `json:"StoreID"`
	Status               int               `json:"Status"`
	IsOpen               bool              `json:"IsOpen"`
	IsOnlineNow          bool              `json:"IsOnlineNow"`
	IsDeliveryStore      bool              `json:"IsDeliveryStore"`
	Phone                string            `json:"Phone"`
	PaymentTypes         []string          `json:"AcceptablePaymentTypes"`
	CreditCardTypes      []string          `json:"AcceptableCreditCards"`
	MinDistance          float64           `json:"MinDistance"`
	MaxDistance          float64           `json:"MaxDistance"`
	PostalCode           string            `json:"PostalCode"`
	City                 string            `json:"City"`
	StoreCoords          map[string]string `json:"StoreCoordinates"`
	ServiceEstimatedWait map[string]struct {
		Min int `json:"Min"`
		Max int `json:"Max"`
	} `json:"ServiceMethodEstimatedWaitMinutes"`
	Hours                map[string][]map[string]string `json:"Hours"`
	MinDeliveryOrderAmnt float64                        `json:"MinimumDeliveryOrderAmount"`
	userAddress          Address
	userService          string
}

// Menu returns the menu for a store object
func (s *Store) Menu() (*Menu, error) {
	return newMenu(s.ID)
}

// NewOrder is a convenience function for creating an order from some of the store variables.
func (s *Store) NewOrder() *Order {
	return &Order{
		LanguageCode:  DefaultLang,
		ServiceMethod: s.userService,
		StoreID:       s.ID,
		Products:      []*Product{},
		Address:       StreetAddrFromAddress(s.userAddress),
		Payments:      []Payment{},
	}
}

// GetProduct finds the menu item that matchs the given product code.
func (s *Store) GetProduct(code string) (*Product, error) {
	// get a menu and find the map that matches the Code
	menu, err := s.Menu()
	if err != nil {
		return nil, err
	}
	var p *Product

	if data, ok := menu.Variants[code]; ok {
		p, err = makeProduct(data.(map[string]interface{}))
	} else if data, ok := menu.Preconfigured[code]; ok {
		p, err = makeProduct(data.(map[string]interface{}))
	} else {
		return nil, errors.New("cannot find product")
	}
	return p, err
}

// WaitTime returns a pair of integers that are the maximum and
// minimum estimated wait time for that store.
func (s *Store) WaitTime() (int, int) {
	m := s.ServiceEstimatedWait
	return m[s.userService].Min, m[s.userService].Max
}

type storeLocs struct {
	Status      int         `json:"Status"`
	Granularity string      `json:"Granularity"`
	Address     *StreetAddr `json:"Address"`
	Stores      []Store     `json:"Stores"`
}

func findNearbyStores(addr Address, service string) (*storeLocs, error) {
	if !(service == "Delivery" || service == "Carryout") {
		panic("service must be either 'Delivery' or 'Carryout'")
	}
	b, err := get("/power/store-locator", &Params{
		"s":    addr.LineOne(),
		"c":    format("%s, %s %s", addr.City(), addr.StateCode(), addr.Zip()),
		"type": service,
	})
	if err != nil {
		return nil, err
	}
	locs := &storeLocs{}
	err = json.Unmarshal(b, locs)
	if err == nil && locs.Status == -1 {
		return locs, errors.New("Dominos server Failure: -1")
	}
	for i := range locs.Stores {
		locs.Stores[i].userAddress, locs.Stores[i].userService = addr, service
	}
	return locs, err
}
