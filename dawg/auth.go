package dawg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type auth struct {
	username string
	password string
	token    *token
	cli      *http.Client
}

func newauth(username, password string) (*auth, error) {
	tok, err := gettoken(username, password)
	if err != nil {
		return nil, err
	}
	a := &auth{
		token:    tok,
		username: username,
		password: password,
		cli: &http.Client{
			Transport: tok,
			Timeout:   30 * time.Second,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	return a, nil
}

func emptyauth() *auth {
	return &auth{
		token: nil,
		cli: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

type token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Type         string `json:"token_type"`

	// ExpiresIn is the time in seconds that it takes for the token to
	// expire.
	ExpiresIn int `json:"expires_in"`

	transport http.RoundTripper
}

func (t *token) authorization() string {
	return fmt.Sprintf("%s %s", t.Type, t.AccessToken)
}

func (t *token) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", t.authorization())
	return t.transport.RoundTrip(req)
}

func gettoken(username, password string) (*token, error) {
	data := url.Values{
		"grant_type": {"password"},
		"client_id":  {"nolo-rm"}, // nolo-rm if you want a refresh token, or just nolo for temporary token
		"scope":      {"customer:card:read customer:profile:read:extended customer:orderHistory:read customer:card:update customer:profile:read:basic customer:loyalty:read customer:orderHistory:update customer:card:create customer:loyaltyHistory:read order:place:cardOnFile customer:card:delete customer:orderHistory:create customer:profile:update easyOrder:optInOut easyOrder:read"},
		"username":   {username},
		"password":   {password},
	}
	req, err := http.NewRequest(
		"POST", "https://api.dominos.com/as/token.oauth2",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	tok := &token{transport: http.DefaultTransport}
	if err = unmarshalToken(resp.Body, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

func (a *auth) login() (*UserProfile, error) {
	data := url.Values{
		"loyaltyIsActive": {"true"},
		"rememberMe":      {"false"},
		"u":               {a.username},
		"p":               {a.password},
	}
	req, err := http.NewRequest(
		"POST", "https://order.dominos.com/power/login",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	res, err := a.cli.Do(req)
	if err != nil {
		return nil, err
	}
	profile := new(UserProfile)
	if err = json.NewDecoder(res.Body).Decode(profile); err != nil {
		return nil, err
	}
	if err = res.Body.Close(); err != nil {
		return nil, err
	}

	profile.auth = a
	return profile, nil
}

type client struct {
	*http.Client
	host string
}

func (c *client) do(req *http.Request) ([]byte, error) {
	var buf bytes.Buffer
	req.Header.Add(
		"User-Agent",
		"Dominos API Wrapper for GO - "+time.Now().String(),
	)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return buf.Bytes(), fmt.Errorf("bad response code: %d", resp.StatusCode)
	}
	return buf.Bytes(), nil
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
	return send(&http.Request{
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

func unmarshalToken(r io.ReadCloser, t *token) error {
	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}
	if err := r.Close(); err != nil {
		return err
	}
	if err := json.NewDecoder(buf).Decode(t); err != nil {
		return err
	}
	return newTokenErr(buf.Bytes())
}

type tokenError struct {
	Err       string `json:"error"`
	ErrorDesc string `json:"error_description"`
}

func (e *tokenError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.ErrorDesc)
}

func newTokenErr(b []byte) error {
	e := &tokenError{}
	if err := json.Unmarshal(b, e); err != nil {
		return err
	}
	if len(e.Err) > 0 {
		return e
	}
	return nil
}