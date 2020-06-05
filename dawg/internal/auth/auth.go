package auth

import (
	"fmt"
	"net/http"
	"time"
)

// NewToken returns an initialized transport.
func NewToken() *Token {
	return &Token{transport: http.DefaultTransport}
}

// Token is a JWT that can be used as a transport for an http.Client
type Token struct {
	// AccessToken is the actual web token
	AccessToken string `json:"access_token"`
	// RefreshToken is the secret used to refresh this token
	RefreshToken string `json:"refresh_token,omitempty"`
	// Type is the type of token
	Type string `json:"token_type"`
	// ExpiresIn is the time in seconds that it takes for the token to
	// expire.
	ExpiresIn int `json:"expires_in"`

	transport http.RoundTripper
}

func (t *Token) authorization() string {
	return fmt.Sprintf("%s %s", t.Type, t.AccessToken)
}

// RoundTrip implements the http.RoundTripper interface.
func (t *Token) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.authorization())
	SetDawgUserAgent(req.Header)
	return t.transport.RoundTrip(req)
}

func (t *Token) SetTransport(rt http.RoundTripper) {
	t.transport = rt
}

// Error is an error that is returned by the oauth endpoint.
type Error struct {
	Err       string `json:"error"`
	ErrorDesc string `json:"error_description"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.ErrorDesc)
}

// SetDawgUserAgent sets the package user agent
func SetDawgUserAgent(head http.Header) {
	head.Set(
		"User-Agent",
		"Dominos API Wrapper for GO - "+time.Now().String(),
	)
}

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

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
