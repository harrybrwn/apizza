package data

import (
	"fmt"
	"io"
	"strings"

	"github.com/harrybrwn/apizza/pkg/cache"
)

// OrderPrefix is the prefix added to user orders when stored in a database.
const OrderPrefix = "user_order_"

// PrintOrders will print all the names of the saved user orders
func PrintOrders(db cache.MapDB, w io.Writer) error {
	all, err := db.Map()
	if err != nil {
		return err
	}
	var orders []string

	for k := range all {
		if strings.Contains(k, OrderPrefix) {
			orders = append(orders, strings.Replace(k, OrderPrefix, "", -1))
		}
	}
	if len(orders) < 1 {
		fmt.Fprintln(w, "No orders saved.")
		return nil
	}
	fmt.Fprintln(w, "Your Orders:")
	for _, o := range orders {
		fmt.Fprintln(w, " ", o)
	}
	return nil
}
