package dawg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/harrybrwn/apizza/dawg/internal/auth"
)

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

const (
	oauthEndpoint = "https://api.dominos.com/as/token.oauth2"
	loginEndpoint = "https://order.dominos.com/power/login"
)

var (
	// As of May 9, 2020, it was discovered that the authentication endpoint was changed from,
	// "api.dominos.com/as/token.oauth2"
	// to,
	// "authproxy.dominos.com/auth-proxy-service/login".
	// I am documenting this change just in case it every changed back in the future.
	//
	// TODO: See comment above. Possible solutions are try both oauth endpoints or let users specify which to use.
	oauthURL = &url.URL{
		Scheme: "https",
		Host:   "authproxy.dominos.com",
		Path:   "/auth-proxy-service/login",
	}

	loginURL = &url.URL{
		Scheme: "https",
		Host:   orderHost,
		Path:   "/power/login",
	}
)

func authorize(c *http.Client, username, password string) error {
	tok, err := gettoken(username, password)
	if err != nil {
		return err
	}
	c.Transport = tok
	return nil
}

var noRedirects = func(r *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// // Token is a JWT that can be used as a transport for an http.Client
// type token struct {
// 	// AccessToken is the actual web token
// 	AccessToken string `json:"access_token"`
// 	// RefreshToken is the secret used to refresh this token
// 	RefreshToken string `json:"refresh_token,omitempty"`
// 	// Type is the type of token
// 	Type string `json:"token_type"`
// 	// ExpiresIn is the time in seconds that it takes for the token to
// 	// expire.
// 	ExpiresIn int `json:"expires_in"`

// 	transport http.RoundTripper
// }

// func (t *token) authorization() string {
// 	return fmt.Sprintf("%s %s", t.Type, t.AccessToken)
// }

// func (t *token) RoundTrip(req *http.Request) (*http.Response, error) {
// 	req.Header.Set("Authorization", t.authorization())
// 	auth.SetDawgUserAgent(req.Header)
// 	return t.transport.RoundTrip(req)
// }

var scopes = []string{
	"customer:card:read",
	"customer:profile:read:extended",
	"customer:orderHistory:read",
	"customer:card:update",
	"customer:profile:read:basic",
	"customer:loyalty:read",
	"customer:orderHistory:update",
	"customer:card:create",
	"customer:loyaltyHistory:read",
	"order:place:cardOnFile",
	"customer:card:delete",
	"customer:orderHistory:create",
	"customer:profile:update",
	"easyOrder:optInOut",
	"easyOrder:read",
}

func gettoken(username, password string) (*auth.Token, error) {
	data := url.Values{
		"grant_type":   {"password"},
		"client_id":    {"nolo-rm"}, // nolo-rm if you want a refresh token, or just nolo for temporary token
		"validator_id": {"VoldemortCredValidator"},
		"scope":        {strings.Join(scopes, " ")},
		"username":     {username},
		"password":     {password},
	}
	req := newPostReq(oauthURL, data)
	resp, err := orderClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := struct {
		*auth.Token
		*auth.Error
	}{Token: auth.NewToken(), Error: nil}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Token, nil
}

func login(c *client) (*UserProfile, error) {
	data := url.Values{
		"loyaltyIsActive": {"true"},
		"rememberMe":      {"true"},
	}
	req := newPostReq(loginURL, data)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	profile := &UserProfile{cli: c}
	b, err := ioutil.ReadAll(res.Body)
	if err = errpair(err, dominosErr(b)); err != nil {
		return nil, err
	}
	return profile, json.Unmarshal(b, profile)
}

func newPostReq(u *url.URL, vals url.Values) *http.Request {
	return &http.Request{
		Method:     "POST",
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       u.Host,
		Header: http.Header{
			"Content-Type": {
				"application/x-www-form-urlencoded; charset=UTF-8"}},
		URL:  u,
		Body: ioutil.NopCloser(strings.NewReader(vals.Encode())),
	}
}

type client struct {
	*http.Client
	host string
}

func (c *client) do(req *http.Request) ([]byte, error) {
	return do(c.Client, req)
}

func do(d doer, req *http.Request) ([]byte, error) {
	var buf bytes.Buffer
	resp, err := d.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dawg.do: bad status code %s", resp.Status)
	}
	_, err = buf.ReadFrom(resp.Body)
	if bytes.HasPrefix(bytes.ToLower(buf.Bytes()[:15]), []byte("<!doctype html>")) {
		return nil, errpair(err, errors.New("got html response"))
	}
	return buf.Bytes(), err
}

func (c *client) dojson(v interface{}, r *http.Request) (err error) {
	return dojson(c, v, r)
}

func dojson(d doer, v interface{}, r *http.Request) (err error) {
	resp, err := d.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *client) get(path string, params URLParam) ([]byte, error) {
	if params == nil {
		params = &Params{}
	}
	return c.do(&http.Request{
		Method: "GET",
		Host:   c.host,
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
		URL: &url.URL{
			Scheme:   "https",
			Host:     c.host,
			Path:     path,
			RawQuery: params.Encode(),
		},
	})
}

func (c *client) post(path string, params URLParam, r io.Reader) ([]byte, error) {
	if params == nil {
		params = &Params{}
	}
	rc, ok := r.(io.ReadCloser)
	if !ok && r != nil {
		rc = ioutil.NopCloser(r)
	}
	return c.do(&http.Request{
		Method: "POST",
		Host:   c.host,
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
		Body:   rc,
		URL: &url.URL{
			Scheme:   "https",
			Host:     c.host,
			Path:     path,
			RawQuery: params.Encode(),
		},
	})
}

func unmarshalToken(r io.ReadCloser, t *auth.Token) error {
	buf := new(bytes.Buffer)
	defer r.Close()

	_, e1 := buf.ReadFrom(r)
	err := errpair(e1, json.Unmarshal(buf.Bytes(), t))
	if err != nil {
		return err
	}
	e := &tokenError{}
	// if there is no token error the the json parsing will fail
	json.Unmarshal(buf.Bytes(), e)
	if len(e.Err) > 0 || len(e.ErrorDesc) > 0 {
		return e
	}
	return nil
}

type tokenError struct {
	Err       string `json:"error"`
	ErrorDesc string `json:"error_description"`
}

func (e *tokenError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.ErrorDesc)
}

func newTokErr(r io.Reader) error {
	e := &tokenError{}
	if err := json.NewDecoder(r).Decode(e); err != nil {
		return err
	}
	if len(e.Err) > 0 || len(e.ErrorDesc) > 0 {
		return e
	}
	return nil
}
