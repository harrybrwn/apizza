package dawg

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
)

// Move this to an items_test.go file
func TestItems(t *testing.T) {
	tests.InitHelpers(t)
	store := testingStore()
	menu, err := store.Menu()
	tests.Check(err)

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
			t.Errorf("expected error from menu.GetProduct(%s)", tc.product)
		} else {
			tests.Check(err)
		}
		v, err := menu.GetVariant(tc.variant)
		if tc.wanterr && err == nil {
			t.Errorf("expected error from menu.GetVariant(%s)", tc.variant)
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
		tests.Check(p.AddTopping(tc.top, ToppingLeft, tc.cover))
		tests.Check(v.AddTopping(tc.top, ToppingFull, tc.cover))
		if err = v.AddTopping(tc.top, "1/1", tc.coverErr); err == nil {
			t.Error("expected error")
		}
		if len(v.opts) < 1 {
			t.Error("should have options in the struct")
		}
	}
}

func TestOPFromItem(t *testing.T) {
	tests.InitHelpers(t)
	m := testingMenu()
	v, err := m.GetVariant("W08PBNLW")
	tests.Check(err)
	p, err := m.GetProduct("S_BONELESS")
	tests.Check(err)

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

	tests.StrEq(opv.Category(), opp.Category(), "the variant and it's parent should have the same product type")
}

func TestFindItem(t *testing.T) {
	tests.InitHelpers(t)
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
	tests.Exp(err)
	_, err = m.GetVariant("nothere")
	tests.Exp(err)
}

func TestTranslateOpt(t *testing.T) {
	tests.InitHelpers(t)
	opts := map[string]interface{}{
		"what": "no",
	}
	tests.StrEq(translateOpt(opts), "what no", "wrong option translation")
	opt := map[string]string{
		ToppingRight: "9.0",
	}
	tests.StrEq(translateOpt(opt), "right 9.0", "wrong option translation")
	opt = map[string]string{
		ToppingLeft: "5.5",
	}
	tests.StrEq(translateOpt(opt), "left 5.5", "wrong option translation")
}

func TestPrintMenu(t *testing.T) {
	m := testingMenu()
	buf := new(bytes.Buffer)

	m.Print(buf)
	if buf.Len() == 0 {
		t.Error("should not have a zero length printout")
	}
}

func TestMenuStorage(t *testing.T) {
	tests.InitHelpers(t)
	testdir := tests.MkTempDir(t.Name())

	m := testingMenu()
	fname := filepath.Join(testdir, "apizza-binary-menu")
	buf := &bytes.Buffer{}
	gob.Register([]interface{}{})
	err := gob.NewEncoder(buf).Encode(m)
	tests.Fatal(err)

	f, err := os.Create(fname)
	tests.Check(err)
	_, err = f.Write(buf.Bytes())
	tests.Check(err)
	tests.Check(f.Close())
	file, err := os.Open(fname)
	tests.Check(err)
	defer func() {
		tests.Check(file.Close())
	}()
	menu := Menu{}
	err = gob.NewDecoder(file).Decode(&menu)
	tests.Fatal(err)

	tests.StrEq(menu.ID, m.ID, "wrong menu id")
	if menu.Preconfigured == nil {
		t.Fatal("should have decoded the Preconfigured products")
	}

	for k := range m.Preconfigured {
		mp := m.Preconfigured[k]
		menup := menu.Preconfigured[k]
		tests.StrEq(mp.Code, menup.Code, "Stored wrong Code - got: %s, want: %s", menup.Code, mp.Code)
		tests.StrEq(mp.Opts, menup.Opts, "Stored wrong opt - got: %s, want: %s", menup.Opts, mp.Opts)
		tests.StrEq(mp.Category(), menup.Category(), "Stored wrong category")
	}
	for k := range m.Products {
		mp := m.Products[k]
		menup := menu.Products[k]
		tests.StrEq(mp.Code, menup.Code, "Stored wrong product code - got: %s, want: %s", menup.Code, mp.Code)
		tests.StrEq(mp.DefaultSides, menup.DefaultSides, "Stored wrong product DefaultSides - got: %s, want: %s", menup.DefaultSides, mp.DefaultSides)
	}
	tests.Check(os.RemoveAll(testdir))
}
