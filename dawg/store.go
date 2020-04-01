package dawg

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	// Delivery is a dominos service method that will result
	// in a pizza delivery.
	Delivery = "Delivery"

	// Carryout is a dominos service method that
	// will require users to go and pickup their pizza.
	Carryout = "Carryout"

	profileEndpoint = "/power/store/%s/profile"
)

// ErrBadService is returned if a service is needed but the service validation failed.
var ErrBadService = errors.New("service must be either 'Delivery' or 'Carryout'")

// NearestStore gets the dominos location closest to the given address.
//
// The addr argument should be the address to deliver to not the address of the
// store itself. The service should be either "Carryout" or "Delivery", this will
// determine wether the final order will be for pickup or delivery.
func NearestStore(addr Address, service string) (*Store, error) {
	return getNearestStore(orderClient, addr, service)
}

// GetNearbyStores is a way of getting all the nearby stores
// except they will by full initialized.
func GetNearbyStores(addr Address, service string) ([]*Store, error) {
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
// store data.
//	store := map[string]interface{}{}
//	err := dawg.InitStore(id, &store)
// This will allow all of the fields sent in the api to be viewed.
func InitStore(id string, obj interface{}) error {
	path := fmt.Sprintf(profileEndpoint, id)
	b, err := orderClient.get(path, nil)
	if err != nil {
		return err
	}
	return errpair(json.Unmarshal(b, obj), dominosErr(b))
}

var orderClient = &client{
	host: orderHost,
	Client: &http.Client{
		Timeout:       60 * time.Second,
		CheckRedirect: noRedirects,
		Transport: newRoundTripper(func(req *http.Request) error {
			setDawgUserAgent(req.Header)
			return nil
		}),
	},
}

func initStore(cli *client, id string, store *Store) error {
	path := fmt.Sprintf(profileEndpoint, id)
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
	index int
	err   error
}

func (sb *storebuilder) initStore(cli *client, id string, index int) {
	defer sb.Done()
	path := fmt.Sprintf(profileEndpoint, id)
	store := &Store{}

	b, err := cli.get(path, nil)
	if err != nil {
		sb.stores <- maybeStore{store: nil, err: err, index: -1}
	}

	err = errpair(json.Unmarshal(b, store), dominosErr(b))
	if err != nil {
		sb.stores <- maybeStore{store: nil, err: err, index: -1}
	}

	sb.stores <- maybeStore{store: store, err: nil, index: index}
}

// The Store object represents a physical dominos location.
type Store struct {
	ID string `json:"StoreID"`

	IsOpen          bool
	IsOnlineNow     bool
	IsDeliveryStore bool
	Phone           string

	PaymentTypes    []string `json:"AcceptablePaymentTypes"`
	CreditCardTypes []string `json:"AcceptableCreditCards"`

	Address     string `json:"AddressDescription"`
	PostalCode  string
	City        string
	StreetName  string
	StoreCoords map[string]string `json:"StoreCoordinates"`
	// Min and Max distance
	MinDistance, MaxDistance float64

	ServiceIsOpen        map[string]bool
	ServiceEstimatedWait map[string]struct {
		Min, Max int
	} `json:"ServiceMethodEstimatedWaitMinutes"`
	AllowCarryoutOrders, AllowDeliveryOrders bool

	// Hours describes when the store will be open
	Hours StoreHours
	// ServiceHours describes when the store supports a given service
	ServiceHours map[string]StoreHours

	MinDeliveryOrderAmnt float64 `json:"MinimumDeliveryOrderAmount"`

	Status int

	userAddress Address
	userService string
	menu        *Menu
	cli         *client
}

// StoreHours is a struct that holds Dominos store hours.
type StoreHours struct {
	Sun, Mon, Tue, Wed, Thu, Fri, Sat []struct {
		OpenTime  string
		CloseTime string
	}
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

// GetVariant will get a fully initialized variant from the menu.
func (s *Store) GetVariant(code string) (*Variant, error) {
	menu, err := s.Menu()
	if err != nil {
		return nil, err
	}
	return menu.GetVariant(code)
}

// FindItem is a helper function for Menu.FindItem
func (s *Store) FindItem(code string) (Item, error) {
	menu, err := s.Menu()
	if err != nil {
		return nil, err
	}
	return menu.FindItem(code), nil
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

func findNearbyStores(c *client, addr Address, service string) (*storeLocs, error) {
	if !(service == Delivery || service == Carryout) {
		// panic("service must be either 'Delivery' or 'Carryout'")
		return nil, ErrBadService
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
		nStores = len(all.Stores)
		stores  = make([]*Store, nStores) // return value

		i       int
		store   *Store
		pair    maybeStore
		builder = storebuilder{
			WaitGroup: sync.WaitGroup{},
			stores:    make(chan maybeStore),
		}
	)
	builder.Add(nStores)

	go func() {
		defer close(builder.stores)
		for i, store = range all.Stores {
			go builder.initStore(cli, store.ID, i)
		}

		builder.Wait()
	}()

	for pair = range builder.stores {
		if pair.err != nil {
			return nil, pair.err
		}
		store = pair.store
		store.userAddress = addr
		store.userService = service
		store.cli = cli

		stores[pair.index] = store
	}

	return stores, nil
}
