package dawg

import (
	"testing"
)

func TestUser(t *testing.T) {
	uname, pass, ok := gettestcreds()
	if !ok {
		t.Skip()
	}

	user, err := SignIn(uname, pass)
	if err != nil {
		t.Error(err)
	}
	if err = user.SetServiceMethod("not correct"); err == nil {
		t.Error("expected error for an invalid service method")
	}
	if err != errBadService {
		t.Error("SetServiceMethod with bad val gave wrong error")
	}
	user.AddAddress(testAddress())
	stores, err := user.StoresNearMe()
	if err == nil {
		t.Error("expedted error")
	}
	if err != errNoServiceMethod {
		t.Error("wrong error")
	}
	if stores != nil {
		t.Error("should not have retured any stores")
	}

	if err = user.SetServiceMethod(Delivery); err != nil {
		t.Error(err)
	}
	stores, err = user.StoresNearMe()
	if err != nil {
		t.Error(err)
	}
	for _, s := range stores {
		if s == nil {
			t.Error("should not have nil store")
		}
		// b, _ := json.MarshalIndent(s, "", "   ")
		// fmt.Printf("%s\n", b)
	}
}
