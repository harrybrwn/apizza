package data

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
)

// OrderPrefix is the prefix added to user orders when stored in a database.
const OrderPrefix = "user_order_"

// PrintOrders will print all the names of the saved user orders
func PrintOrders(db cache.MapDB, w io.Writer, verbose bool) error {
	all, err := db.Map()
	if err != nil {
		return err
	}
	out.SetOutput(w)

	var (
		orders    []string
		uOrders   []*dawg.Order
		tempOrder *dawg.Order
	)

	for k, v := range all {
		if strings.Contains(k, OrderPrefix) {
			name := strings.Replace(k, OrderPrefix, "", -1)
			orders = append(orders, name)

			if verbose {
				tempOrder = new(dawg.Order)
				if err = json.Unmarshal(v, tempOrder); err != nil {
					return err
				}
				tempOrder.OrderName = name
				uOrders = append(uOrders, tempOrder)
			}
		}
	}
	if len(orders) < 1 {
		fmt.Fprintln(w, "No orders saved.")
		return nil
	}

	fmt.Fprintln(w, "Your Orders:")
	for i, o := range orders {
		if verbose {
			err = out.PrintOrder(uOrders[i], false)
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintln(w, " ", o)
		}
	}
	return nil
}

// GetOrder will get an order from a database.
func GetOrder(name string, db cache.Getter) (*dawg.Order, error) {
	raw, err := db.Get(OrderPrefix + name)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("cannot find order %s", name)
	}
	order := &dawg.Order{}
	if err = json.Unmarshal(raw, order); err != nil {
		return nil, err
	}
	order.SetName(name)
	return order, nil
}

// SaveOrder will save an order to a database.
func SaveOrder(o *dawg.Order, w io.Writer, db cache.Putter) error {
	raw, err := json.Marshal(o)
	if err != nil {
		return err
	}
	err = db.Put(OrderPrefix+o.Name(), raw)
	if err == nil {
		fmt.Fprintln(w, "order successfully updated.")
	}
	return err
}
