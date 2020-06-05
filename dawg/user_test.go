package dawg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/harrybrwn/apizza/dawg/internal/auth"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestSignIn(t *testing.T) {
	client, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(client)()
	addUserHandlers(t, mux)
	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	tests.InitHelpers(t)

	user, err := getTestUser(username, password) // calls SignIn if global user is nil
	tests.Check(err)
	tests.NotNil(user)
	user, err = SignIn("blah", "blahblah")
	tests.Exp(err)
	if user != nil {
		t.Errorf("expected nil %T", user)
	}
}

func TestUser_WithProxy(t *testing.T) {
	client, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(client)()
	addUserHandlers(t, mux)
	uname, pass, _ := gettestcreds()
	user, err := SignIn(uname, pass)
	if err != nil {
		t.Error(err)
	}
	if user == nil {
		t.Fatal("nil user")
	}
	if user.ID != "123" {
		t.Error("wrong id")
	}
	if user.cli.Client.Transport.(*auth.Token).ExpiresIn != 42069 {
		t.Error("wrong expiration number")
	}
}

func TestUser(t *testing.T) {
	t.Skip("too wild")
	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(20)()
	tests.InitHelpers(t)
	user, err := getTestUser(username, password)
	tests.Check(err)
	if user == nil {
		t.Fatalf("got nil %T", user)
	}
	if user.cli == nil {
		t.Fatalf("got nil %T", user.cli)
	}
	if user.cli.Client == nil {
		t.Fatalf("got nil %T", user.cli.Client)
	}
	user.SetServiceMethod(Delivery)
	user.AddAddress(testAddress())
	user.Addresses[0].StreetNumber = ""
	user.Addresses[0].StreetName = ""
	user.AddAddress(user.Addresses[0])
	a1 := user.Addresses[0]
	a2 := user.Addresses[1]
	if a1.StreetName != a2.StreetName {
		t.Error("did not copy address name correctly")
	}
	if a1.StreetNumber != a2.StreetNumber {
		t.Error("did not copy address number correctly")
	}
	a1.Street = ""
	if user.Addresses[0].LineOne() != a2.LineOne() {
		t.Error("line one for UserAddress is broken")
	}

	if testing.Short() {
		return
	}
	store, err := user.NearestStore(Delivery)
	tests.Check(err)
	tests.NotNil(store)
	tests.NotNil(store.cli)
	tests.StrEq(store.cli.host, "order.dominos.com", "store client has the wrong host")

	if _, ok = store.cli.Client.Transport.(*auth.Token); !ok {
		t.Fatal("store's client should have gotten a token as its transport")
	}

	// Checking that the authorization header is carried accross a request
	req := &http.Request{
		Method: "GET", Host: orderHost, Proto: "HTTP/1.1",
		URL: &url.URL{
			Scheme: "https", Host: orderHost,
			Path:     fmt.Sprintf("/power/store/%s/menu", store.ID),
			RawQuery: (&Params{"lang": DefaultLang, "structured": "true"}).Encode()},
	}
	res, err := store.cli.Do(req)
	tests.Check(err)
	defer func() { tests.Check(res.Body.Close()) }()
	authhead := res.Request.Header.Get("Authorization")
	if len(authhead) <= len("Bearer ") {
		t.Error("store client didn't get the token")
	}
	b, err := ioutil.ReadAll(res.Body)
	tests.Check(err)
	if len(b) == 0 {
		t.Error("zero length response")
	}
}

func TestUserProfile_NearestStore(t *testing.T) {
	t.Skip("this test is really broken")
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
	store, err := user.NearestStore(Delivery)
	tests.Check(err)
	tests.NotNil(store)
	tests.NotNil(user.store)
	storeAgain, err := user.NearestStore(Delivery)
	tests.Check(err)
	if store != storeAgain {
		t.Error("should have returned the same store")
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
	cli, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(cli)()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", r)
	})
	addUserHandlers(t, mux)
	mux.HandleFunc("/power/store-locator", func(w http.ResponseWriter, r *http.Request) {
		storeLocatorHandlerFunc(t)(w, r)
	})
	// mux.HandleFunc("/power/store/", storeProfileHandlerFunc(t))
	mux.HandleFunc("/power/store/", func(w http.ResponseWriter, r *http.Request) {
		storeProfileHandlerFunc(t)(w, r)
	})

	tests.InitHelpers(t)
	user, err := SignIn(uname, pass)
	tests.Check(err)
	if user == nil {
		t.Fatal("user should not be nil")
	}
	fmt.Printf("%p\n", user.cli.Client)
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
	cli, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(cli)()
	addUserHandlers(t, mux)
	mux.HandleFunc("/power/store-locator", storeLocatorHandlerFunc(t))
	mux.HandleFunc("/power/store/4344/profile", storeProfileHandlerFunc(t))
	tests.InitHelpers(t)

	user, err := getTestUser(uname, pass)
	tests.Check(err)
	if user == nil {
		t.Fatal("user should not be nil")
	}
	user.store = &Store{userAddress: testAddress()}
	user.AddAddress(testAddress())
	user.SetServiceMethod(Carryout)
	order, err := user.NewOrder()
	tests.Check(err)
	if order == nil {
		t.Fatal("user.NewOrder() returend a nil order")
	}

	tests.StrEq(order.ServiceMethod, Carryout, "wrong service method")
	tests.StrEq(order.ServiceMethod, user.ServiceMethod, "service method should carry over from the user")
	tests.StrEq(order.Phone, user.Phone, "phone should carry over from user")
	tests.StrEq(order.FirstName, user.FirstName, "first name should carry over from user")
	tests.StrEq(order.LastName, user.LastName, "last name should carry over from user")
	tests.StrEq(order.CustomerID, user.ID, "customer id should carry over")
	tests.StrEq(order.Email, user.Email, "order email should carry over from user")
	tests.StrEq(order.StoreID, user.store.ID, "store id should carry over")
	if order.Address == nil {
		t.Error("order should get and address from the user")
	}
}

func addUserHandlers(t *testing.T, mux *http.ServeMux) {
	t.Helper()
	username, password, ok := gettestcreds()
	if !ok {
		t.Error("could not get test credentials")
	}
	mux.HandleFunc("/auth-proxy-service/login", func(w http.ResponseWriter, r *http.Request) {
		var b bytes.Buffer
		b.ReadFrom(r.Body)
		creds, err := url.ParseQuery(b.String())
		if err != nil || creds["password"][0] == "" || creds["username"][0] == "" {
			w.WriteHeader(http.StatusUnauthorized)
			t.Error("Bad request", err)
			return
		}
		if creds["password"][0] != password || creds["username"][0] != username {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if _, err = w.Write([]byte(`{"access_token":"testtoken","token_type":"Bearer","expires_in":42069}`)); err != nil {
			t.Error(err)
		}
	})
	mux.HandleFunc("/power/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer testtoken" {
			t.Error("bad authorization header")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		u := UserProfile{FirstName: "tester", LastName: "tester", ID: "123", Email: "testing@test.com"}
		json.NewEncoder(w).Encode(&u)
	})
}

var testdataMutex sync.Mutex

func fileHandleFunc(t *testing.T, filename string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// testdataMutex.Lock()
		// defer testdataMutex.Unlock()
		file, err := os.Open(filename)
		if err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if _, err = io.Copy(w, file); err != nil {
			t.Error(err)
		}
	}
}
