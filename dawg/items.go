package dawg

import (
	"fmt"
	"strings"
)

var productOptQtys = map[string][]string{
	"Wings":    {"0", "0.5", "1", "1.5", "2", "3", "4", "5"},
	"Pizza":    {"0", "0.5", "1", "1.5", "2"},
	"Sandwich": {"0", "0.5", "1", "1.5"},
	"Pasta":    {"0", "1"},
}

// Item defines an interface for all objects that are items on the dominos menu.
type Item interface {
	// ToOrderProduct converts the Item into an OrderProduct so that it can be
	// sent to dominos in an order.
	ToOrderProduct() *OrderProduct

	// Options returns a map of the Item's options.
	Options() map[string]interface{}

	// AddToppings adds toppings to the item for when.
	AddTopping(string, string, string) error

	// get the Code from the object
	ItemCode() string
}

// item has the common fields between Product and Varient.
type item struct {
	Code string
	Name string
	Tags map[string]interface{}

	// Local will tell you if the item was made locally
	Local bool
}

// ItemCode is a getter method for the Code field.
func (im *item) ItemCode() string {
	return im.Code
}

// Product is the structure representing a dominos product. The Product struct
// is meant to instaniated with json data and should be treated as such.
//
// Product is not a the most basic component of the Dominos menu; this is where
// the Variant structure comes in. The Product structure can be seen as more of
// a category that houses a list of Variants.
type Product struct {
	item

	// Variants is the list of menu items that are a subset of this product.
	Variants []string

	// Short description of the product
	Description string

	// The possible toppings that can be added to this product. Formatted as a
	// a string of comma separated key-value pairs.
	AvailableToppings string

	// The possible sides that can be added to this product. Formatted as a
	// string of comma separated key-value pairs.
	AvailableSides string

	// The default toppings that the product has. Formatted as a string of
	// comma separated key-value pairs.
	DefaultToppings string

	// The default sides that the product has. Formatted as a string of
	// comma separated key-value pairs.
	DefaultSides string

	opts map[string]interface{}
}

// ToOrderProduct converts the Product into an OrderProduct so that it can be
// sent to dominos in an order.
func (p *Product) ToOrderProduct() *OrderProduct {
	return &OrderProduct{
		item: p.item,
		Opts: p.Options(),
		Qty:  1,
		ID:   1,
	}
}

// Options returns a map of the Product's options.
func (p *Product) Options() map[string]interface{} {
	codes, amounts, n := splitDefaults(p.DefaultToppings)
	for i := 0; i < n; i++ {
		// if the default topping is not already in the options then add it
		if _, ok := p.opts[codes[i]]; !ok {
			p.opts[codes[i]] = map[string]string{ToppingFull: amounts[i]}
		}
	}

	return p.opts
}

// AddTopping will add a topping to the product, see Item.
func (p *Product) AddTopping(code, side, amount string) error {
	if p.opts == nil {
		p.opts = make(map[string]interface{})
	}
	top := makeTopping(side, amount, p.optionQtys())
	if top == nil {
		return fmt.Errorf("could not make a %s topping", code)
	}
	p.opts[code] = top
	return nil
}

// GetVariants will initialize all the Varients the are a subset of the product.
//
// The function needs a menu to get the data for each variant code.
func (p *Product) GetVariants(m *Menu) (varients []*Variant) {
	for _, code := range p.Variants {
		v, err := m.GetVariant(code)
		if err != nil {
			continue
		}
		varients = append(varients, v)
	}
	return varients
}

func (p *Product) optionQtys() (optqtys []string) {
	if qtys, ok := p.Tags["OptionQtys"]; ok {
		oq := qtys.([]interface{})
		optqtys = make([]string, len(oq))
		for i, q := range oq {
			optqtys[i] = q.(string)
		}
		return optqtys
	}
	return nil
}

// Variant is a structure that represents a base component of the Dominos menu.
// It will be a subset of a Product (see Product).
type Variant struct {
	item

	// the price of the variant.
	Price string

	// Product Code is the code for the set of variants that the variant belongs
	// to. Will coorespond with the code field of one Product.
	ProductCode string

	// true if the variant is prepared by dominos
	Prepared bool

	product *Product
	opts    map[string]interface{}
}

// ToOrderProduct converts the Variant into an OrderProduct so that it can be
// sent to dominos in an order.
func (v *Variant) ToOrderProduct() *OrderProduct {
	return &OrderProduct{
		item: v.item,
		Opts: v.Options(),
		Qty:  1,
		ID:   1,
	}
}

// Options returns a map of the Variant's options.
func (v *Variant) Options() map[string]interface{} {
	if options, ok := v.Tags["DefaultToppings"]; ok {
		codes, amounts, n := splitDefaults(options.(string))

		if v.opts == nil {
			v.opts = make(map[string]interface{})
		}

		for i := 0; i < n; i++ {
			// if the default topping is not already in the options then add it
			if _, ok := v.opts[codes[i]]; !ok {
				v.opts[codes[i]] = map[string]string{ToppingFull: amounts[i]}
			}
		}
	}
	return v.opts
}

// AddTopping will add a topping to the variant, see Item.
func (v *Variant) AddTopping(code, side, amount string) error {
	if v.opts == nil {
		v.opts = make(map[string]interface{})
	}
	var qtys []string

	if v.product != nil {
		qtys = v.product.optionQtys()
	} else {
		qtys = nil
	}

	top := makeTopping(side, amount, qtys)
	if top == nil {
		return fmt.Errorf("could not make %s topping", code)
	}
	v.opts[code] = top
	return nil
}

// GetProduct will return the set of variants (Product) that the variant
// is a member of.
func (v *Variant) GetProduct() *Product {
	if v.product != nil {
		return v.product
	}
	return nil
}

// PreConfiguredProduct is pre-configured product.
type PreConfiguredProduct struct {
	item

	Description string `json:"Description"`
	Opts        string `json:"Options"`
	Size        string `json:"Size"`
}

// ToOrderProduct converts the Variant into an OrderProduct so that it can be
// sent to dominos in an order.
func (pc *PreConfiguredProduct) ToOrderProduct() *OrderProduct { return nil }

// Options returns a map of the Variant's options.
func (pc *PreConfiguredProduct) Options() map[string]interface{} { return nil }

func splitDefaults(defs string) (keys, vals []string, n int) {
	if defs == "" {
		return nil, nil, 0
	}
	for _, kv := range strings.Split(defs, ",") {
		keyval := strings.Split(kv, "=")
		keys = append(keys, keyval[0])
		vals = append(vals, keyval[1])
	}
	return keys, vals, shortest(keys, vals)
}

func makeTopping(cover, amount string, optionQtys []string) map[string]string {
	var key string

	if !strings.HasSuffix(amount, ".0") && !strings.HasSuffix(amount, ".5") {
		amount += ".0"
	}
	if optionQtys != nil {
		if !validateQtys(amount, optionQtys) {
			return nil
		}
	}

	switch cover {
	case ToppingFull, ToppingLeft, ToppingRight:
		key = cover
	default:
		return nil
	}

	return map[string]string{key: amount}
}

func validateQtys(amount string, qtys []string) bool {
	for _, qty := range qtys {
		if len(qty) == 1 {
			qty += ".0"
		}
		if qty == amount {
			return true
		}
	}
	return false
}

func shortest(a, b []string) int {
	if len(a) > len(b) {
		return len(a)
	}
	return len(b)
}
