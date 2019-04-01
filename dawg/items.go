package dawg

// Item defines an interface for all objects that are items on the dominos menu.
type Item interface {
	// ToOrderProduct converts the Item into an OrderProduct so that it can be
	// sent to dominos in an order.
	ToOrderProduct() *OrderProduct

	// Options returns a map of the Item's options.
	Options() map[string]interface{}
}

// item has the common fields between Product and Varient.
type item struct {
	Code string
	Name string
	Tags map[string]interface{}

	// Local will tell you if the item was made locally
	Local bool
}

// Options returns a map of the item's options.
func (it *item) Options() map[string]interface{} {
	return it.Tags
}

// Product is the structure representing a dominos product.
//
// Product is not a the most basic component of the Dominos menu; this is where
// the Variant structure comes in. The Product structure can be seen as more of
// a category that houses a list of Variants.
type Product struct {
	item

	// Variants is the list of menu items that are a subset of this product.
	Variants []string

	AvailableToppings string
	AvailableSides    string
	DefaultToppings   string
	DefaultSides      string
	Description       string
}

// ToOrderProduct converts the Product into an OrderProduct so that it can be
// sent to dominos in an order.
func (p *Product) ToOrderProduct() *OrderProduct { return nil }

// Options returns a map of the Product's options.
func (p *Product) Options() map[string]interface{} { return nil }

// Variant is a structure that represents a base component of the Dominos menu.
// It will be a subset of a Product (see Product).
type Variant struct {
	item

	Price       string
	ProductCode string
	Prepared    bool
	product     *Product
}

// ToOrderProduct converts the Variant into an OrderProduct so that it can be
// sent to dominos in an order.
func (v *Variant) ToOrderProduct() *OrderProduct { return nil }

// Options returns a map of the Variant's options.
func (v *Variant) Options() map[string]interface{} { return nil }

// GetProduct will return the set of variants (Product) that the variant
// is a member of.
func (v *Variant) GetProduct() *Product {
	if v.product != nil {
		return v.product
	}
	panic("varient not initialized with a menu")
}
