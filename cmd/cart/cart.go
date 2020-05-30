package cart

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/cmd/opts"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
)

// New will create a new cart
func New(b cartBuilder) *Cart {
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
		out:    DefaultOutput,
		MenuCacher: data.NewMenuCacher(
			opts.MenuUpdateTime,
			b.DB(),
			storefinder.Store,
		),
	}
}

type cartBuilder interface {
	cli.AddrDBBuilder
	cli.StateBuilder
}

var (
	// ErrNoCurrentOrder tells when a method of the cart struct is called
	// that requires the current order to be set but it cannot find one.
	ErrNoCurrentOrder = errors.New("cart has no current order set")

	// ErrOrderNotFound is raised when the cart cannot find the order
	// the it was asked to get.
	ErrOrderNotFound = errors.New("could not find that order")

	// DefaultOutput is the cart package's default output writer.
	DefaultOutput io.Writer = os.Stdout
)

// Cart is an abstraction on the cache.DataBase struct
// representing the user's cart for persistant orders
type Cart struct {
	data.MenuCacher
	// CurrentOrder is only set when SetCurrentOrder is called.
	// Most functions that the cart has will fail if this is nil.
	CurrentOrder *dawg.Order

	db     *cache.DataBase
	finder client.StoreFinder
	out    io.Writer
}

// SetCurrentOrder sets the order that the cart is currently working with.
func (c *Cart) SetCurrentOrder(name string) (err error) {
	c.CurrentOrder, err = c.GetOrder(name)
	return err
}

// SetOutput sets the output of logging messages.
func (c *Cart) SetOutput(w io.Writer) {
	c.out = w
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
	if len(raw) == 0 {
		return nil, ErrOrderNotFound
	}
	order := &dawg.Order{}
	order.Init()
	order.SetName(name)
	order.Address = dawg.StreetAddrFromAddress(c.finder.Address())
	return order, json.Unmarshal(raw, order)
}

// Save will save the current order and reset the current order.
func (c *Cart) Save() error {
	return data.SaveOrder(c.CurrentOrder, c.out, c.db)
}

// SaveAndReset will save the order and set it to nil so that
// it is not accidentally changed.
func (c *Cart) SaveAndReset() error {
	err := c.Save()
	c.CurrentOrder = nil
	return err
}

// ListOrders will return a list of the orders stored in the cart.
func (c *Cart) ListOrders() ([]string, error) {
	mp, err := c.db.Map()
	names := make([]string, 0, len(mp))
	if err != nil {
		return nil, err
	}
	for k := range mp {
		if strings.HasPrefix(k, data.OrderPrefix) {
			names = append(names, strings.ReplaceAll(k, data.OrderPrefix, ""))
		}
	}
	return names, nil
}

// OrdersCompletion is a cobra valide args function for getting order names.
func (c *Cart) OrdersCompletion(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	names, err := c.ListOrders()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// Validate the current order
func (c *Cart) Validate() error {
	if c.CurrentOrder == nil {
		return ErrNoCurrentOrder
	}
	fmt.Fprintf(c.out, "validating order '%s'...\n", c.CurrentOrder.Name())
	err := c.CurrentOrder.Validate()
	if dawg.IsWarning(err) {
		goto ValidOrder
	}
	if err != nil {
		return err
	}

ValidOrder:
	fmt.Fprintln(c.out, "Order is ok.")
	return nil
}

// PrintCurrentOrder will print out the current order.
func (c *Cart) PrintCurrentOrder(full, color, price bool) error {
	out.SetOutput(c.out)
	return out.PrintOrder(c.CurrentOrder, full, color, price)
}

// UpdateAddressAndOrderID will update the current order's address and then update
// the current order's StoreID by finding the nearest store for that address.
func (c *Cart) UpdateAddressAndOrderID(currentAddr dawg.Address) error {
	c.CurrentOrder.Address = dawg.StreetAddrFromAddress(currentAddr)
	s, err := dawg.NearestStore(currentAddr, c.CurrentOrder.ServiceMethod)
	if err != nil {
		return err
	}
	c.CurrentOrder.StoreID = s.ID
	return nil
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
	if c.CurrentOrder == nil {
		return ErrNoCurrentOrder
	}
	return addToppingsToOrder(c.CurrentOrder, product, toppings)
}

// AddProducts adds a list of products to the current order
func (c *Cart) AddProducts(products []string) error {
	if c.CurrentOrder == nil {
		return ErrNoCurrentOrder
	}
	if err := c.db.UpdateTS("menu", c); err != nil {
		return err
	}
	return addProducts(c.CurrentOrder, c.Menu(), products)
}

// PrintOrders will print out all the orders saved in the database
func (c *Cart) PrintOrders(verbose bool, color string) error {
	return data.PrintOrders(c.db, c.out, verbose, color)
}

func addToppingsToOrder(o *dawg.Order, product string, toppings []string) (err error) {
	if product == "" {
		return errors.New("what product are these toppings being added to")
	}
	for _, top := range toppings {
		p := getOrderItem(o, product)
		if p == nil {
			return fmt.Errorf("cannot find '%s' in the '%s' order", product, o.Name())
		}

		err = internal.AddTopping(top, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func addProducts(o *dawg.Order, menu *dawg.Menu, products []string) (err error) {
	var itm dawg.Item
	for _, newP := range products {
		itm, err = menu.GetVariant(newP)
		if err != nil {
			return err
		}
		err = o.AddProduct(itm)
		if err != nil {
			return err
		}
	}
	return nil
}

func getOrderItem(order *dawg.Order, code string) dawg.Item {
	for _, itm := range order.Products {
		if itm.ItemCode() == code {
			return itm
		}
	}
	return nil
}
