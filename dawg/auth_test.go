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
		fmt.Println(err)
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
		t.Error("wow i accidentlly cracked someone's password:", tok)
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
		t.Fatal("nil token")
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
		t.Fatal("got nil auth")
	}
	if auth.token == nil {
		t.Fatal("needs token")
	}
	if len(auth.username) == 0 {
		t.Error("didnt save username")
	}
	if len(auth.password) == 0 {
		t.Error("didnt save password")
	}
	if auth.cli == nil {
		t.Fatal("needs to have client")
	}

	user, err := auth.login()
	if err != nil {
		t.Error(err)
	}
	if user == nil {
		t.Fatal("got nil user-profile")
	}
	user.AddAddress(testAddress())
	store, err := user.NearestStore("Delivery")
	if err != nil {
		t.Error(err)
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

func TestSignIn(t *testing.T) {
	username := os.Getenv("DOMINOS_TEST_NAME")
	password := os.Getenv("DOMINOS_TEST_PASS")
	if username == "" || password == "" {
		t.Skip()
	}

	user, err := SignIn(username, password)
	if err != nil {
		t.Error(err)
	}
	if user == nil {
		t.Fatal("got nil user from SignIn")
	}
}
