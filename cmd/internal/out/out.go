package out

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/harrybrwn/apizza/dawg"
)

var (
	output  io.Writer = os.Stdout
	_output           = os.Stdout // don't change this
)

const space = ' '

// SetOutput gives the out package an output variable.
func SetOutput(w io.Writer) {
	output = w
}

// ResetOutput will reset the package output to it's default io.Writer
func ResetOutput() {
	output = _output
}

// FormatLine will take a string and make sure it does not cross a certain lenth
// by slicing it at a space closest to the length argument.
func FormatLine(str string, length int) (lines []string) {
	strLen := utf8.RuneCountInString(str)

	if strLen <= length {
		return []string{str}
	}
	var head = 0

	for head < strLen {
		l, end := lineone(str, head, length)
		lines = append(lines, l)
		head += end
	}
	return lines
}

// FormatLineIndent cuts a string at the linelen and indents the string that
// is left over.
func FormatLineIndent(str string, linelen, tablen int) (formatted string) {
	lines := FormatLine(str, linelen)
	var line string
	for i := range lines {
		if i != 0 {
			line = strings.Repeat(" ", tablen) + lines[i]
		} else {
			line = lines[i]
		}
		if i != len(lines)-1 {
			line += "\n"
		}

		formatted += line
	}
	return formatted
}

func lineone(str string, start, length int) (string, int) {
	str = str[start:]
	if l := len(str); l < length {
		return str, l
	}

	var i int
	for i = length; i >= 0; i-- {
		if str[i] == space {
			i++
			break
		}
	}
	return str[:i], i
}

// PrintOrder will print the order given.
func PrintOrder(o *dawg.Order, full bool) error {
	var t string

	if full {
		t = defaultOrderTmpl
	} else {
		t = cartOrderTmpl
	}
	// data := struct {
	// 	*dawg.Order
	// 	Options map[string]string
	// }{
	// 	Order: o,
	// 	Options:
	// }

	return tmpl(output, t, o)
}

// ItemInfo prints the common information for an Item
func ItemInfo(i dawg.Item, menu *dawg.Menu) {
	fmt.Fprintf(output, "%s\n", i.ItemName())
	fmt.Fprintf(output, "  Code: %s\n", i.ItemCode())
	if c := i.Category(); c != "" {
		fmt.Fprintf(output, "  Category: %s\n", c)
	}
	if len(i.Options()) > 0 {
		fmt.Fprintln(output, "  Toppings:")
		tops := dawg.ReadableToppings(i, menu)
		for tname, param := range tops {
			fmt.Fprintf(output, "    %s:%s%s\n", tname, " ", param)
		}
	}
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

// PrintProduct will print a dawg.Product
func PrintProduct(p *dawg.Product) error {
	data := struct {
		*dawg.Product
		Description string
	}{Product: p, Description: FormatLineIndent(p.Description, 70, 16)}
	return tmpl(output, productTmpl, data)
}
