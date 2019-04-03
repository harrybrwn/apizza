package dawg

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

const (
	// ToppingFull is the code sent to dominos to tell them you
	// want a topping that covers the whole pizza.
	ToppingFull = "1/1"

	// ToppingLeft is the code sent to dominos to tell them you
	// want a topping that covers the left side of the pizza
	ToppingLeft = "1/2"

	// ToppingRight is the code sent to dominos to tell them you
	// want a topping that covers the right side of the pizza
	ToppingRight = "2/2"
)

var _ Item = (*OrderProduct)(nil)

// OrderProduct represents an item that will be sent to and from dominos within
// the Order struct.
type OrderProduct struct {
	item

	// Qty is the number of products to be ordered.
	Qty int `json:"Qty"`

	// ID is the index of the product within an order.
	ID int `json:"ID"`

	IsNew              bool                   `json:"isNew"`
	NeedsCustomization bool                   `json:"NeedsCustomization"`
	Opts               map[string]interface{} `json:"Options"`
	other              map[string]interface{}
}

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

// ToOrderProduct converts the OrderProdut into an OrderProduct so that it can
// be sent to dominos in an order.
func (p *OrderProduct) ToOrderProduct() *OrderProduct {
	return p
}

// Options returns a map of the OrderProdut's options.
func (p *OrderProduct) Options() map[string]interface{} {
	return p.Opts
}

// AddTopping adds a topping to the product. The 'code' parameter is a
// topping code which can be found in the menu object. The 'coverage'
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

// Menu represents the dominos menu. It is best if this comes from
// the Store.Menu() method.
type Menu struct {
	ID             string `json:"ID"`
	Categorization struct {
		Food          MenuCategory `json:"Food"`
		Coupons       MenuCategory `json:"Coupons"`
		Preconfigured MenuCategory `json:"PreconfiguredProducts"`
	} `json:"Categorization"`
	Products      map[string]*Product
	Variants      map[string]*Variant
	Toppings      map[string]interface{}
	Preconfigured map[string]*PreConfiguredProduct `json:"PreconfiguredProducts"`
}

// MenuCategory is a category on the dominos menu.
type MenuCategory struct {
	Categories  []MenuCategory `json:"Categories"`
	Products    []string
	Name        string
	Code        string
	Description string
}

// HasItems will return true if the category has items and false if only has
// sub-categories.
func (m MenuCategory) HasItems() bool {
	return len(m.Products) > 0 && len(m.Categories) == 0
}

// IsEmpty returns true when the category has nothing in it.
func (m MenuCategory) IsEmpty() bool {
	return len(m.Products) == 0 && len(m.Categories) == 0
}

// GetProduct find the menu item given a product code.
func (m *Menu) GetProduct(code string) (prod *Product, err error) {
	var ok bool
	if prod, ok = m.Products[code]; ok {
		return prod, nil
	}
	return nil, fmt.Errorf("could not find product '%s'", code)
}

// GetVariant will get a fully initialized varient from the menu.
func (m *Menu) GetVariant(code string) (*Variant, error) {
	if vr, ok := m.Variants[code]; ok {
		vr.product = m.Products[vr.ProductCode]
		return vr, nil
	}
	return nil, fmt.Errorf("could not find variant '%s'", code)
}

func (m *Menu) FindItem(code string) (itm Item) {
	var (
		ok bool
		i  interface{}
	)

	if i, ok = m.Products[code]; ok {
		return i.(*Product)
	} else if i, ok = m.Preconfigured[code]; ok {
		return i.(*PreConfiguredProduct)
	} else if i, ok = m.Variants[code]; ok {
		return i.(*Variant)
	}
	return nil
}

func newMenu(id string) (*Menu, error) {
	path := format("/power/store/%s/menu", id)
	b, err := get(path, Params{"lang": DefaultLang, "structured": "true"})
	if err != nil {
		return nil, err
	}
	menu := &Menu{ID: id}
	if err = json.Unmarshal(b, menu); err != nil {
		return nil, err
	}
	return menu, dominosErr(b)
}
