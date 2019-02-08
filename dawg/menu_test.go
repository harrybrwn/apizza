package dawg

import (
	"fmt"
	"testing"
)

func TestMakeProduct(t *testing.T) {
	id := "4336"
	service := "Delivery"
	store, err := NewStore(id, service, nil)
	if err != nil {
		t.Error(err)
	}
	p, err := store.GetProduct("12SCREEN")
	if err != nil {
		t.Error(err)
		fmt.Println(p)
	}
	t.Run("badToppingCoverage", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		p.AddTopping("X", "invalid cover", 2.0)
	})
	t.Run("badToppingAmount", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		p.AddTopping("C", ToppingLeft, 1.6)
	})

	p, err = store.GetProduct("MARBRWNE")
	if err != nil {
		t.Error(err)
	}
	p.AddTopping("C", ToppingLeft, 1.0)
	p.AddTopping("X", ToppingRight, 1.5)

	m, err := store.Menu()
	if err != nil {
		t.Error(err)
	}
	p, err = makeProduct(m.Variants["12THIN"].(map[string]interface{}))
	if err != nil {
		t.Error(err)
		fmt.Println(p)
	}
	if p.Price() < 0 {
		t.Error("error in finding product.Price(): returned -1.0")
	}
	if p.Size() < 0 {
		t.Error("error in finding product.Size(): returned -1")
	}
	if !p.Prepared() {
		t.Error("should have been false. got true")
	}

	p = &Product{}
	if p.Price() != -1 {
		t.Error("expected -1")
	}
	if p.Size() != -1 {
		t.Error("expected -1")
	}
	if p.Prepared() {
		t.Error("expected false")
	}
}

func TestNewMenu(t *testing.T) {
	cachedMenu = nil
	if cachedMenu != nil {
		t.Error("wtf, if this broke than you should reinstall the go compiler")
	}
	m, err := newMenu("4336")
	if cachedMenu == nil {
		t.Error("menu caching failed")
	}
	if err != nil {
		t.Error(err)
	}
	if m == nil {
		t.Error("newMenu returned a nil menu")
	}
	if m.ID != "4336" {
		t.Error("newMenu returned a menu for the wrong store")
	}
}
