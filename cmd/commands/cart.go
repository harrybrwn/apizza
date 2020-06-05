package commands

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cart"
	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/spf13/cobra"
)

// NewCartCmd creates a new cart command.
func NewCartCmd(b cli.Builder) cli.CliCommand {
	c := &cartCmd{
		cart:       cart.New(b),
		price:      false,
		delete:     false,
		verbose:    false,
		topping:    false,
		color:      Color,
		getaddress: b.Address,
	}

	c.CliCommand = b.Build("cart <order name>", "Manage user created orders", c)
	cmd := c.Cmd()

	cmd.Long = `The cart command gets information on and edit all of the user
created orders.`

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("cannot handle multiple orders")
		}
		return nil
	}
	cmd.ValidArgsFunction = c.cart.OrdersCompletion

	c.Flags().BoolVar(&c.validate, "validate", c.validate, "Send an order to the dominos order-validation endpoint")
	c.Flags().BoolVar(&c.price, "price", c.price, "Show to price of an order")
	c.Flags().BoolVarP(&c.delete, "delete", "d", c.delete, "Delete the order from the database")

	c.Flags().StringSliceVarP(&c.add, "add", "a", c.add, "Add any number of products to a specific order")
	c.Flags().StringVarP(&c.remove, "remove", "r", c.remove, "Remove a product from the order")
	c.Flags().StringVarP(&c.product, "product", "p", "", "Give the product that will be effected by --add or --remove")

	c.Flags().BoolVarP(&c.verbose, "verbose", "v", c.verbose, "Print cart verbosely")

	c.Addcmd(newAddOrderCmd(b))
	return c
}

// `apizza cart`
type cartCmd struct {
	cli.CliCommand
	cart *cart.Cart

	validate bool
	price    bool
	delete   bool
	verbose  bool
	color    bool

	add     []string
	remove  string // yes, you can only remove one thing at a time
	product string

	topping    bool // not actually a flag anymore
	getaddress func() dawg.Address
}

func (c *cartCmd) Run(cmd *cobra.Command, args []string) (err error) {
	c.cart.SetOutput(c.Output())
	if len(args) < 1 {
		var colstr string
		if c.color {
			colstr = "\033[01;34m"
		}
		return c.cart.PrintOrders(c.verbose, colstr)
	}

	if c.topping && c.product == "" {
		return errors.New("must specify an item code with '--product' to edit an order's toppings")
	} else if !c.topping && c.product != "" {
		c.topping = true
	}

	name := args[0]

	if c.delete {
		if err = c.cart.DeleteOrder(name); err != nil {
			return err
		}
		c.Printf("%s successfully deleted.\n", name)
		return nil
	}
	// Set the order that will be used will the cart functions
	if err = c.cart.SetCurrentOrder(name); err != nil {
		return err
	}

	var order *dawg.Order = c.cart.CurrentOrder

	if !order.Address.Equal(c.getaddress()) {
		if err = c.cart.UpdateAddressAndOrderID(c.getaddress()); err != nil {
			return err
		}
	}

	if c.validate {
		// validate the current order and stop
		return c.cart.Validate()
	}

	if len(c.remove) > 0 {
		if c.topping {
			for _, p := range order.Products {
				if _, ok := p.Options()[c.remove]; ok || p.Code == c.product {
					delete(p.Opts, c.remove)
					break
				}
			}
		} else {
			if err = order.RemoveProduct(c.remove); err != nil {
				return err
			}
		}
		return c.cart.SaveAndReset()
	}

	if len(c.add) > 0 {
		if c.topping {
			err = c.cart.AddToppings(c.product, c.add)
		} else {
			err = c.cart.AddProducts(c.add)
		}
		if err != nil {
			return err
		}
		// save order and return early before order is printed out
		return c.cart.SaveAndReset()
	}
	return c.cart.PrintCurrentOrder(true, c.color, c.price)
}

func newAddOrderCmd(b cli.Builder) cli.CliCommand {
	c := &addOrderCmd{name: "", product: ""}
	c.CliCommand = b.Build("new <order name>",
		"Create a new order that will be stored in the cart.", c)
	c.db = b.DB()
	c.StoreFinder = client.NewStoreGetter(b)

	c.Flags().StringVarP(&c.name, "name", "n", c.name, "set the name of a new order")
	c.Flags().StringVarP(&c.product, "product", "p", c.product, "product codes for the new order")
	c.Flags().StringSliceVarP(&c.toppings, "toppings", "t", c.toppings, "toppings for the products being added")
	return c
}

// `apizza cart new` command
type addOrderCmd struct {
	cli.CliCommand
	client.StoreFinder
	db *cache.DataBase

	name     string
	product  string
	toppings []string
}

func (c *addOrderCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if c.name == "" && len(args) < 1 {
		return internal.ErrNoOrderName
	}
	order := c.Store().NewOrder()

	if c.name == "" {
		order.SetName(args[0])
	} else {
		order.SetName(c.name)
	}

	// User interface options:
	// - only add one product but a list of toppings
	// - add a list of products in parallel with a list of toppings (vectorized approach)
	// - add some weird extra syntax to do both (bad idea)
	if c.product != "" {
		prod, err := c.Store().GetVariant(c.product)
		if err != nil {
			return err
		}
		for _, t := range c.toppings {
			if err = internal.AddTopping(t, prod); err != nil {
				return err
			}
		}
		if err = order.AddProduct(prod); err != nil {
			return err
		}
	} else if len(c.toppings) > 0 {
		return errors.New("cannot add just a toppings without products")
	}
	return data.SaveOrder(order, &bytes.Buffer{}, c.db)
}

// NewOrderCmd creates a new order command.
func NewOrderCmd(b cli.Builder) cli.CliCommand {
	c := &orderCmd{
		verbose:    false,
		color:      Color,
		getaddress: b.Address,
	}
	c.CliCommand = b.Build("order", "Send an order from the cart to dominos.", c)
	c.db = b.DB()
	c.Cmd().Long = `The order command is the final destination for an order. This is where
the order will be populated with payment information and sent off to dominos.

The --cvv flag must be specified, and the config file will never store the
cvv. In addition to keeping the cvv safe, payment information will never be
stored the program cache with orders.
`
	c.Cmd().PreRunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("cannot handle multiple orders")
		}
		return nil
	}

	flags := c.Cmd().Flags()
	flags.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "output the order command verbosely")

	flags.StringVar(&c.phone, "phone", "", "Set the phone number that will be used for this order")
	flags.StringVar(&c.email, "email", "", "Set the email that will be used for this order")
	flags.StringVar(&c.fname, "first-name", "", "Set the first name that will be used for this order")
	flags.StringVar(&c.fname, "last-name", "", "Set the last name that will be used for this order")

	flags.IntVar(&c.cvv, "cvv", 0, "Set the card's cvv number for this order")
	flags.StringVar(&c.number, "number", "", "the card number used for orderings")
	flags.StringVar(&c.expiration, "expiration", "", "the card's expiration date")

	flags.BoolVarP(&c.yes, "yes", "y", c.yes, "do not prompt the user with a question")
	flags.BoolVar(&c.logonly, "log-only", false, "")
	flags.MarkHidden("log-only")
	return c
}

// `apizza order`
type orderCmd struct {
	cli.CliCommand
	db *cache.DataBase

	verbose bool
	track   bool

	email, phone string
	fname, lname string
	cvv          int
	number       string
	expiration   string
	yes          bool
	color        bool

	logonly    bool
	getaddress func() dawg.Address
}

func (c *orderCmd) Run(cmd *cobra.Command, args []string) (err error) {
	if len(args) < 1 {
		var colorstr string
		if c.color {
			colorstr = "\033[01;34m"
		}
		return data.PrintOrders(c.db, c.Output(), c.verbose, colorstr)
	} else if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}

	if c.cvv == 0 {
		return errors.New("must have cvv number. (see --cvv)")
	}
	order, err := data.GetOrder(args[0], c.db)
	if err != nil {
		return err
	}

	num := eitherOr(c.number, config.GetString("card.number"))
	exp := eitherOr(c.expiration, config.GetString("card.expiration"))
	if num == "" {
		return errors.New("no card number given")
	}
	if exp == "" {
		return errors.New("no card expiration date given")
	}

	card := dawg.NewCard(num, exp, c.cvv)
	if err = dawg.ValidateCard(card); err != nil {
		return err
	}
	order.AddCard(card)

	names := strings.Split(config.GetString("name"), " ")
	if len(names) >= 1 {
		order.FirstName = eitherOr(c.fname, names[0])
	}
	if len(names) >= 2 {
		order.LastName = eitherOr(c.lname, strings.Join(names[1:], " "))
	}
	order.Email = eitherOr(c.email, config.GetString("email"))
	order.Phone = eitherOr(c.phone, config.GetString("phone"))

	if !order.Address.Equal(c.getaddress()) {
		order.Address = dawg.StreetAddrFromAddress(c.getaddress())
		s, err := dawg.NearestStore(c.getaddress(), order.ServiceMethod)
		if err != nil {
			return err
		}
		order.StoreID = s.ID
	}

	c.Printf("Ordering dominos for %s to %s\n\n", order.ServiceMethod, strings.Replace(obj.AddressFmt(order.Address), "\n", " ", -1))

	if c.logonly {
		log.Println("logging order:", dawg.OrderToJSON(order))
		return nil
	}

	if !c.yes {
		if !internal.YesOrNo(os.Stdin, "Would you like to purchase this order? (y/n)") {
			return nil
		}
	}

	c.Printf("sending order '%s'...\n", order.Name())
	err = order.PlaceOrder()
	// logging happens after so any data from placeorder is included
	log.Println("sending order:", dawg.OrderToJSON(order))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
	c.Printf("sent to %s %s\n", order.Address.LineOne(), order.Address.City())

	if c.verbose {
		if order.ServiceMethod == dawg.Delivery {
			c.Printf("sent by %s to %s %s\n", order.ServiceMethod,
				order.Address.LineOne(), order.Address.City())
		} else {
			c.Printf("sent order for %s\n", order.ServiceMethod)
		}
		c.Printf("%+v\n", order)
	}
	return nil
}

func eitherOr(s1, s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	return s1
}
