package dawg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestBadCreds(t *testing.T) {
	// swap the default client with one that has a
	// 10s timeout then defer the cleanup.
	defer swapclient(2)()
	tests.InitHelpers(t)

	tok, err := gettoken("no", "and no")
	tests.Exp(err)
	if tok != nil {
		t.Error("expected nil token")
	}

	tok, err = gettoken("", "")
	tests.Exp(err)
	if tok != nil {
		t.Error("expected nil token")
	}

	tok, err = gettoken("5uup;hrg];ht8bijer$u9tot", "hurieahgr9[0249eingurivja")
	tests.Exp(err)
	if tok != nil {
		t.Error("expected nil token")
	}
}

func gettestcreds() (string, string, bool) {
	u, p := os.Getenv("DOMINOS_TEST_USER"), os.Getenv("DOMINOS_TEST_PASS")
	if len(u) == 0 || len(p) == 0 {
		return u, p, false
	}
	return u, p, true
}

var (
	testAuth *auth
	testUser *UserProfile
)

func getTestAuth(uname, pass string) (*auth, error) {
	var err error
	if testAuth == nil {
		testAuth, err = newauth(uname, pass)
	}
	return testAuth, err
}

func getTestUser(uname, pass string) (*UserProfile, error) {
	var err error
	if testUser == nil {
		testUser, err = SignIn(uname, pass)
	}
	return testUser, err
}

// swapclient is meant to swap out the default client for a test
// function and return a function that resets the default client.
// This is only here because the dominos auth endpoint times out
// kind of frequently so im just making the so that it doesn't
// halt my tests for like 60+ seconds per test.
//
// usage:
// 	defer swapclient(15)()
// this will call swapclient and defer the cleanup function
// that it returns so that the default client is reset.
//
// if someone is actually reading this, im sorry, i know this
// is not very go-like, i know its hacky... sorry
func swapclient(timeout int) func() {
	copyclient := orderClient
	orderClient = &client{
		host: orderHost,
		Client: &http.Client{
			Timeout:       time.Duration(timeout) * time.Second,
			CheckRedirect: noRedirects,
			Transport: newRoundTripper(func(req *http.Request) error {
				agent := fmt.Sprintf("TestClient: %d%d", rand.Int(), time.Now().Nanosecond())
				// fmt.Printf("setting user agent to '%s'\n", agent)
				req.Header.Set("User-Agent", agent)
				return nil
			}),
		},
	}
	return func() { orderClient = copyclient }
}

func TestToken(t *testing.T) {
	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	// swapclient is called first and the cleanup
	// function it returns is deferred.
	defer swapclient(5)()
	tests.InitHelpers(t)

	tok, err := gettoken(username, password)
	tests.Check(err)
	if tok == nil {
		t.Fatal("nil token")
	}
	if len(tok.AccessToken) == 0 {
		t.Error("didn't get a auth token")
	}
	if !strings.HasPrefix(tok.authorization(), "Bearer ") {
		t.Error("bad auth format")
	}
	if tok.transport == nil {
		t.Error("token should have a transport")
	}
	if tok.Type != "Bearer" {
		t.Error("these tokens are usually bearer tokens")
	}
}

func TestAuth(t *testing.T) {
	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(5)()
	tests.InitHelpers(t)
	user, err := SignIn(username, password)
	tests.Check(err)
	if user == nil {
		t.Fatal("got nil user-profile")
	}
	auth := user.auth
	if auth == nil {
		t.Fatal("got nil auth")
	}
	if auth.token == nil {
		t.Fatal("needs token")
	}
	if len(auth.username) == 0 {
		t.Error("didn't save username")
	}
	if len(auth.password) == 0 {
		t.Error("didn't save password")
	}
	if auth.cli == nil {
		t.Fatal("needs to have client")
	}
	if len(auth.token.AccessToken) == 0 {
		t.Error("no access token")
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
	store, err := user.NearestStore("Delivery")
	tests.Check(err)
	if store == nil {
		t.Fatal("store is nil")
	}
	store1, err := user.NearestStore(Delivery)
	tests.Check(err)
	if store != store1 {
		t.Error("should be the same store")
	}
	if store.cli == nil {
		t.Fatal("store did not get a client")
	}
	if store.cli.host != "order.dominos.com" {
		t.Error("store client has the wrong host")
	}

	req := &http.Request{
		Method: "GET", Host: store.cli.host, Proto: "HTTP/1.1",
		URL: &url.URL{
			Scheme: "https", Host: store.cli.host,
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

	menu, err := store.Menu()
	tests.Check(err)
	if menu == nil {
		t.Error("got nil menu")
	}
	o := store.NewOrder()
	if o == nil {
		t.Error("nil order")
	}
	_, err = o.Price()
	tests.Check(err)
}

func TestAuth_Err(t *testing.T) {
	defer swapclient(2)()
	tests.InitHelpers(t)
	a, err := newauth("not a", "valid password")
	tests.Exp(err)
	if a != nil {
		t.Error("expected a nil auth")
	}
	a = &auth{
		username: "not a",
		password: "valid password",
		token:    &token{}, // assume we already have a token
		cli: &client{
			host: "order.dominos.com",
			Client: &http.Client{
				Timeout:       15 * time.Second,
				CheckRedirect: noRedirects,
			},
		},
	}

	user, err := a.login()
	tests.Exp(err)
	if user != nil {
		t.Errorf("expected a nil user: %+v", user)
	}
	a.cli.host = "invalid_host.com"
	user, err = a.login()
	tests.Exp(err)
	if user != nil {
		t.Error("user should still be nil")
	}
}

func TestAuthClient(t *testing.T) {
	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	defer swapclient(5)()
	tests.InitHelpers(t)

	auth, err := getTestAuth(username, password)
	tests.Check(err)
	if auth == nil {
		t.Fatal("got nil auth")
	}

	if auth.cli == nil {
		t.Error("client should not be nil")
	}
	tests.Exp(auth.cli.CheckRedirect(nil, nil), "order Client should not allow redirects")
	tests.Exp(auth.cli.CheckRedirect(&http.Request{}, []*http.Request{}))
	cleanup := swapclient(2)
	tok, err := gettoken("bad", "creds")
	tests.Exp(err, "should return error")
	cleanup()

	req := newAuthRequest(oauthURL, url.Values{})
	resp, err := http.DefaultClient.Do(req)
	tests.Check(err)
	tok = &token{}
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)

	err = unmarshalToken(ioutil.NopCloser(buf), tok)
	tests.Exp(err)
	if e, ok := err.(*tokenError); !ok {
		t.Error("expected a *tokenError as the error")
		fmt.Println(buf.String())
	} else if e.Error() != fmt.Sprintf("%s: %s", e.Err, e.ErrorDesc) {
		t.Error("wrong error message")
	}
	if IsOk(err) {
		t.Error("this shouldn't happen")
	}
}
