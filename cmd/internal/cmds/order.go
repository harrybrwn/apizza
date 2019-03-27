package cmds

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	storeID = ""
)

func init() {
	OrderCmd.RunE = runOrder
}

// OrderCmd is the order command
var OrderCmd = &cobra.Command{
	Use: "order", Short: "Send an order from the cart to dominos.",
}

func runOrder(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Println("your orders will show up here")
	} else if len(args) > 1 {
		return errors.New("cannot handle multiple orders")
	}

	if storeID != "" {
		cmd.Println("i told you not to use the store-id flag!")
	}
	return errors.New("not finished with this command")
}

func init() {
	OrderCmd.Flags().StringVar(&storeID, "store-id", "", "actually, this flag does nothing so don't use it")
}
