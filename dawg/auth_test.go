package dawg

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/harrybrwn/apizza/dawg/internal/auth"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestBadCreds(t *testing.T) {
	t.Skip("test takes too long")
	// swap the default client with one that has a
	// 10s timeout then defer the cleanup.
	defer swapclient(8)()
	tests.InitHelpers(t)

	err := authorize(orderClient.Client, "5uup;hrg];ht8bijer$u9tot", "hurieahgr9[0249eingurivja")
	tests.Exp(err)
	if _, ok := orderClient.Transport.(*auth.Token); ok {
		t.Error("bad authorization should not set the client transport to a token")
	}
	tok, err := gettoken("no", "and no")
	tests.Exp(err)
	if tok != nil {
		t.Errorf("expected nil %T", tok)
	}
	if _, ok := err.(*auth.Error); !ok {
		t.Errorf("expected an *auth.Error got %T:\n%v", err, err)
	}
	user, err := login(orderClient)
	tests.Exp(err)
	if user != nil {
		t.Errorf("expected nil %T", user)
	}

	username, password, ok := gettestcreds()
	if !ok {
		t.Skip()
	}
	orderClient.Client.Transport = newRoundTripper(func(*http.Request) error {
		return errors.New("this should make the client fail")
	})
	tok, err = gettoken(username, password)
	tests.Exp(err)
	if tok != nil {
		t.Errorf("expected nil %T", tok)
	}
}

func gettestcreds() (string, string, bool) {
	u, p := os.Getenv("DOMINOS_TEST_USER"), os.Getenv("DOMINOS_TEST_PASS")
	if len(u) == 0 || len(p) == 0 {
		fmt.Println("Warning: could not find test credentials")
		return u, p, false
	}
	return u, p, true
}

var testUser *UserProfile

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
	dur := time.Duration(timeout) * time.Second
	copyclient := orderClient
	orderClient = &client{
		host: orderHost,
		Client: &http.Client{
			Timeout: dur,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   dur,
				IdleConnTimeout:       dur,
				ResponseHeaderTimeout: dur,
			},
			CheckRedirect: noRedirects,
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
	// defer swapclient(8)()
	client, mux, server := testServer()
	defer server.Close()
	defer swapClientWith(client)()
	addUserHandlers(t, mux)
	tests.InitHelpers(t)

	tok, err := gettoken(username, password)
	tests.Check(err)
	if tok == nil {
		t.Fatalf("got nil %T got %v", tok, tok)
	}
	if len(tok.AccessToken) == 0 {
		t.Error("didn't get a auth token")
	}
	if tok.Type != "Bearer" {
		t.Error("these tokens are usually bearer tokens")
	}
	if len(tok.AccessToken) == 0 {
		t.Error("did not get the access token")
	}
}

func TestToken_Err(t *testing.T) {
	tokErr := newTokErr(strings.NewReader(`{"error":"","error_description":""}`))
	if tokErr != nil {
		t.Error("this error should be nil")
	}
	tokErr = newTokErr(strings.NewReader(`{"error":"test","error_description":"test"}`))
	if tokErr == nil {
		t.Error("this error should not be nil")
	}
	if tokErr.Error() != "test: test" {
		t.Error("got wrong error msg")
	}
}
