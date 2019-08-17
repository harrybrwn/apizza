package dawg

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	// WarnigStatus is the status code dominos serves use for a warning
	WarnigStatus = 1

	// FailureStatus  is the status code dominos serves use for a failure
	FailureStatus = -1

	// OkStatus  is the status code dominos serves use to signify no problems
	OkStatus = 0
)

var (
	// Warnings is a package switch for turning warings on or off
	Warnings = false

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
