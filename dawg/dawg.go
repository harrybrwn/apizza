package dawg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/mitchellh/mapstructure"
)

const (
	// WarnigStatus is the status code dominos serves use for a warning
	WarnigStatus = 1

	// FailureStatus  is the status code dominos serves use for a failure
	FailureStatus = -1

	// OkStatus  is the status code dominos serves use to signify no problems
	OkStatus = 0

	// DefaultLang is the package language variable
	DefaultLang = "en"

	host = "order.dominos.com"
)

var (
	// Warnings is a package switch for turning warings on or off
	Warnings = false

	cli = &http.Client{
		Timeout: time.Duration(10 * time.Second),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	errCodes = map[int]string{
		FailureStatus: "Failure -1",
		WarnigStatus:  "Warning 1",
		OkStatus:      "Ok 0",
	}
)

func dominosErr(resp []byte) error {
	e := &DominosError{}
	if err := e.init(resp); err != nil {
		return err
	}

	if IsOk(e) {
		return nil
	}
	return e
}

// DominosError represents an error sent back by the dominos servers
type DominosError struct {
	Status      int
	StatusItems []statusItem
	Order       struct {
		Status      int
		StatusItems []statusItem
		OrderID     string
	}
	Msg     string
	fullErr map[string]interface{}
}

type statusItem struct {
	Code      string
	Message   string
	PulseCode int
	PulseText string
}

// init initializes the error from json data.
func (err *DominosError) init(jsonData []byte) error {
	err.fullErr = map[string]interface{}{}

	if e := json.Unmarshal(jsonData, &err.fullErr); e != nil {
		return e
	}
	return mapstructure.Decode(err.fullErr, err)
}

func (err *DominosError) Error() string {
	var errmsg string

	for _, item := range err.StatusItems {
		errmsg += fmt.Sprintf("Dominos %s:\n", item.Code)
	}
	for _, item := range err.Order.StatusItems {
		if item.Code != "" {
			errmsg += fmt.Sprintf("    Code: '%s'", item.Code)
		}
		if item.Message != "" {
			errmsg += fmt.Sprintf(":\n        %s\n", item.Message)
		} else if item.PulseText != "" {
			errmsg += fmt.Sprintf("    PulseCode %d:\n        %s", item.PulseCode, item.PulseText)
		} else {
			errmsg += "\n"
		}
	}
	return errmsg
}

// IsFailure will tell you if an error given by a function from the dawg package
// is an error thrown from dominos' servers.
func IsFailure(err error) bool {
	e, ok := isDominosErr(err)
	if !ok {
		return false
	}
	return e.Status == FailureStatus
}

// IsWarning will tell you if the error given contains a warning from the
// dominos server.
func IsWarning(err error) bool {
	e, ok := isDominosErr(err)
	if !ok {
		return false
	}
	return e.Status == WarnigStatus
}

// IsOk will tell you if the error returned does not contain any fatal errors or
// warnings from Dominos' servers.
func IsOk(err error) bool {
	e, ok := isDominosErr(err)
	if !ok {
		return false
	}
	return e.Status == OkStatus
}

func isDominosErr(err error) (*DominosError, bool) {
	if err == nil {
		return nil, false
	}
	e, ok := err.(*DominosError)
	if !ok {
		return nil, false
	}
	return e, true
}

func get(path string, params URLParam) ([]byte, error) {
	if params == nil {
		params = &Params{}
	}
	return send(&http.Request{
		Method: "GET",
		Host:   host,
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
		URL: &url.URL{
			Scheme:   "https",
			Host:     host,
			Path:     path,
			RawQuery: params.Encode(),
		},
	})
}

func post(path string, data []byte) ([]byte, error) {
	return send(&http.Request{
		Method: "POST",
		Host:   host,
		Proto:  "HTTP/1.1",
		Body:   ioutil.NopCloser(bytes.NewReader(data)),
		Header: make(http.Header),
		URL: &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   path,
		},
	})
}

func send(req *http.Request) ([]byte, error) {
	var buf bytes.Buffer

	req.Header.Add("User-Agent", "Dominos API Wrapper for GO - "+time.Now().String())

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return buf.Bytes(), fmt.Errorf("bad response code: %d", resp.StatusCode)
	}
	return buf.Bytes(), nil
}
