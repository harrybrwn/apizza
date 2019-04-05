package dawg

import (
	"errors"
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
	// Options returns a map of the Item's options.
	Options() map[string]interface{}

	// AddToppings adds toppings to the item for when.
	AddTopping(string, string, string) error

	// get the Code from the object
	ItemCode() string

	// get the name of an item
	ItemName() string

	// Type returns the product type of the item.
	Type() string
}

// item has the common fields between Product and Varient.
type item struct {
	Code string
	Name string
	Tags map[string]interface{}

	// Local will tell you if the item was made locally
	Local bool

	menu *Menu // not readlly sure how i feel about this... smells like OOP
}

// ItemCode is a getter method for the Code field.
func (im *item) ItemCode() string {
	return im.Code
}

func (im *item) ItemName() string {
	return im.Name
}

// Product is the structure representing a dominos product. The Product struct
// is meant to instaniated with json data and should be treated as such.
//
// Product is not a the most basic component of the Dominos menu; this is where
// the Variant structure comes in. The Product structure can be seen as more of
// a category that houses a list of Variants.
// All exported field are initialized from json data.
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

	// ProductType is the type of item (ie. 'Bread', 'Pizza'). Used for getting
	// toppings, sides, sizes, or flavors.
	ProductType string

	opts map[string]interface{}
}

// Options returns a map of the Product's options.
func (p *Product) Options() map[string]interface{} {
	codes, amounts, n := splitDefaults(p.DefaultToppings)

	if p.opts == nil {
		p.opts = make(map[string]interface{})
	}

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

// Type returns the product type. see Item
func (p *Product) Type() string {
	return p.ProductType
}

// GetVariants will initialize all the Varients the are a subset of the product.
//
// The function needs a menu to get the data for each variant code.
func (p *Product) GetVariants(container ItemContainer) (varients []*Variant) {
	for _, code := range p.Variants {
		v, err := container.GetVariant(code)
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

// Type returns the product type. see Item
func (v *Variant) Type() string {
	return v.GetProduct().Type()
}

// GetProduct will return the set of variants (Product) that the variant
// is a member of.
func (v *Variant) GetProduct() *Product {
	if v.product == nil {

		return nil
	}
	return v.product
}

// PreConfiguredProduct is pre-configured product.
type PreConfiguredProduct struct {
	item

	// Description of the product
	Description string `json:"Description"`

	// Opts are a string of options that come with the preconfigured-product.
	Opts string `json:"Options"`

	// Size is the size name of the product. It's not a code or anything, its
	// more for user level stuff.
	Size string `json:"Size"`

	// ptype string // just in case there is a way to use this
}

// Options returns a map of the Variant's options.
func (pc *PreConfiguredProduct) Options() map[string]interface{} {
	var opts = map[string]interface{}{}

	codes, amounts, n := splitDefaults(pc.Opts)
	for i := 0; i < n; i++ {
		opts[codes[i]] = map[string]string{ToppingFull: amounts[i]}
	}
	return opts
}

// AddTopping adds a topping to the product.
func (pc *PreConfiguredProduct) AddTopping(code, cover, amnt string) error {
	return errors.New("not implitmented")
}

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

// Type returns the product type. see Item
func (pc *PreConfiguredProduct) Type() string {
	return ""
}

func shortest(a, b []string) int {
	if len(a) > len(b) {
		return len(a)
	}
	return len(b)
}

// interface checks
var (
	_ Item = (*Product)(nil)
	_ Item = (*Variant)(nil)
	_ Item = (*PreConfiguredProduct)(nil)
	_ Item = (*OrderProduct)(nil)
)
