package dawg

import (
	"net/http"
	"net/url"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	// WarnigStatus is the status code dominos serves use for a warning
	WarnigStatus = 1
	// FailureStatus  is the status code dominos serves use for a failure
	FailureStatus = -1
	// OkStatus  is the status code dominos serves use to signify no problems
	OkStatus = 0
	host     = "order.dominos.com"
)

var (
	// Lang is the package language variable
	Lang = "en"
	// Warnings is a package switch for turning warings on or off
	Warnings = false

	cli = &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	errCodes = map[int]string{
		FailureStatus: "Failure -1",
		WarnigStatus:  "Warning 1",
		OkStatus:      "Ok 0",
	}
)

// SetLang sets the package language variable which is used by default to
// send info to dominos.
func SetLang(code string) {
	Lang = code
}

// DominosError represents an error sent back by the dominos servers
type DominosError struct {
	Status      int                 `json:"Status"`
	StatusItems []map[string]string `json:"StatusItems"`
	Order       struct {
		Status      int                      `json:"Status"`
		StatusItems []map[string]interface{} `json:"StatusItems"`
	} `json:"Order"`
	Msg string
}

// Init initializes the error from json data.
func (err *DominosError) Init(jsonData []byte) error {
	return json.Unmarshal(jsonData, err)
}

func (err *DominosError) Error() string {
	var errmsg string
	// errmsg += "Tip: use err.IsWarning() or err.IsFailure() for handling type DominosError.\n"
	// errmsg += "\tex. if e, ok := err.(*DominosError); ok && e.IsFailure() { panic(e) }\n\n"
	for i := range err.StatusItems {
		if v, ok := err.StatusItems[i]["Code"]; ok {
			errmsg += fmt.Sprintf("Dominos %s:\n", v)
		}
	}
	for _, item := range err.Order.StatusItems {
		errmsg += "    " + item["Code"].(string)
		if msg, ok := item["Message"]; ok {
			errmsg += ":\n        " + msg.(string)
		} else if msg, ok := item["PulseText"]; ok {
			errmsg += ":\n        " + msg.(string)
		} else {
			errmsg += "\n"
		}
	}
	return errmsg
}

// IsWarning returns true when the error sent by dominos is a warning
func (err *DominosError) IsWarning() bool {
	return err.Status == WarnigStatus
}

// IsFailure returns true if the error that dominos sent back prevents the
// system from working
func (err *DominosError) IsFailure() bool {
	return err.Status == FailureStatus
}

func get(path string, params URLParam) ([]byte, error) {
	if params == nil {
		params = &Params{}
	}
	// to prevent future suffering
	if path[0] != '/' {
		path = "/" + path
	}
	req := &http.Request{
		Method: "GET",
		Host:   host,
		Proto:  "HTTP/1.1",
		Header: http.Header{},
		URL: &url.URL{
			Scheme:   "https",
			Host:     host,
			Path:     path,
			RawQuery: params.Encode(),
		},
	}
	t := time.Now()
	req.Header.Add("User-Agent", "Dominos API Wrapper for GO"+t.String())
	resp, err := cli.Do(req)

	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return buf.Bytes(), fmt.Errorf("bad response code: %d", resp.StatusCode)
	}
	return buf.Bytes(), nil
}

func post(path string, data []byte) ([]byte, error) {
	if path[0] != '/' {
		path = "/" + path
	}
	req := &http.Request{
		Method: "POST",
		Host:   host,
		Proto:  "HTTP/1.1",
		Body:   ioutil.NopCloser(bytes.NewReader(data)),
		URL: &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   path,
		},
	}
	var buf bytes.Buffer
	if resp, err := cli.Do(req); err == nil {
		_, err = buf.ReadFrom(resp.Body)
	} else {
		return nil, err
	}
	return buf.Bytes(), nil
}
