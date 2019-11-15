package dawg

import (
	"errors"
	"fmt"
)

// SignIn will create a new UserProfile and sign in the account.
func SignIn(username, password string) (*UserProfile, error) {
	a, err := newauth(username, password)
	if err != nil {
		return nil, err
	}
	return a.login()
}

// UserProfile is a Dominos user profile.
type UserProfile struct {
	FirstName  string
	LastName   string
	Email      string
	CustomerID string
	Phone      string
	Addresses  []UserAddress
	Status     int

	auth *auth
}

// AddAddress will add an address to the dominos account.
func (u *UserProfile) AddAddress(a Address) error {
	return errors.New("not implimented")
}

// StoresNearMe will find the stores closest to the user's default address.
func (u *UserProfile) StoresNearMe() ([]*Store, error) {
	return nil, errors.New("not implimented")
}

// NearestStore will find the the store that is closest to the user's default address.
func (u *UserProfile) NearestStore() (*Store, error) {
	return nil, errors.New("not implimented")
}

// UserAddress is an address that is saved by dominos and returned when
// a user signs in.
type UserAddress struct {
	AddressType          string
	StreetNumber         string
	StreetRange          string
	AddressLine2         string
	PropertyType         string
	StreetField2         string
	LocationName         string
	SubNeighborhood      string
	StreetField1         string
	UnitNumber           string
	AddressLine4         string
	PostalCode           string
	BuildingID           string
	IsDefault            bool
	UpdateTime           string
	PropertyNumber       string
	UnitType             string
	Coordinates          map[string]float32
	Neighborhood         string
	Street               string
	CityName             string `json:"City"`
	Region               string
	Name                 string
	StreetName           string
	DeliveryInstructions string
	AddressLine3         string
	CampusID             string
}

// LineOne returns the first line of the address.
func (ua *UserAddress) LineOne() string {
	if len(ua.Street) != 0 {
		return ua.Street
	}
	return fmt.Sprintf("%s %s", ua.StreetNumber, ua.StreetName)
}

// City retuns the city of the address.
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

var _ Address = (*UserAddress)(nil)
