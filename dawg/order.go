package dawg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

// The Order struct is the main work horse of the api wrapper. The Order struct
// is what will end up being sent to dominos as a json object.
//
// It is suggensted that the order object be constructed from the Store.NewOrder
// method.
type Order struct {
	// LanguageCode is an ISO international language code.
	LanguageCode string `json:"LanguageCode"`

	ServiceMethod string                 `json:"ServiceMethod"`
	Products      []*OrderProduct        `json:"Products"`
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
func (o *Order) AddProduct(item Item) error {
	if item == nil {
		return errors.New("cannot add a null item")
	}
	o.Products = append(o.Products, OPFromItem(item))
	return nil
}

// RemoveProduct will remove the product with a given code from the order.
func (o *Order) RemoveProduct(code string) error {
	var (
		found     = false
		tempProds = []*OrderProduct{}
	)

	for _, p := range o.Products {
		if p.ItemCode() == code {
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

// OrderProduct represents an item that will be sent to and from dominos within
// the Order struct.
type OrderProduct struct {
	item

	// Qty is the number of products to be ordered.
	Qty int `json:"Qty"`

	// ID is the index of the product within an order. Unless the Product is
	// being sent back by dominos, then I have no idea what ID means.
	ID int `json:"ID"`

	IsNew              bool                   `json:"isNew"`
	NeedsCustomization bool                   `json:"NeedsCustomization"`
	Opts               map[string]interface{} `json:"Options"`
	other              map[string]interface{}
	pType              string
}

// OPFromItem will construct an order product from an Item.
func OPFromItem(itm Item) *OrderProduct {
	return &OrderProduct{
		item: item{
			Code: itm.ItemCode(),
			Name: itm.ItemName(),
		},
		Qty:   1,
		Opts:  itm.Options(),
		pType: itm.Category(),
	}
}

// decodes an OrderProduct and puts unrecognized fields into the other map.
func makeProduct(data map[string]interface{}) (*OrderProduct, error) {

	_, file, line, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "Dev Warning: makeProduct is deprecated called on %s:%d", file, line)

	p := &OrderProduct{Qty: 1}
	var md mapstructure.Metadata
	config := &mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   p,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return p, err
	}
	err = decoder.Decode(data)

	other := map[string]interface{}{}
	for _, key := range md.Unused {
		other[key] = data[key]
	}
	p.other = other
	return p, err
}

// Options returns a map of the OrderProdut's options.
func (p *OrderProduct) Options() map[string]interface{} {
	return p.Opts
}

// Category returns the product category of the product
func (p *OrderProduct) Category() string {
	return p.pType
}

// ReadableOptions gives the options that are meant for humas to view.
func (p *OrderProduct) ReadableOptions() map[string]string {
	if p.menu != nil { // this menu that is passed along with item is temporary
		return ReadableToppings(p, p.menu)
	}
	return ReadableOptions(p)
}

// AddTopping adds a topping to the product. The 'code' parameter is a
// topping code, a list of which can be found in the menu object. The 'coverage'
// parameter is for specifieing which side of the topping should be on for
// pizza. The 'amount' parameter is 2.0, 1.5, 1.0, o.5, or 0 and gives the amount
// of topping should be given.
func (p *OrderProduct) AddTopping(code, coverage, amount string) error {
	top := makeTopping(coverage, amount, nil)
	if top == nil {
		return fmt.Errorf("could not make %s topping", code)
	}
	p.Opts[code] = top
	return nil
}

// Size gets the size code of the product. Defaults to -1 if the size
// cannot be found.
func (p *OrderProduct) Size() int64 {
	if v, ok := p.other["SizeCode"]; ok {
		if rt, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
			return rt
		}
	}
	return -1
}

// Price gets the price of the individual product and will return
// -1.0 if the price is not found.
func (p *OrderProduct) Price() float64 {
	if v, ok := p.other["Price"]; ok {
		if rt, err := strconv.ParseFloat(v.(string), 64); err == nil {
			return rt
		}
	}
	return -1.0
}

// Prepared returns a boolean representing whether or not the
// product is prepared. Default is false.
func (p *OrderProduct) Prepared() bool {
	v, ok := p.other["Prepared"]
	if ok {
		return v.(bool)
	}
	return false
}

// Payment just a way to compartmentalize a payment sent to dominos.
type Payment struct {
	Number     string `json:"Number"`
	Expiration string `json:"Expiration"`
	CardType   string `json:"Type"`
	CVV        string `json:"SecurityCode"`
}

// does not take a pointer because ordr.Payments = nil should not be remembered
func getOrderPrice(ordr Order) (map[string]interface{}, error) {
	ordr.Payments = nil
	return sendOrder("/power/price-order", &ordr)
}
