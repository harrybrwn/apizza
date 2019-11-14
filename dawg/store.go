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
			store = allnearby.Stores[i]
			break
		}
	}
	store.userAddress, store.userService = addr, service
	return store, InitStore(store.ID, store)
}

// GetNearbyStores is a way of getting all the nearby stores
// except they will by full initialized.
func GetNearbyStores(addr Address, service string) ([]*Store, error) {
	all, err := findNearbyStores(addr, service)
	if err != nil {
		return nil, err
	}
	for i := range all.Stores {
		if err = InitStore(all.Stores[i].ID, &all.Stores[i]); err != nil {
			return nil, err
		}
	}
	return all.Stores, nil
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
	return store, InitStore(id, store)
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
	if err := dominosErr(b); err != nil {
		return err
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
	Address              string            `json:"AddressDescription"`
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
	menu                 *Menu
}

// Menu returns the menu for a store object
func (s *Store) Menu() (*Menu, error) {
	if s.menu != nil && s.menu.ID == s.ID {
		return s.menu, nil
	}
	return newMenu(s.ID)
}

// NewOrder is a convenience function for creating an order from some of the store variables.
func (s *Store) NewOrder() *Order {
	return &Order{
		LanguageCode:  DefaultLang,
		ServiceMethod: s.userService,
		StoreID:       s.ID,
		Products:      []*OrderProduct{},
		Address:       StreetAddrFromAddress(s.userAddress),
		Payments:      []*orderPayment{},
	}
}

// MakeOrder will make an order from a first and last name and an emil.
func (s *Store) MakeOrder(firstname, lastname, email string) *Order {
	return &Order{
		FirstName:     firstname,
		LastName:      lastname,
		Email:         email,
		LanguageCode:  DefaultLang,
		ServiceMethod: s.userService,
		StoreID:       s.ID,
		Products:      []*OrderProduct{},
		Address:       StreetAddrFromAddress(s.userAddress),
		Payments:      []*orderPayment{},
	}
}

// GetProduct finds the menu Product that matchs the given product code.
func (s *Store) GetProduct(code string) (*Product, error) {
	menu, err := s.Menu()
	if err != nil {
		return nil, err
	}
	return menu.GetProduct(code)
}

// GetVariant will get a fully initialized varient from the menu.
func (s *Store) GetVariant(code string) (*Variant, error) {
	menu, err := s.Menu()
	if err != nil {
		return nil, err
	}
	return menu.GetVariant(code)
}

// WaitTime returns a pair of integers that are the maximum and
// minimum estimated wait time for that store.
func (s *Store) WaitTime() (min int, max int) {
	m := s.ServiceEstimatedWait
	return m[s.userService].Min, m[s.userService].Max
}

type storeLocs struct {
	Granularity string      `json:"Granularity"`
	Address     *StreetAddr `json:"Address"`
	Stores      []*Store    `json:"Stores"`
}

func findNearbyStores(addr Address, service string) (*storeLocs, error) {
	if !(service == "Delivery" || service == "Carryout") {
		panic("service must be either 'Delivery' or 'Carryout'")
	}
	// TODO: on the dominos website, the c param can sometimes be just the zip code
	// and it still works.
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
	if err != nil {
		return nil, err
	}
	for i := range locs.Stores {
		locs.Stores[i].userAddress, locs.Stores[i].userService = addr, service
	}
	return locs, dominosErr(b)
}

// storeErrWrapper and initStores are first drafts of concurrent store infrastructure
type storeErrWrapper struct {
	store *Store
	e     error
}

func initStoreChan(s *Store, c chan *Store) {
	if err := InitStore(s.ID, s); err != nil {
		panic(err)
	}
	c <- s
}

func initStores(stores []*Store) chan *Store {
	c := make(chan *Store)

	for i := range stores {
		go initStoreChan(stores[i], c) // send the initialized store through the channel
	}
	return c
}
