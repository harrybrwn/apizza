package internal

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/harrybrwn/apizza/dawg"
)

// YesOrNo asks a yes or no question.
func YesOrNo(in *os.File, msg string) bool {
	var res string
	fmt.Printf("%s ", msg)
	_, err := fmt.Fscan(in, &res)
	if err != nil {
		return false
	}

	switch strings.ToLower(res) {
	case "y", "yes", "si":
		return true
	}
	return false
}

// AddTopping parses and adds a topping from the raw string.
//
// formated as <name>:<side>:<amount>
// name is the only one that is required.
func AddTopping(topStr string, p dawg.Item) error {
	var side, amount string

	topping := strings.Split(topStr, ":")

	// assuming strings.Split cannot return zero length array
	if topping[0] == "" || len(topping) > 3 {
		return errors.New("incorrect topping format")
	}

	// TODO: need to check for bad values and use appropriate error handling
	if len(topping) == 1 {
		side = dawg.ToppingFull
	} else if len(topping) >= 2 {
		side = topping[1]
		switch strings.ToLower(side) {
		case "left":
			side = dawg.ToppingLeft
		case "right":
			side = dawg.ToppingRight
		case "full":
			side = dawg.ToppingFull
		default:
			return errors.New("invalid topping side, should be either 'full', 'left', or 'right'")
		}
	}
	amount = "1.0"
	if len(topping) == 3 {
		amount = topping[2]
	}

	switch amount {
	case "1", "2":
		amount += ".0"
	case "0.5", "1.0", "1.5", "2.0":
		break
	default:
		return errors.New("invalid topping amount, should be any of '0.5', '1.0', '1.5', or '2.0'")
	}
	return p.AddTopping(topping[0], side, amount)
}
