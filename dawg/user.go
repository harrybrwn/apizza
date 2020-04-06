package dawg

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SignIn will create a new UserProfile and sign in the account.
func SignIn(username, password string) (*UserProfile, error) {
	a, err := newauth(username, password)
	if err != nil {
		return nil, err
	}
	return a.login()
}

// TODO: find out how to update a profile on domino's end

// UserProfile is a Dominos user profile.
type UserProfile struct {
	FirstName string
	LastName  string
	Phone     string

	// Type of dominos account
	Type string
	// Dominos internal user id
	CustomerID string
	// Identifiers are the pieces of information used to identify the
	// user (even if they are not signed in)
	Identifiers []string `json:"CustomerIdentifiers"`
	// AgreedToTermsOfUse is true if the user has agreed to dominos terms of use.
	AgreedToTermsOfUse bool `json:"AgreeToTermsOfUse"`
	// User's gender
	Gender string
	// List of all the addresses saved in the dominos account
	Addresses []*UserAddress

	// Email is the user's email address
	Email string
	// EmailOptIn tells wether the user is opted in for email updates or not
	EmailOptIn bool
	// EmailOptInTime shows what time the user last opted in for email updates
	EmailOptInTime string

	// SmsPhone is the phone number used for sms updates
	SmsPhone string
	// SmsOptIn tells wether the use is opted for sms updates or not
	SmsOptIn bool
	// SmsOptInTime shows the last time the user opted in for sms updates
	SmsOptInTime string

	// UpdateTime shows the last time the user's profile was updated
	UpdateTime string

	// ServiceMethod should be "Delivery" or "Carryout"
	ServiceMethod string `json:"-"` // this is a package specific field (not from the api)
	ordersMeta    *customerOrders

	auth        *auth
	store       *Store
	loyaltyData *CustomerLoyalty
}

// AddAddress will add an address to the dominos account.
func (u *UserProfile) AddAddress(a Address) {
	// TODO: consider sending a request to dominos to update the user with this address.
	// this can be done in a separate go-routine
	u.Addresses = append(u.Addresses, UserAddressFromAddress(a))
}

var (
	errNoServiceMethod     = errors.New("no service method given")
	errUserNoServiceMethod = errors.New("User has no ServiceMethod set")
)

// StoresNearMe will find the stores closest to the user's default address.
func (u *UserProfile) StoresNearMe() ([]*Store, error) {
	if u.ServiceMethod == "" {
		return nil, errUserNoServiceMethod
	}
	if err := u.addressCheck(); err != nil {
		return nil, err
	}
	return asyncNearbyStores(u.auth.cli, u.DefaultAddress(), u.ServiceMethod)
}

// NearestStore will find the the store that is closest to the user's default address.
func (u *UserProfile) NearestStore(service string) (*Store, error) {
	var err error
	if u.store != nil {
		return u.store, nil
	}

	// Pass the authorized user's client along to the
	// store which will use the user's credentials
	// on each request.
	c := &client{host: orderHost, Client: u.auth.cli.Client}
	if err = u.addressCheck(); err != nil {
		return nil, err
	}
	u.store, err = getNearestStore(c, u.DefaultAddress(), service)
	return u.store, err
}

// DefaultAddress will return the address that Dominos has marked as the default.
// If dominos has not marked any of them as the default, the
// the first one will be returned and nil if there are no addresses.
func (u *UserProfile) DefaultAddress() *UserAddress {
	if len(u.Addresses) == 0 || u.Addresses == nil {
		return nil
	}

	for _, a := range u.Addresses {
		if a.IsDefault {
			return a
		}
	}
	return u.Addresses[0]
}

// SetServiceMethod will set the user's default service method,
// should be "Delivery" or "Carryout"
func (u *UserProfile) SetServiceMethod(service string) error {
	if !(service == Delivery || service == Carryout) {
		return ErrBadService
	}
	u.ServiceMethod = service
	return nil
}

// SetStore will set the UserProfile struct's internal store variable.
func (u *UserProfile) SetStore(store *Store) error {
	if store == nil {
		return errors.New("cannot set UserProfile store to a nil value")
	}
	if store.ID == "" {
		return errors.New("UserProfile.SetStore: store is uninitialized")
	}
	u.store = store
	return nil
}

// TODO: write tests for GetCards, Loyalty, PreviousOrders, GetEasyOrder, initOrdersMeta, and customerEndpoint

// GetCards will get the cards that Dominos has saved in their database. (see UserCard)
func (u *UserProfile) GetCards() ([]*UserCard, error) {
	cards := make([]*UserCard, 0)
	return cards, u.customerEndpoint("card", nil, &cards)
}

// Loyalty returns the user's loyalty meta-data (see CustomerLoyalty)
func (u *UserProfile) Loyalty() (*CustomerLoyalty, error) {
	u.loyaltyData = new(CustomerLoyalty)
	return u.loyaltyData, u.customerEndpoint("loyalty", nil, u.loyaltyData)
}

// for internal use (caches the loyalty data)
func (u *UserProfile) getLoyalty() (*CustomerLoyalty, error) {
	if u.loyaltyData != nil {
		return u.loyaltyData, nil
	}
	return u.Loyalty()
}

// PreviousOrders will return `n` of the user's previous orders.
func (u *UserProfile) PreviousOrders(n int) ([]*EasyOrder, error) {
	return u.ordersMeta.CustomerOrders, u.initOrdersMeta(n)
}

// GetEasyOrder will return the user's easy order.
func (u *UserProfile) GetEasyOrder() (*EasyOrder, error) {
	var err error
	if u.ordersMeta == nil {
		if err = u.initOrdersMeta(3); err != nil {
			return nil, err
		}
	}
	return u.ordersMeta.EasyOrder, nil
}

// Orders returns a variety of meta-data on the user's previous and saved orders.
func (u *UserProfile) initOrdersMeta(limit int) error {
	u.ordersMeta = &customerOrders{}
	return u.customerEndpoint(
		"order",
		Params{"limit": limit, "lang": DefaultLang},
		&u.ordersMeta,
	)
}

// NewOrder will create a new *dawg.Order struct with all of the user's information.
func (u *UserProfile) NewOrder() (*Order, error) {
	var err error
	if u.store == nil {
		_, err = u.NearestStore(u.ServiceMethod)
		if err != nil {
			return nil, err
		}
	}
	order := &Order{
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Email:         u.Email,
		LanguageCode:  DefaultLang,
		ServiceMethod: u.ServiceMethod,
		StoreID:       u.store.ID,
		CustomerID:    u.CustomerID,
		Phone:         u.Phone,
		Products:      []*OrderProduct{},
		Address:       StreetAddrFromAddress(u.store.userAddress),
		Payments:      []*orderPayment{},
		cli:           u.auth.cli,
	}
	return order, nil
}

// returns and error if the user has no address
//
// addressCheck is meant to be a check before DefaultAddress is called internally.
func (u *UserProfile) addressCheck() error {
	if len(u.Addresses) == 0 {
		return errors.New("UserProfile has no addresses")
	}
	return nil
}

func (u *UserProfile) serviceCheck() error {
	if u.ServiceMethod == "" {
		return ErrNoUserService
	}
	return nil
}

func (u *UserProfile) customerEndpoint(
	path string,
	params Params,
	obj interface{},
) error {
	if u.CustomerID == "" {
		return errors.New("UserProfile not fully initialized: needs CustomerID")
	}
	if params == nil {
		params = make(Params)
	}
	params["_"] = time.Now().Nanosecond()

	return u.auth.cli.dojson(obj, &http.Request{
		Method: "GET",
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
		URL: &url.URL{
			Scheme:   "https",
			Host:     u.auth.cli.host,
			Path:     fmt.Sprintf("/power/customer/%s/%s", u.CustomerID, path),
			RawQuery: params.Encode(),
		},
	})
}

// UserAddress is an address that is saved by dominos and returned when
// a user signs in.
type UserAddress struct {
	// Basic address fields
	Street       string
	StreetName   string
	StreetNumber string
	CityName     string `json:"City"`
	Region       string
	PostalCode   string
	AddressType  string

	// Dominos specific meta-data
	Name                 string
	IsDefault            bool
	DeliveryInstructions string
	UpdateTime           string

	// Other address specific fields
	AddressLine2 string
	AddressLine3 string
	AddressLine4 string
	StreetField1 string
	StreetField2 string
	StreetRange  string

	// Other rarely-used address meta-data fields
	PropertyType    string
	PropertyNumber  string
	UnitType        string
	UnitNumber      string
	BuildingID      string
	CampusID        string
	Neighborhood    string
	SubNeighborhood string
	LocationName    string
	Coordinates     map[string]float32
}

var _ Address = (*UserAddress)(nil)

// UserAddressFromAddress converts an address to a UserAddress.
func UserAddressFromAddress(a Address) *UserAddress {
	var streetNum, streetName string
	parts := strings.Split(a.LineOne(), " ")

	if _, err := strconv.Atoi(parts[0]); err == nil {
		streetNum = parts[0]
	}
	streetName = strings.Join(parts[1:], " ")

	if addr, ok := a.(*UserAddress); ok {
		if len(addr.StreetNumber) == 0 {
			addr.StreetNumber = streetNum
		}
		if len(addr.StreetName) == 0 {
			addr.StreetName = streetName
		}
		return addr
	}

	return &UserAddress{
		Street:       a.LineOne(),
		StreetNumber: streetNum,
		StreetName:   streetName,
		CityName:     a.City(),
		PostalCode:   a.Zip(),
		Region:       a.StateCode(),
	}
}

// LineOne returns the first line of the address.
func (ua *UserAddress) LineOne() string {
	if len(ua.Street) != 0 {
		return ua.Street
	}
	return fmt.Sprintf("%s %s", ua.StreetNumber, ua.StreetName)
}

// City returns the city of the address.
func (ua *UserAddress) City() string {
	return ua.CityName
}

// StateCode returns the region of the address.
func (ua *UserAddress) StateCode() string {
	return ua.Region
}

// Zip returns the postal code.
func (ua *UserAddress) Zip() string {
	return ua.PostalCode
}

// UserCard holds the card data that Dominos stores and send back to users.
// For security reasons, Dominos does not send the raw card number or the
// raw security code. Insted they send a card ID that is used to reference
// that card. This allows users to reference that card using the card's
// nickname (see 'NickName' field)
type UserCard struct {
	// ID is the card id used by dominos internally to reference a user's
	// card without sending the actual card number over the internet
	ID string `json:"id"`
	// NickName is the cards name, given when a user saves a card as a named card
	NickName  string `json:"nickName"`
	IsDefault bool   `json:"isDefault"`

	TimesCharged        int  `json:"timesCharged"`
	TimesChargedIsValid bool `json:"timesChargedIsValid"`

	// LastFour is a field that gives the last four didgets of the card number
	LastFour string `json:"lastFour"`
	// true if the card has expired
	IsExpired       bool `json:"isExpired"`
	ExpirationMonth int  `json:"expirationMonth"`
	ExpirationYear  int  `json:"expirationYear"`
	// LastUpdated shows the date that this card was last updated in
	// dominos databases
	LastUpdated string `json:"lastUpdated"`

	CardType   string `json:"cardType"`
	BillingZip string `json:"billingZip"`
}

// CustomerLoyalty is a struct that holds account meta-data used by
// Dominos to keep track of customer rewards.
type CustomerLoyalty struct {
	CustomerID              string
	EnrollDate              string
	LastActivityDate        string
	BasePointExpirationDate string
	PendingPointBalance     string
	AccountStatus           string
	// VestedPointBalance is the points you have
	// saved up in order to get a free pizza.
	VestedPointBalance int
	// This is a list of possible coupons that a
	// customer can receive.
	LoyaltyCoupons []struct {
		CouponCode    string
		PointValue    int
		BaseCoupon    bool
		LimitPerOrder string
	}
}

// TODO: figure out how the dominos website sends an easy order to the servers

type customerOrders struct {
	CustomerOrders []*EasyOrder            `json:"customerOrders"`
	EasyOrder      *EasyOrder              `json:"easyOrder"`
	Products       map[string]OrderProduct `json:"products"`

	ProductsByFrequencyRecency []struct {
		ProductKey string `json:"productKey"`
		Frequency  int    `json:"frequency"`
	} `json:"productsByFrequencyRecency"`

	ProductsByCategory []struct {
		Category    string   `json:"category"`
		ProductKeys []string `json:"productKeys"`
	} `json:"productsByCategory"`
}

// EasyOrder is an easy order.
type EasyOrder struct {
	AddressNickName      string `json:"addressNickName"`
	OrderNickName        string `json:"easyOrderNickName"`
	EasyOrder            bool   `json:"easyOrder"`
	ID                   string `json:"id"`
	DeliveryInstructions string `json:"deliveryInstructions"`
	Cards                []struct {
		ID       string `json:"id"`
		NickName string `json:"nickName"`
	}

	Store struct {
		Address              *StreetAddr `json:"address"`
		CarryoutServiceHours string      `json:"carryoutServiceHours"`
		DeliveryServiceHours string      `json:"deliveryServiceHours"`
	} `json:"store"`
	Order previousOrder `json:"order"`
}

type previousOrder struct {
	Order
	pricedOrder

	Partners            interface{}
	StoreOrderID        string
	StorePlaceOrderTime string
	OrderMethod         string
	IP                  string

	PlaceOrderTime string // YYYY-MM-DD H:M:S
	BusinessDate   string // YYYY-MM-DD

	OrderInfoCollection []interface{}
}
