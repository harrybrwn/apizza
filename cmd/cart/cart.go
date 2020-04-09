package cart

import (
	"errors"
	"fmt"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/opts"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
)

// New will create a new cart
func New(b cli.Builder) *Cart {
	storefinder := client.NewStoreGetterFunc(func() string {
		opts := b.GlobalOptions()
		if opts.Service != "" {
			return opts.Service
		}
		return b.Config().Service
	}, b.Address)

	return &Cart{
		db:     b.DB(),
		finder: storefinder,
		cacher: data.NewMenuCacher(
			opts.MenuUpdateTime,
			b.DB(),
			storefinder.Store,
		),
	}
}

// ErrNoCurrentOrder tells when a method of the cart struct is called
// that requires the current order to be set but it cannot find one.
var ErrNoCurrentOrder = errors.New("cart has no current order set")

// Cart is an abstraction on the cache.DataBase struct
// representing the user's cart for persistant orders
type Cart struct {
	data.MenuCacher
	db     cache.DB
	finder client.StoreFinder
	cacher data.MenuCacher

	currentName  string
	currentOrder *dawg.Order
}

// SetCurrentOrder sets the order that the cart is currently working with.
func (c *Cart) SetCurrentOrder(name string) (err error) {
	c.currentName = name
	c.currentOrder, err = c.GetOrder(name)
	return err
}

// DeleteOrder will delete an order from the database.
func (c *Cart) DeleteOrder(name string) error {
	return c.db.Delete(data.OrderPrefix + name)
}

// GetOrder will get an order from the database.
func (c *Cart) GetOrder(name string) (*dawg.Order, error) {
	raw, err := c.db.Get(data.OrderPrefix + name)
	if err != nil {
		return nil, err
	}
	order := &dawg.Order{}
	order.Init()
	order.SetName(name)
	order.Address = c.finder.Address()
	return order, nil
}

// Validate the current order
func (c *Cart) Validate() error {
	if c.currentOrder == nil {
		return ErrNoCurrentOrder
	}
	err = c.currentOrder.Validate()
	if dawg.IsWarning(err) {
		return nil
	}
	return err
}

// ValidateOrder will retrieve an order from the database and validate it.
func (c *Cart) ValidateOrder(name string) error {
	o, err := c.GetOrder(name)
	if err != nil {
		return err
	}
	err = o.Validate()
	if dawg.IsWarning(err) {
		return nil
	}
	return err
}

// AddToppings will add toppings to a product in the current order.
func (c *Cart) AddToppings(product string, toppings []string) error {
	if c.currentOrder == nil {
		return ErrNoCurrentOrder
	}
	return addToppingsToOrder(c.currentOrder, product, toppings)
}

// AddToppingsToOrder will get an order from the database and add toppings
// to a product in that order.
func (c *Cart) AddToppingsToOrder(name, product string, toppings []string) error {
	order, err := c.GetOrder(name)
	if err != nil {
		return err
	}
	return addToppingsToOrder(order, product, toppings)
}

// AddProducts adds a list of products to the current order
func (c *Cart) AddProducts(products []string) error {
	if c.currentOrder == nil {
		return ErrNoCurrentOrder
	}
	if err := c.db.UpdateTS("menu", c.cacher); err != nil {
		return err
	}
	return addProducts(c.currentOrder, c.Menu(), products)
}

// AddProductsToOrder adds a list of products to an order that needs to
// be retrived from the database.
func (c *Cart) AddProductsToOrder(name string, products []string) error {
	if err := c.db.UpdateTS("menu", c.cacher); err != nil {
		return err
	}
	order, err := c.GetOrder(name)
	if err != nil {
		return err
	}
	menu := c.Menu()
	return addProducts(order, menu, products)
}

func addToppingsToOrder(o *dawg.Order, product string, toppings []string) error {
	if product == "" {
		return errors.New("what product are these toppings being added to")
	}
	for _, top := range toppings {
		p := getOrderItem(order, product)
		if p == nil {
			return fmt.Errorf("cannot find '%s' in the '%s' order", product, order.Name())
		}

		err = addTopping(top, p)
		if err != nil {
			return err
		}
	}
}

func addProducts(o *dawg.Order, menu *dawg.Menu, products []string) error {
	var itm dawg.Item
	for _, newP := range products {
		itm, err = menu.GetVariant(newP)
		if err != nil {
			return err
		}
		err = order.AddProduct(itm)
		if err != nil {
			return err
		}
	}
}

func getOrderItem(order *dawg.Order, code string) dawg.Item {
	for _, itm := range order.Products {
		if itm.ItemCode() == code {
			return itm
		}
	}
	return nil
}

// adds a topping.
//
// formated as <name>:<side>:<amount>
// name is the only one that is required.
func addTopping(topStr string, p dawg.Item) error {
	var side, amount string

	topping := strings.Split(topStr, ":")

	if len(topping) < 1 {
		return errors.New("incorrect topping format")
	}

	if len(topping) == 1 {
		side = dawg.ToppingFull
	} else if len(topping) >= 2 {
		side = topping[1]
	}

	if len(topping) == 3 {
		amount = topping[2]
	} else {
		amount = "1.0"
	}
	p.AddTopping(topping[0], side, amount)
	return nil
}
