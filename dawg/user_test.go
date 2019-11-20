package dawg

import (
	"fmt"
	"testing"
	"time"
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
	addr := user.DefaultAddress()

	storeFuncs := []func() ([]*Store, error){
		user.StoresNearMe, user.StoresNearMeAsync}

	for _, fn := range storeFuncs {

		tm := time.Now()
		stores, err = fn()
		fmt.Println(time.Now().Sub(tm))
		if err != nil {
			t.Error(err)
		}

		for _, s := range stores {
			if s == nil {
				t.Error("should not have nil store")
			}
			if s.userAddress == nil {
				t.Fatal("nil store.userAddress")
			}
			if s.userService != user.ServiceMethod {
				t.Error("wrong service method")
			}
			if s.userAddress.City() != addr.City() {
				t.Error("wrong city")
			}
			if s.userAddress.LineOne() != addr.LineOne() {
				t.Error("wrong line one")
			}
			if s.userAddress.StateCode() != addr.StateCode() {
				t.Error("wrong state code")
			}
			if s.userAddress.Zip() != addr.Zip() {
				t.Error("wrong zip code")
			}
		}
	}
}
