package out

import (
	"errors"
	"io"
	"os"

	"github.com/harrybrwn/apizza/dawg"
)

var output io.Writer = os.Stdout

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

// PrintVariant will display a dawg.Variant in a pretty way.
func PrintVariant(v *dawg.Variant, verbose bool) error {
	var template string
	if verbose {
		template = variantTmpl
	} else {
		return errors.New("none verbose mode not supported")
	}

	return tmpl(output, template, v)
}
