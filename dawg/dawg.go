package dawg

import (
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"

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

	// DefaultLang is the package language variable
	DefaultLang = "en"
)

var (
	// Warnings is a package switch for turning warings on or off
	Warnings = false

	cli = &http.Client{
		Timeout: time.Duration(5 * time.Second),
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

func dominosErr(resp []byte) *DominosError {
	e := &DominosError{}
	if err := e.init(resp); err != nil {
		panic(err)
	}

	if e.IsOk() {
		return nil
	}
	return e
}

// DominosError represents an error sent back by the dominos servers
type DominosError struct {
	Status      int          `json:"Status"`
	StatusItems []statusItem `json:"StatusItems"`
	Order       struct {
		Status      int          `json:"Status"`
		StatusItems []statusItem `json:"StatusItems"`
	} `json:"Order"`
	Msg     string
	fullErr map[string]interface{}
}

type statusItem struct {
	Code      string
	Message   string
	PulseCode int
	PulseText string
}

// Init initializes the error from json data.
func (err *DominosError) init(jsonData []byte) error {
	err.fullErr = map[string]interface{}{}

	if err := json.Unmarshal(jsonData, &err.fullErr); err != nil {
		return err
	}
	return mapstructure.Decode(err.fullErr, err)
}

func (err *DominosError) Error() string {
	var errmsg string
	for _, item := range err.StatusItems {
		errmsg += fmt.Sprintf("Dominos %s:\n", item.Code)
	}
	for _, item := range err.Order.StatusItems {
		errmsg += "    " + item.Code
		if item.Message != "" {
			errmsg += ":\n        " + item.Message
		} else if item.PulseText != "" {
			errmsg += fmt.Sprint(item.PulseCode) + ":\n        " + item.PulseText
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

// IsOk returns true is the error is not a failure else returns false
func (err *DominosError) IsOk() bool {
	return err.Status != FailureStatus
}

func get(path string, params URLParam) ([]byte, error) {
	if params == nil {
		params = &Params{}
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
	req.Header.Add("User-Agent", "Dominos API Wrapper for GO-"+t.String())
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
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return buf.Bytes(), fmt.Errorf("bad response code: %d", resp.StatusCode)
	}
	return buf.Bytes(), nil
}
