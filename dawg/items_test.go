package dawg

import (
	"testing"
)

func TestProduct(t *testing.T) {
	menu, err := testingStore().Menu()
	if err != nil {
		t.Error(err)
	}

	product, err := menu.GetProduct("S_BONELESS")
	if err != nil {
		t.Error(err)
	}

	expected := []*Variant{}
	for _, code := range []string{"W08PBNLW", "W14PBNLW", "W40PBNLW"} {
		if product.ItemCode() != "S_BONELESS" {
			t.Fatal("tests wont work if the product doesn't correspond with the variants.")
		}
		v, err := menu.GetVariant(code)
		if err != nil {
			t.Error(err)
		}
		expected = append(expected, v)
	}

	pVariants := product.GetVariants(menu)

	if len(pVariants) != len(expected) {
		t.Fatal("got different length of variants from Product.GetVariants()")
	} else {
		for i := range expected {
			if expected[i] != pVariants[i] {
				t.Error("got different list of variants from Product.GetVariants()")
			}
		}
	}
}
