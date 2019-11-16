package dawg

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestBadCreds(t *testing.T) {
	tok, err := gettoken("no", "and no")
	if err == nil {
		t.Error("expected an error")
	}
	if tok != nil {
		t.Error("expected nil token")
	}

	tok, err = gettoken("", "")
	if err == nil {
		t.Error("expected an error")
	}
	if tok != nil {
		t.Error("expected nil token")
	}

	tok, err = gettoken("5uup;hrg];ht8bijer$u9tot", "hurieahgr9[0249eingurivja")
	if err == nil {
		t.Error("wow i accidentlly cracked someone's password")
	}
}

func TestToken(t *testing.T) {
	username := os.Getenv("DOMINOS_TEST_NAME")
	password := os.Getenv("DOMINOS_TEST_PASS")
	if username == "" || password == "" {
		t.Skip()
	}

	tok, err := gettoken(username, password)
	if err != nil {
		t.Error(err)
	}
	if tok == nil {
		t.Error("nil token")
	}
	if len(tok.AccessToken) == 0 {
		t.Error("didnt get a auth token")
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
	username := os.Getenv("DOMINOS_TEST_NAME")
	password := os.Getenv("DOMINOS_TEST_PASS")
	if username == "" || password == "" {
		t.Skip()
	}

	auth, err := newauth(username, password)
	if err != nil {
		t.Error(err)
	}
	if auth == nil {
		t.Error("got nil auth")
	}
	if auth.token == nil {
		t.Error("needs token")
	}
	if len(auth.username) == 0 {
		t.Error("didnt save username")
	}
	if len(auth.password) == 0 {
		t.Error("didnt save password")
	}
	if auth.cli == nil {
		t.Error("needs to have client")
	}

	user, err := auth.login()
	if err != nil {
		t.Error(err)
	}
	if user == nil {
		t.Error("got nil user-profile")
	}
	if len(user.Addresses) == 0 {
		user.Addresses = make([]*UserAddress, 2)
	}
	user.Addresses[0] = UserAddressFromAddress(testAddress())
	store, err := user.NearestStore("Delivery")
	if err != nil {
		t.Error(err)
	}
	if store.cli == nil {
		t.Error("store did not get a client")
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
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()
	authhead := res.Request.Header.Get("Authorization")
	if authhead != auth.token.authorization() {
		t.Error("store client didnt get the token")
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	if len(b) == 0 {
		t.Error("zero length response")
	}
}
