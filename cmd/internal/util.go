package internal

import (
	"fmt"
	"os"
	"strings"
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
