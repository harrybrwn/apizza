package dawg

import (
	"bytes"
	"encoding/gob"
	"os"
	"testing"
)

// Move this to an items_test.go file
func TestItems(t *testing.T) {
	store := testingStore()
	menu, err := store.Menu()
	if err != nil {
		t.Error(err)
	}

	testcases := []struct {
		product, variant, top, cover, coverErr string
		isSubset, wanterr                      bool
	}{
		// {product: "F_PARMT", variant: "B8PCPT", top: "K", isSubset: true, wanterr: false},
		{
			product:  "S_MX",
			variant:  "14TMEATZA",
			top:      "B",
			isSubset: true,
			wanterr:  false,
			cover:    "2",
			coverErr: "1.7",
		},
		{
			product:  "S_PISPF",
			variant:  "P10IRESPF",
			top:      "B",
			isSubset: true,
			wanterr:  false,
			cover:    "2",
			coverErr: "-1.7",
		},
		{
			product:  "S_BONELESS",
			variant:  "W08PBNLW",
			top:      "",
			isSubset: true,
			wanterr:  false,
			cover:    "2",
			coverErr: "-1.7",
		},
	}

	for _, tc := range testcases {
		p, err := menu.GetProduct(tc.product)
		if tc.wanterr && err == nil {
			t.Error("expected error")
		} else if err != nil {
			t.Error(err)
		}
		v, err := menu.GetVariant(tc.variant)
		if tc.wanterr && err == nil {
			t.Error("expected error")
		} else if err != nil {
			t.Error(err)
		}

		if tc.isSubset {
			for _, variant := range p.Variants {
				if variant == tc.variant {
					goto foundVariant
				}
			}
			t.Errorf("%s should be a variant of %s", tc.variant, tc.product)
		foundVariant:
		}
		if err = p.AddTopping(tc.top, ToppingLeft, tc.cover); err != nil {
			t.Error(err)
		}
		if err = v.AddTopping(tc.top, ToppingFull, tc.cover); err != nil {
			t.Error(err)
		}
		if err = v.AddTopping(tc.top, "1/1", tc.coverErr); err == nil {
			t.Error("expected error")
		}
		if len(v.opts) < 1 {
			t.Error("should have options in the struct")
		}
	}
}

func TestOPFromItem(t *testing.T) {
	m := testingMenu()
	v, err := m.GetVariant("W08PBNLW")
	if err != nil {
		t.Error(err)
	}
	p, err := m.GetProduct("S_BONELESS")
	if err != nil {
		t.Error(err)
	}

	opv := OrderProductFromItem(v)
	opp := OrderProductFromItem(p)

	opvOpts := opv.Options()
	oppOpts := opp.Options()

	for k := range opvOpts {
		if _, ok := oppOpts[k]; !ok {
			t.Errorf("order product should have %s", k)
		}
	}
	for k := range v.Options() {
		if _, ok := opvOpts[k]; !ok {
			t.Error("options should be the same")
		}
	}
	for k := range p.Options() {
		if _, ok := oppOpts[k]; !ok {
			t.Error("options should be the same")
		}
	}

	if opv.Category() != opp.Category() {
		t.Error("the variant and it's parent should have the same product type")
	}
}

func TestFindItem(t *testing.T) {
	m := testingMenu()

	tt := []string{"W08PBNLW", "S_BONELESS", "F_PARMT", "P_14SCREEN"}
	for _, tc := range tt {
		itm := m.FindItem(tc)
		if itm == nil {
			t.Error("item is nil")
		}
	}

	itm := m.FindItem("badCode")
	if itm != nil {
		t.Error("item should be nil")
	}

	_, err := m.GetProduct("nothere")
	if err == nil {
		t.Error("expected error")
	}
	_, err = m.GetVariant("nothere")
	if err == nil {
		t.Error("expected error")
	}
}

func TestTranslateOpt(t *testing.T) {
	opts := map[string]interface{}{
		"what": "no",
	}
	if translateOpt(opts) != "what no" {
		t.Error("wrong output")
	}
	opt := map[string]string{
		ToppingRight: "9.0",
	}
	if translateOpt(opt) != "right 9.0" {
		t.Error("wrong option translation")
	}
	opt = map[string]string{
		ToppingLeft: "5.5",
	}
	if translateOpt(opt) != "left 5.5" {
		t.Error("wrong")
	}
}

var testmenu = testingMenu()

func TestPrintMenu(t *testing.T) {
	m := testmenu
	buf := new(bytes.Buffer)

	m.Print(buf)
	if buf.Len() == 0 {
		t.Error("should not have a zero length printout")
	}
}

func TestMenuStorage(t *testing.T) {
	check := func(e error) {
		if e != nil {
			t.Error(e)
		}
	}
	m := testmenu
	fname := "/tmp/apizza-binary-menu"
	buf := &bytes.Buffer{}
	gob.Register([]interface{}{})
	err := gob.NewEncoder(buf).Encode(m)
	if err != nil {
		t.Fatal("gob encoding error:", err)
	}

	f, err := os.Create(fname)
	check(err)
	_, err = f.Write(buf.Bytes())
	check(err)
	check(f.Close())
	file, err := os.Open(fname)
	check(err)
	menu := Menu{}
	err = gob.NewDecoder(file).Decode(&menu)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	if menu.ID != m.ID {
		t.Error("wrong id")
	}
	if menu.Preconfigured == nil {
		t.Fatal("should have decoded the Preconfigured products")
	}

	for k := range m.Preconfigured {
		mp := m.Preconfigured[k]
		menup := menu.Preconfigured[k]
		if mp.Code != menup.Code {
			t.Errorf("Stored wrong Code - got: %s, want: %s\n", menup.Code, mp.Code)
		}
		if mp.Opts != menup.Opts {
			t.Errorf("Stored wrong opt - got: %s, want: %s\n", menup.Opts, mp.Opts)
		}
		if mp.Category() != menup.Category() {
			t.Error("Stored wrong category")
		}
	}
	for k := range m.Products {
		mp := m.Products[k]
		menup := menu.Products[k]
		if mp.Code != menup.Code {
			t.Errorf("Stored wrong product code - got: %s, want: %s\n", menup.Code, mp.Code)
		}
		if mp.DefaultSides != menup.DefaultSides {
			t.Errorf("Stored wrong product DefaultSides - got: %s, want: %s\n", menup.DefaultSides, mp.DefaultSides)
		}
	}
	if err = os.Remove(fname); err != nil {
		t.Error(err)
	}
}
