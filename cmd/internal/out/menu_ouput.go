package out

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/harrybrwn/apizza/dawg"
)

// PrintMenu is the function that prints out a menu category and any of its
// sub-categories.
func PrintMenu(cat dawg.MenuCategory, depth int, m *dawg.Menu) error {
	if cat.IsEmpty() {
		return nil
	}
	fmt.Fprintf(output, "\n%s%s%s%s\n", spaces(depth*2), strings.Repeat("-", 8),
		cat.Name, strings.Repeat("-", 60-len(cat.Name)-(depth*2)))
	if cat.HasItems() {
		for _, p := range cat.Products {
			printCategory(p, depth+1, m)
		}
	} else {
		for _, category := range cat.Categories {
			PrintMenu(category, depth+1, m)
		}
	}
	return nil
}

// ItemInfo prints the common information for an Item.
// Used by the menu command.
func ItemInfo(prod dawg.Item, menu *dawg.Menu) error {
	o := &bytes.Buffer{}
	iteminfo(prod, menu)

	switch p := prod.(type) {
	case *dawg.Variant:
		fmt.Fprintf(o, "  Price: %s\n", p.Price)
		parent := p.GetProduct()
		if parent == nil {
			break
		}
		fmt.Fprintf(o, "  Parent Product: '%s' [%s]\n", parent.ItemName(), parent.ItemCode())

	case *dawg.PreConfiguredProduct:
		fmt.Fprintf(o, "  Description: '%s'\n", FormatLineIndent(p.Description, 70, 16))
		fmt.Fprintf(o, "  Size: %s\n", p.Size)

	case *dawg.Product:
		PrintProduct(p)
	}
	_, err := output.Write(o.Bytes())
	return err
}

func iteminfo(i dawg.Item, menu *dawg.Menu) {
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

func printCategory(code string, indent int, m *dawg.Menu) {
	item := m.FindItem(code)

	switch product := item.(type) {
	case *dawg.Product:
		if len(product.Variants) == 1 {
			p, err := m.GetVariant(product.Variants[0])
			if err != nil {
				panic(err)
			}

			fmt.Fprintf(output, "%s%s  %s\n", spaces(indent*2), p.Code, p.Name)
			return
		}

		fmt.Fprintf(output, "%s%s [%s]\n", spaces(indent*2), item.ItemName(), item.ItemCode())
		n := maxStrLen(product.Variants)
		for _, variant := range product.Variants {
			v, err := m.GetVariant(variant)
			if err != nil {
				continue
			}
			fmt.Fprintf(output, "%s%s %s %s\n",
				spaces(2*(indent+1)), variant,
				spaces(n-strLen(variant)), v.Name)
		}

	case *dawg.PreConfiguredProduct:
		fmt.Fprintf(output, "%s%s   %s\n", spaces(indent*2),
			item.ItemCode(), item.ItemName())
	}
}

func maxStrLen(list []string) int {
	max := 0
	for _, s := range list {
		max = getmax(s, max)
	}

	return max
}

func getmax(s string, i int) int {
	if strLen(s) > i {
		return len(s)
	}
	return i
}

func spaces(i int) string {
	return strings.Repeat(" ", i)
}

var strLen = utf8.RuneCountInString // this is a function
