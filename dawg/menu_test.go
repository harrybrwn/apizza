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
	p, err = store.GetProduct("MARBRWNE")
	if err != nil {
		t.Error(err)
	}
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
	if p.Prepared() {
		t.Error("p.Prepared() should have been false. got true")
	}
	p, err = makeProduct(map[string]interface{}{})
	t.Log(p)
	t.Log(err)
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
		t.Error("newMenu returned a menu for the wrong store	")
	}
}
