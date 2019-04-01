package dawg

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Payment just a way to compartmentalize a payment sent to dominos.
type Payment struct {
	Number     string `json:"Number"`
	Expiration string `json:"Expiration"`
	CardType   string `json:"Type"`
	CVV        string `json:"SecurityCode"`
}

// The Order struct is the main work horse of the api wrapper. The Order struct
// is what will end up being sent to dominos as a json object.
//
// It is suggensted that the order object be constructed from the Store.NewOrder
// method.
type Order struct {
	// LanguageCode-Payments are the fields that are sent to dominos.

	// LanguageCode is an ISO international language code.
	LanguageCode string `json:"LanguageCode"`

	// either
	ServiceMethod string                 `json:"ServiceMethod"`
	Products      []*Product             `json:"Products"`
	StoreID       string                 `json:"StoreID"`
	OrderID       string                 `json:"OrderID"`
	Address       *StreetAddr            `json:"Address"`
	MetaData      map[string]interface{} `json:"metaData"` // only for orders sent back
	FirstName     string                 `json:"FirstName"`
	LastName      string                 `json:"LastName"`
	Payments      []Payment              `json:"Payments"`

	// OrderName is not a field that is sent to dominos
	OrderName string `json:"-"`
	price     float64
}

// PlaceOrder is the method that sends the final order to dominos
func (o *Order) PlaceOrder() error {
	_, err := sendOrder("/power/place-order", o)
	return err
}

// Price method returns the total price of an order.
func (o *Order) Price() (float64, error) {
	data, err := getOrderPrice(*o)
	if err != nil {
		return -1.0, err
	}
	data = data["Order"].(map[string]interface{})
	price, ok := data["Amounts"].(map[string]interface{})["Customer"]
	if !ok {
		return -1.0, errors.New("Price not found")
	}
	return price.(float64), nil
}

// AddProduct adds a product to the Order from a Product Object
func (o *Order) AddProduct(item *Product) {
	o.Products = append(o.Products, item)
}

// RemoveProduct will remove the product with a given code from the order.
func (o *Order) RemoveProduct(code string) error {
	var (
		found     = false
		tempProds = []*Product{}
	)

	for _, p := range o.Products {
		if p.Code == code {
			found = true
			continue
		}
		tempProds = append(tempProds, p)
	}
	if !found {
		return errors.New("product not in order")
	}
	o.Products = tempProds
	return nil
}

// AddPayment adds a payment object to an order
func (o *Order) AddPayment(payment Payment) {
	o.Payments = append(o.Payments, payment)
}

// Name returns the name that was set by the user.
func (o *Order) Name() string {
	return o.OrderName
}

// SetName allows users to name a particular order.
func (o *Order) SetName(name string) {
	o.OrderName = name
}

func (o *Order) rawData() []byte {
	data, err := json.Marshal(o)
	if err != nil {
		return nil
	}
	return []byte(fmt.Sprintf(`{"Order":%s}`, data))
}

func sendOrder(path string, ordr *Order) (map[string]interface{}, error) {
	b, err := post(path, ordr.rawData())
	if err != nil {
		return nil, err
	}
	respData := map[string]interface{}{}
	if err := json.Unmarshal(b, &respData); err != nil {
		return nil, err
	}
	return respData, dominosErr(b)
}

// does not take a pointer because ordr.Payments = nil should not be remembered
func getOrderPrice(ordr Order) (map[string]interface{}, error) {
	ordr.Payments = nil
	return sendOrder("/power/price-order", &ordr)
}
