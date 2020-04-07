package dawg

import (
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestSignIn(t *testing.T) {
	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(5)()

	// user, err := getTestUser(username, password) // calls SignIn if global user is nil
	user, err := SignIn(username, password)
	if err != nil {
		t.Error(err)
	}
	if user == nil {
		t.Fatal("got nil user from SignIn")
	}
	testUser = user
}

func TestUserProfile_NearestStore(t *testing.T) {
	uname, pass, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(5)()
	tests.InitHelpers(t)

	user, err := getTestUser(uname, pass)
	tests.Check(err)
	if user == nil {
		t.Fatal("user is nil")
	}
	user.SetServiceMethod(Carryout)
	user.Addresses = []*UserAddress{}
	if user.DefaultAddress() != nil {
		t.Error("we just set this to an empty array, why is it not so")
	}
	user.AddAddress(testAddress())
	if user.DefaultAddress() == nil {
		t.Error("ok, we just added an address, why am i not getting one")
	}
	_, err = user.NearestStore(Delivery)
	tests.Check(err)
	if user.store == nil {
		t.Error("ok, now this variable should be stored")
	}
	s, err := user.NearestStore(Delivery)
	tests.Check(err)
	if s != user.store {
		t.Error("user.NearestStore should return the cached store on the second call")
	}
}

func TestUserProfile_StoresNearMe(t *testing.T) {
	uname, pass, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(5)()
	tests.InitHelpers(t)

	user, err := getTestUser(uname, pass)
	tests.Check(err)
	if user == nil {
		t.Fatal("user should not be nil")
	}
	err = user.SetServiceMethod("not correct")
	tests.Exp(err, "expected error for an invalid service method")
	if err != ErrBadService {
		t.Error("SetServiceMethod with bad val gave wrong error")
	}
	user.ServiceMethod = ""
	user.AddAddress(testAddress())
	stores, err := user.StoresNearMe()
	tests.Exp(err)
	if stores != nil {
		t.Error("should not have returned any stores")
	}

	tests.Check(user.SetServiceMethod(Delivery))
	addr := user.DefaultAddress()

	stores, err = user.StoresNearMe()
	tests.PrintErrType = true
	tests.Check(err)
	for _, s := range stores {
		if s == nil {
			t.Error("should not have nil store")
		}
		if s.userAddress == nil {
			t.Fatal("nil store.userAddress")
		}
		tests.StrEq(s.userService, user.ServiceMethod, "wrong service method")
		tests.StrEq(s.userAddress.City(), addr.City(), "wrong city")
		tests.StrEq(s.userAddress.LineOne(), addr.LineOne(), "wrong line one")
		tests.StrEq(s.userAddress.StateCode(), addr.StateCode(), "wrong state code")
		tests.StrEq(s.userAddress.Zip(), addr.Zip(), "wrong zip code")
	}
}

func TestUserProfile_NewOrder(t *testing.T) {
	uname, pass, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(5)()
	tests.InitHelpers(t)

	user, err := getTestUser(uname, pass)
	tests.Check(err)
	if user == nil {
		t.Fatal("user should not be nil")
	}
	user.SetServiceMethod(Carryout)
	order, err := user.NewOrder()
	tests.Check(err)

	tests.StrEq(order.ServiceMethod, Carryout, "wrong service method")
	tests.StrEq(order.ServiceMethod, user.ServiceMethod, "service method should carry over from the user")
	tests.StrEq(order.Phone, user.Phone, "phone should carry over from user")
	tests.StrEq(order.FirstName, user.FirstName, "first name should carry over from user")
	tests.StrEq(order.LastName, user.LastName, "last name should carry over from user")
	tests.StrEq(order.CustomerID, user.CustomerID, "customer id should carry over")
	tests.StrEq(order.Email, user.Email, "order email should carry over from user")
	tests.StrEq(order.StoreID, user.store.ID, "store id should carry over")
	if order.Address == nil {
		t.Error("order should get and address from the user")
	}
}
