package dawg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// TODO: alphabetize the Order struct fields and add some more documentation

// The Order struct is the main work horse of the api wrapper. The Order struct
// is what will end up being sent to dominos as a json object.
//
// It is suggested that the order object be constructed from the Store.NewOrder
// method.
type Order struct {
	// CustomerID is a id for a customer (see UserProfile)
	CustomerID string `json:",omitempty"`
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
	Email         string                 `json:"Email"`
	Phone         string
	Payments      []*orderPayment `json:"Payments"`

	// OrderName is not a field that is sent to dominos, but is just a way for
	// users to name a specific order.
	OrderName string `json:"-"`
	price     float64
	cli       *client
}

// InitOrder will make sure that an order is initialized correctly. An order
// that is not initialized correctly cannot send itself to dominos.
func InitOrder(o *Order) {
	o.cli = orderClient
}

// Init will make sure that an order is initialized correctly. An order
// that is not initialized correctly cannot send itself to dominos.
func (o *Order) Init() {
	o.cli = orderClient
}

// PlaceOrder is the method that sends the final order to dominos
func (o *Order) PlaceOrder() error {
	if err := o.prepare(); err != nil {
		return err
	}
	return sendOrder("/power/place-order", *o)
}

// Price method returns the total price of an order.
func (o *Order) Price() (float64, error) {
	if o.price == 0.0 {
		if err := o.prepare(); err != nil {
			return -1.0, err
		}
	}
	return o.price, nil
}

// AddProduct adds a product to the Order from a Product Object
func (o *Order) AddProduct(item Item) error {
	if item == nil {
		return errors.New("cannot add a nil item")
	}
	o.Products = append(o.Products, OrderProductFromItem(item))
	return nil
}

// AddProductQty adds a product to the Order with a quantity of n.
func (o *Order) AddProductQty(item Item, n int) error {
	if item == nil {
		return errors.New("cannot add a nil item")
	}
	p := OrderProductFromItem(item)
	p.Qty = n
	o.Products = append(o.Products, p)
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
//
// Deprecated. use AddCard
func (o *Order) AddPayment(payment Payment) {
	p := makeOrderPaymentFromCard(&payment)
	o.Payments = append(o.Payments, p)
}

// AddCard will add a card as a method of payment.
func (o *Order) AddCard(c Card) {
	o.Payments = append(o.Payments, makeOrderPaymentFromCard(c))
}

// Name returns the name that was set by the user.
func (o *Order) Name() string {
	return o.OrderName
}

// SetName allows users to name a particular order.
func (o *Order) SetName(name string) {
	o.OrderName = name
}

// Validate sends and order to the validation endpoint to be validated by
// Dominos' servers.
func (o *Order) Validate() error {
	return ValidateOrder(o)
}

// only returns dominos failures or non-dominos errors.
func (o *Order) prepare() error {
	odata, err := getPricingData(*o)
	if err != nil && !IsWarning(err) {
		return err
	}
	o.OrderID = odata.Order.OrderID

	p, ok := odata.Order.Amounts["Customer"]
	if ok {
		o.price = p

		n := len(o.Payments)
		for i := 0; i < n; i++ {
			o.Payments[i].Amount = p
		}
	}
	return nil
}

// ValidateOrder sends and order to the validation endpoint to be validated by
// Dominos' servers.
func ValidateOrder(order *Order) error {
	err := sendOrder("/power/validate-order", *order)
	if IsWarning(err) {
		// TODO: make it possible to recognize the warning as an 'AutoAddedOrderId' warning.
		e := err.(*DominosError)
		order.OrderID = e.Order.OrderID
	}
	return err
}

// OrderToJSON converts an Order to the json string.
func OrderToJSON(o *Order) string {
	s := new(bytes.Buffer)
	err := json.Indent(s, o.raw().Bytes(), "", "    ")
	if err != nil {
		return "{\"error\":\"bad json indentation\"}"
	}
	return s.String()
}

func (o *Order) raw() *bytes.Buffer {
	buf := new(bytes.Buffer)
	err := errpair( // args are executed in order
		eatint(buf.WriteString("{\"Order\":")),
		json.NewEncoder(buf).Encode(o),
	)
	if err != nil {
		return nil
	}
	_, err = buf.WriteString("}")
	if err != nil {
		return nil
	}

	return buf
}

func sendOrder(path string, order Order) error {
	b, err := order.cli.post(path, nil, order.raw())
	if err != nil {
		return err
	}
	return dominosErr(b)
}

func orderRequest(path string, order *Order) (map[string]interface{}, error) {
	b, err := order.cli.post(path, nil, order.raw())
	respData := map[string]interface{}{}

	if err := errpair(err, json.Unmarshal(b, &respData)); err != nil {
		return nil, err
	}
	return respData, dominosErr(b)
}

// does not take a pointer because order.Payments = nil should not be remembered
func getOrderPrice(order Order) (map[string]interface{}, error) {
	// fmt.Println("deprecated... use getPricingData")
	order.Payments = []*orderPayment{}
	return orderRequest("/power/price-order", &order)
}

func getPricingData(order Order) (*priceingData, error) {
	order.Payments = []*orderPayment{}
	b, err := order.cli.post("/power/price-order", nil, order.raw())
	resp := &priceingData{}
	if err := errpair(err, json.Unmarshal(b, resp)); err != nil {
		return nil, err
	}
	return resp, dominosErr(b)
}

type priceingData struct {
	Order pricedOrder
}

type pricedOrder struct {
	OrderID          string
	Amounts          map[string]float64
	AmountsBreakdown map[string]interface{}
	PulseOrderGUID   string `json:"PulseOrderGuid"`
}

// OrderProduct represents an item that will be sent to and from dominos within
// the Order struct.
type OrderProduct struct {
	ItemCommon

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

// OrderProductFromItem will construct an order product from an Item.
func OrderProductFromItem(itm Item) *OrderProduct {
	return &OrderProduct{
		ItemCommon: ItemCommon{
			Code: itm.ItemCode(),
			Name: itm.ItemName(),
		},
		Qty:   1,
		Opts:  itm.Options(),
		pType: itm.Category(),
	}
}

// Options returns a map of the OrderProdut's options.
func (p *OrderProduct) Options() map[string]interface{} {
	return p.Opts
}

// Category returns the product category of the product
func (p *OrderProduct) Category() string {
	return p.pType
}

// ReadableOptions gives the options that are meant for humans to view.
func (p *OrderProduct) ReadableOptions() map[string]string {
	if p.menu != nil { // this menu that is passed along with item is temporary
		return ReadableToppings(p, p.menu)
	}
	return ReadableOptions(p)
}

// AddTopping adds a topping to the product. The 'code' parameter is a
// topping code, a list of which can be found in the menu object. The 'coverage'
// parameter is for specifying which side of the topping should be on for
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
