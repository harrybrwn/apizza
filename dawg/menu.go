package dawg

import (
	"encoding/json"
	"fmt"
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

// Product represents a product on the dominos menu.
type Product struct {
	Code    string                 `json:"Code"`
	IsNew   bool                   `json:"isNew"`
	Qty     int                    `json:"Qty"`
	Options map[string]interface{} `json:"Options"`
	Name    string                 `json:"Name"`
	Tags    map[string]interface{} `json:"-"`
	other   map[string]interface{}
}

func makeProduct(data map[string]interface{}) (*Product, error) {
	p := &Product{Qty: 1}
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

// AddTopping adds a topping to the product. The 'code' parameter is a
// topping code which can be found in the menu object. The 'coverage'
// parameter is for specifieing which side of the topping should be on for
// pizza. The 'amount' parameter is 2.0, 1.5, 1.0, or 0.5 and gives the amount
// of topping should be given.
func (p *Product) AddTopping(code, coverage string, amount float64) {
	ok := amount == 2.0 || amount == 1.5 || amount == 1.0 || amount == 0.5
	if !ok {
		panic("amount must be 2.0, 1.5, 1.0, or 0.5")
	}
	var key string
	switch coverage {
	case ToppingFull:
		key = coverage
	case ToppingLeft:
		key = coverage
	case ToppingRight:
		key = coverage
	default:
		panic("topping coverage must be dawg.ToppingFull, dawg.ToppingRight, or dawg.ToppingLeft.")
	}
	if p.Options == nil {
		p.Options = map[string]interface{}{}
	}
	p.Options[code] = map[string]interface{}{
		key: strconv.FormatFloat(amount, 'g', 1, 64),
	}
}

// Size gets the size code of the product. Defaults to -1 if the size
// cannot be found.
func (p *Product) Size() int64 {
	if v, ok := p.other["SizeCode"]; ok {
		if rt, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
			return rt
		}
	}
	return -1
}

// Price gets the price of the individual product and will return
// -1.0 if the price is not found.
func (p *Product) Price() float64 {
	if v, ok := p.other["Price"]; ok {
		if rt, err := strconv.ParseFloat(v.(string), 64); err == nil {
			return rt
		}
	}
	return -1.0
}

// Prepared returns a boolean representing whether or not the
// product is prepared. Default is false.
func (p *Product) Prepared() bool {
	v, ok := p.other["Prepared"]
	if ok {
		return v.(bool)
	}
	return false
}

// Menu represents the dominos menu. It is best if this comes from
// the (*Store).Menu() method.
type Menu struct {
	ID             string
	Categorization struct {
		Food          MenuCategory
		Coupons       MenuCategory
		Preconfigured MenuCategory `json:"PreconfiguredProducts"`
	}
	Products      map[string]interface{}
	Variants      map[string]interface{}
	Toppings      map[string]interface{}
	Preconfigured map[string]interface{} `json:"PreconfiguredProducts"`
}

// MenuCategory is a category on the dominos menu.
type MenuCategory struct {
	Categories  []MenuCategory
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
func (m *Menu) GetProduct(code string) (p *Product, err error) {
	if data, ok := m.Variants[code]; ok {
		p, err = makeProduct(data.(map[string]interface{}))
	} else if data, ok := m.Preconfigured[code]; ok {
		p, err = makeProduct(data.(map[string]interface{}))
	} else {
		return nil, fmt.Errorf("could not find %s", code)
	}
	return p, err
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
