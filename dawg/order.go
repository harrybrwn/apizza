package dawg

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Payment just a way to compartimentalize a payment sent to dominos.
type Payment struct {
	Number     string `json:"Number"`
	Expiration string `json:"Expiration"`
	CardType   string `json:"Type"`
	CVV        string `json:"SecurityCode"`
}

// The Order struct is the main work horse of the api wrapper. When an Order
// is sent to dominos, it is converted to json and sent as a post requests.
type Order struct {
	LanguageCode  string                 `json:"LanguageCode"`
	ServiceMethod string                 `json:"ServiceMethod"`
	Products      []*Product             `json:"Products"`
	StoreID       string                 `json:"StoreID"`
	OrderID       string                 `json:"OrderID"`
	Address       *StreetAddr            `json:"Address"`
	MetaData      map[string]interface{} `json:"metaData"` // only for orders sent back
	FirstName     string                 `json:"FirstName"`
	LastName      string                 `json:"LastName"`
	Payments      []Payment              `json:"Payments"`
	price         float64
}

// PlaceOrder is the method that sends the final order to dominos
func (o *Order) PlaceOrder(p Payment) {}

// Price method returns the total price of an order.
func (o *Order) Price() (float64, error) {
	data, err := getOrderPrice(*o)
	if err != nil {
		return -1.0, err
	}
	data = data["Order"].(map[string]interface{})
	price, ok := data["Amounts"].(map[string]interface{})["Customer"]
	if ok {
		return price.(float64), nil
	}
	return -1.0, errors.New("Price not found")
}

// AddProduct adds a product to the Order from a Product Object
func (o *Order) AddProduct(item *Product) {
	o.Products = append(o.Products, item)
}

// AddPayment adds a payment object to an order
func (o *Order) AddPayment(payment Payment) {
	o.Payments = append(o.Payments, payment)
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

func getOrderPrice(ordr Order) (map[string]interface{}, error) {
	ordr.Payments = nil
	return sendOrder("/power/price-order", &ordr)
}
