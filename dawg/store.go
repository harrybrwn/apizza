package dawg

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

const (
	// Delivery is a dominos service method that will result
	// in a pizza delivery.
	Delivery = "Delivery"

	// Carryout is a dominos service method that
	// will require users to go and pickup their pizza.
	Carryout = "Carryout"
)

// NearestStore gets the dominos location closest to the given address.
//
// The addr argument should be the address to deliver to not the address of the
// store itself. The service should be either "Carryout" or "Delivery", this will
// deturmine wether the final order will be for pickup or delivery.
func NearestStore(addr Address, service string) (*Store, error) {
	return getNearestStore(orderClient, addr, service)
}

// GetNearbyStores is a way of getting all the nearby stores
// except they will by full initialized.
func GetNearbyStores(addr Address, service string) ([]*Store, error) {
	all, err := findNearbyStores(orderClient, addr, service)
	if err != nil {
		return nil, err
	}
	for i, store := range all.Stores {
		store.userAddress = addr
		store.userService = service
		if err = initStore(orderClient, all.Stores[i].ID, all.Stores[i]); err != nil {
			return nil, err
		}
	}
	return all.Stores, nil
}

// GetNearbyStoresAsync will retrive a list of nearby stores asyncronously, meaning that
// the first item on the list is not the closest.
func GetNearbyStoresAsync(addr Address, service string) ([]*Store, error) {
	return asyncNearbyStores(orderClient, addr, service)
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
	store := &Store{userService: service, userAddress: addr, cli: orderClient}
	return store, InitStore(id, store)
}

// InitStore allows for the creation of arbitrary store objects. The main
// use-case is when data id needed that is not in the default Store object.
// The obj argument is anything that would be used to decode json data.
//
// Use type '*map[string]interface{}' in the object argument for all the
// store data
func InitStore(id string, obj interface{}) error {
	path := fmt.Sprintf("/power/store/%s/profile", id)
	b, err := orderClient.get(path, nil)
	if err != nil {
		return err
	}
	return errpair(json.Unmarshal(b, obj), dominosErr(b))
}

var orderClient = &client{Client: http.DefaultClient, host: orderHost}

func initStore(cli *client, id string, store *Store) error {
	path := fmt.Sprintf("/power/store/%s/profile", id)
	b, err := cli.get(path, nil)
	if err != nil {
		return err
	}
	store.cli = cli
	return errpair(json.Unmarshal(b, store), dominosErr(b))
}

type storebuilder struct {
	sync.WaitGroup
	stores chan maybeStore
}

type maybeStore struct {
	store *Store
	err   error
}

func (sb *storebuilder) initStore(cli *client, id string) {
	defer sb.Done()
	path := fmt.Sprintf("/power/store/%s/profile", id)
	store := &Store{}

	b, err := cli.get(path, nil)
	if err != nil {
		sb.stores <- maybeStore{store: nil, err: err}
	}

	err = errpair(json.Unmarshal(b, store), dominosErr(b))
	if err != nil {
		sb.stores <- maybeStore{store: nil, err: err}
	}

	sb.stores <- maybeStore{store: store, err: nil}
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

	menu *Menu
	cli  *client
}

// Menu returns the menu for a store object
func (s *Store) Menu() (*Menu, error) {
	var err error
	if s.menu != nil && s.menu.ID == s.ID {
		return s.menu, nil
	}
	s.menu, err = newMenu(s.cli, s.ID)
	return s.menu, err
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
		cli:           orderClient,
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
		cli:           orderClient,
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

func getNearestStore(c *client, addr Address, service string) (*Store, error) {
	if addr == nil {
		return nil, errors.New("no address")
	}
	locs, err := findNearbyStores(c, addr, service)
	if err != nil {
		return nil, err
	}

	var store = locs.Stores[0]
	for i := range locs.Stores {
		if locs.Stores[i].IsOnlineNow {
			store = locs.Stores[i]
			break
		}
	}
	store.userAddress, store.userService = addr, service
	return store, initStore(c, store.ID, store)
}

var errBadService = errors.New("service must be either 'Delivery' or 'Carryout'")

func findNearbyStores(c *client, addr Address, service string) (*storeLocs, error) {
	if !(service == Delivery || service == Carryout) {
		// panic("service must be either 'Delivery' or 'Carryout'")
		return nil, errBadService
	}
	// TODO: on the dominos website, the c param can sometimes be just the zip code
	// and it still works.
	b, err := c.get("/power/store-locator", &Params{
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
	return locs, dominosErr(b)
}

func asyncNearbyStores(cli *client, addr Address, service string) ([]*Store, error) {
	all, err := findNearbyStores(cli, addr, service)
	if err != nil {
		return nil, err
	}

	var (
		stores  []*Store // return value
		store   *Store
		pair    maybeStore
		builder = storebuilder{
			WaitGroup: sync.WaitGroup{},
			stores:    make(chan maybeStore),
		}
	)

	go func() {
		builder.Wait()
		close(builder.stores)
	}()

	for _, store = range all.Stores {
		builder.Add(1)
		go builder.initStore(cli, store.ID)
	}

	for pair = range builder.stores {
		if pair.err != nil {
			return nil, pair.err
		}
		store = pair.store
		store.userAddress = addr
		store.userService = service
		store.cli = cli

		stores = append(stores, store)
	}

	return stores, nil
}
