package dawg

import (
	"encoding/json"
	"fmt"
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
		return m.initVariant(vr), nil
	}
	return nil, fmt.Errorf("could not find variant '%s'", code)
}

// FindItem looks in all the different menu categories for an item code given
// as an argument.
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
		return m.initVariant(i.(*Variant))
	}
	return nil
}

// ViewOptions returns a map that makes it easier for humans to view a topping.
func (m *Menu) ViewOptions(itm Item) map[string]string {
	var itmType string

	switch p := itm.(type) {
	case *Variant:
		itmType = p.product.Type
	case *Product:
		itmType = p.Type
	default:
		return nil
	}

	opts := itm.Options()
	tops := m.Toppings[itmType]

	view := map[string]string{}
	for k, v := range opts {
		fmt.Printf("%+v  %v\n", tops[k], v)
		view[k] = tops[k].Name
	}
	return view
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

func (m *Menu) initVariant(v *Variant) *Variant {
	if parent, ok := m.Products[v.ProductCode]; ok {
		v.product = parent
	}
	return v
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
