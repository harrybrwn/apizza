package dawg

import (
	"net/http"
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestProduct(t *testing.T) {
	cli, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(cli)()
	mux.HandleFunc("/power/store-locator", storeLocatorHandlerFunc(t))
	mux.HandleFunc("/power/store/4344/profile", storeProfileHandlerFunc(t))
	mux.HandleFunc("/power/store/4328/menu", func(w http.ResponseWriter, r *http.Request) {
		fileHandleFunc(t, "./testdata/menu.json")(w, r)
	})

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

func TestProductToppings(t *testing.T) {
	cli, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(cli)()
	mux.HandleFunc("/power/store-locator", storeLocatorHandlerFunc(t))
	mux.HandleFunc("/power/store/4344/profile", storeProfileHandlerFunc(t))
	mux.HandleFunc("/power/store/4328/menu", func(w http.ResponseWriter, r *http.Request) {
		fileHandleFunc(t, "./testdata/menu.json")(w, r)
	})

	tests.InitHelpers(t)
	m := testingMenu()
	p, err := m.GetProduct("S_PIZZA") // pizza
	tests.Check(err)

	err = p.AddTopping("notatopping", ToppingFull, "1.9")
	tests.Exp(err) // expect an error
	p.opts = nil
	if len(p.Options()) == 0 {
		t.Error("should not be len zero even after set to nil (see Options impl for Product)")
	}
	tests.Check(p.AddTopping("K", ToppingLeft, "2.0"))
	if _, ok := p.Options()["K"]; !ok {
		t.Error("bacon should have been added")
	}
	topps := ReadableToppings(p, m)
	if len(topps) == 0 {
		t.Error("should still have some toppings")
	}
	if _, ok := topps["Bacon (K)"]; !ok {
		t.Error("bacon should have been included")
	}
	if _, ok := topps["Cheese (C)"]; !ok {
		t.Error("should be able to refere to 'C' as Cheese")
	}

	v, err := m.GetVariant("14SCREEN")
	tests.Check(err)
	if v.FindProduct(m) == nil {
		t.Error("should not be nil, pizza has a category")
	}
	if v.GetProduct() == nil {
		t.Error("this should have a product")
	}
	if v.FindProduct(nil) == nil {
		t.Error("should not be nil")
	}
	old := v.ProductCode
	v.ProductCode = "nothere"
	v.product = nil
	if v.FindProduct(m) != nil {
		t.Error("expected nil response")
	}
	v.ProductCode = old

	for _, v := range m.Variants {
		if v.FindProduct(m) == nil {
			t.Error("should not be nil:", v.Code)
		}
		if v.GetProduct() == nil {
			t.Error("this should not be nil")
		}
	}
	if longerlength(make([]string, 1), make([]string, 5)) != 5 {
		t.Error("what?????")
	}
	if longerlength(make([]string, 2), make([]string, 6)) != 6 {
		t.Error("that is not the length of the longest one")
	}
}

func TestViewOptions(t *testing.T) {
	cli, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(cli)()
	mux.HandleFunc("/power/store-locator", storeLocatorHandlerFunc(t))
	mux.HandleFunc("/power/store/4344/profile", storeProfileHandlerFunc(t))
	mux.HandleFunc("/power/store/4328/menu", func(w http.ResponseWriter, r *http.Request) {
		fileHandleFunc(t, "./testdata/menu.json")(w, r)
	})
	m := testingMenu()

	itm, err := m.GetVariant("P10IRECK")
	if err != nil {
		t.Error(err)
	}
	opts := m.ViewOptions(itm)
	exp := map[string]string{"Cheese (C)": "full 1", "BBQ Sauce (Bq)": "full 1", "Onions (O)": "full 1", "Premium Chicken (Du)": "full 1", "Cheddar Cheese (E)": "full 1", "Shredded Provolone Cheese (Cp)": "full 1"}

	for k := range opts {
		if opts[k] != exp[k] {
			t.Error("bad topping format")
		}
	}
}
