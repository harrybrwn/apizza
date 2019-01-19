package dawg

import (
	"encoding/json"
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

var cachedMenu *Menu

// Product represents a product on the dominos menu.
type Product struct {
	Code    string                 `json:"Code"`
	IsNew   bool                   `json:"isNew"`
	Qty     int                    `json:"Qty"`
	Options map[string]interface{} `json:"Options"`
	Name    string                 `json:"-"`
	Tags    map[string]interface{} `json:"-"`
	other   map[string]interface{}
}

func makeProduct(data map[string]interface{}) (*Product, error) {
	p := &Product{}
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

	if err != nil {
		return p, err
	}
	return p, nil
}

// AddTopping adds a topping to the product. The 'code' parameter is a
// topping code which can be found in the menu object. The 'coverage'
// parameter is for specifieing which side of the topping should be on for
// pizza. The 'amount' parameter is 2.0, 1.5, 1.0, or 0.5 and gives the amount
// of topping should be given.
func (p *Product) AddTopping(code string, coverage string, amount float64) {
	ok := amount == 2.0 || amount == 1.5 || amount == 1.0 || amount == 0.5
	if !ok {
		panic("amount must be 2.0, 1.5, 1.0, or 0.5")
	}
	var key string
	// come on now, this is shity, but i don't know how to fix it
	switch coverage {
	case "full":
		key = ToppingFull
	case "left":
		key = ToppingLeft
	case "right":
		key = ToppingRight
	case ToppingFull:
		key = coverage
	case ToppingLeft:
		key = coverage
	case ToppingRight:
		key = coverage
	default:
		panic(`topping coverage must be "full", "left", or "right".`)
	}
	if p.Options == nil {
		p.Options = map[string]interface{}{}
	}
	p.Options[code] = map[string]interface{}{
		key: strconv.FormatFloat(amount, 'g', 1, 64),
	}
}

// Size gets the size code of the product
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
// product is prepared.
func (p *Product) Prepared() bool {
	v, ok := p.other["pre"].(string)
	if rt, err := strconv.ParseBool(v); err == nil && ok {
		return rt
	}
	return false
}

// Menu represents the dominos menu. It is best if this comes from
// the (*Store).Menu() method.
type Menu struct {
	ID             string
	Products       map[string]interface{} `json:"Products"`
	Variants       map[string]interface{} `json:"Variants"`
	Toppings       map[string]interface{} `json:"Toppings"`
	Categorization map[string]interface{} `json:"Categorization"`
	Preconfigured  map[string]interface{} `json:"PreconfiguredProducts"`
}

func newMenu(id string) (*Menu, error) {
	// checks to if the menu has been cached in memory
	if cachedMenu != nil && cachedMenu.ID == id {
		return cachedMenu, nil
	}
	path := format("/power/store/%s/menu", id)
	b, err := get(path, Params{"lang": Lang, "structured": "true"})
	if err != nil {
		return nil, err
	}
	menu := &Menu{ID: id}
	err = json.Unmarshal(b, menu)
	cachedMenu = menu
	return menu, err
}
