package dawg

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
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

// ItemContainer defines an interface for objects that get
// Variants and Products.
type ItemContainer interface {
	GetVariant(string) (*Variant, error)
	GetProduct(string) (*Product, error)
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
	Toppings      map[string]map[string]Topping
	Preconfigured map[string]*PreConfiguredProduct `json:"PreconfiguredProducts"`
	Sides         map[string]map[string]struct {
		item
		Description string
	}

	cli *client
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

// GetVariant will get a fully initialized variant from the menu.
func (m *Menu) GetVariant(code string) (*Variant, error) {
	if vr, ok := m.Variants[code]; ok {
		return m.initVariant(vr), nil
	}
	return nil, fmt.Errorf("could not find variant '%s'", code)
}

// FindItem looks in all the different menu categories for an item code given
// as an argument.
func (m *Menu) FindItem(code string) (itm Item) {
	var ok bool
	var i interface{}

	if i, ok = m.Products[code]; ok {
		return i.(*Product)
	} else if i, ok = m.Preconfigured[code]; ok {
		return i.(*PreConfiguredProduct)
	} else if i, ok = m.Variants[code]; ok {
		return m.initVariant(i.(*Variant))
	}
	return nil
}

// Print will write the menu to an io.Writer.
func (m *Menu) Print(w io.Writer) {
	writeMenuCategory(w, m.Categorization.Food, 0)
	writeMenuCategory(w, m.Categorization.Preconfigured, 0)
	writeMenuCategory(w, m.Categorization.Coupons, 0)
}

func writeMenuCategory(w io.Writer, mc MenuCategory, depth int) {
	if mc.IsEmpty() {
		return
	}
	fmt.Fprint(w, strings.Repeat(" ", depth*2), mc.Name, " ", mc.Code)
	fmt.Fprintln(w)

	if mc.HasItems() {
		for _, p := range mc.Products {
			fmt.Fprintln(w, strings.Repeat(" ", (depth*3)-1), p)
		}
	} else {
		for _, submc := range mc.Categories {
			writeMenuCategory(w, submc, depth+1)
		}
	}
}

// ViewOptions returns a map that makes it easier for humans to view a topping.
func (m *Menu) ViewOptions(itm Item) map[string]string {
	return ReadableToppings(itm, m)
}

// Topping is a simple struct that represents a topping on the menu.
//
// Note: this struct does not rempresent a topping that is added to an Item
// and sent to dominos.
type Topping struct {
	item

	Description  string
	Availability []interface{}
}

// ReadableOptions gives an Item's options in a format meant for humas.
func ReadableOptions(item Item) map[string]string {
	var out = map[string]string{}

	for topping, options := range item.Options() {
		out[topping] = translateOpt(options)
	}
	return out
}

// ReadableToppings is the same as ReadableOptions but it looks in the menu for
// the names of toppings instead outputting the topping code.
func ReadableToppings(item Item, m *Menu) map[string]string {
	var (
		out        = map[string]string{}
		toppingSet map[string]Topping
	)

	t := item.Category()
	toppingSet = m.Toppings[t]

	var key string
	for topping, options := range item.Options() {
		key = fmt.Sprintf("%s (%s)", toppingSet[topping].Name, topping)
		out[key] = translateOpt(options)
	}
	return out
}

func translateOpt(opt interface{}) string {
	var param string

	toppingParams, ok := opt.(map[string]string)
	if !ok {
		for k, v := range opt.(map[string]interface{}) {
			param += k + " "
			param += v.(string)
		}
		return param
	}
	for cover, amnt := range toppingParams {
		switch cover {
		case ToppingFull:
			param += "full "
		case ToppingLeft:
			param += "left "
		case ToppingRight:
			param += "right "
		}
		param += amnt
	}
	return param
}

func makeTopping(cover, amount string, optionQtys []string) map[string]string {
	var key string

	if !(strings.HasSuffix(amount, ".0") || strings.HasSuffix(amount, ".5")) {
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

func (m *Menu) initVariant(v *Variant) *Variant {
	if parent, ok := m.Products[v.ProductCode]; ok {
		v.product = parent
	}
	return v
}

func newMenu(c *client, id string) (*Menu, error) {
	path := format("/power/store/%s/menu", id)
	b, err := c.get(path, Params{"lang": DefaultLang, "structured": "true"})
	if err != nil {
		return nil, err
	}
	menu := &Menu{ID: id, cli: c}
	return menu, errpair(json.Unmarshal(b, menu), dominosErr(b))
}
