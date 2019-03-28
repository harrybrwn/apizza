package out

import (
	"io"
	"os"

	"github.com/harrybrwn/apizza/dawg"
)

var output io.Writer = os.Stdout

func init() {
	if output == nil {
		panic("please run out.SetOutput")
	}
}

// SetOutput gives the out package an output variable.
func SetOutput(w io.Writer) {
	output = w
}

// PrintOrder will print the order given.
func PrintOrder(o *dawg.Order, full bool) error {
	var t string

	if full {
		t = defaultOrderTmpl
	} else {
		t = cartOrderTmpl
	}
	return tmpl(output, t, o)
}
