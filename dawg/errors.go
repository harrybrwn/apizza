package dawg

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const (
	// WarningStatus is the status code dominos serves use for a warning
	WarningStatus = 1

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
		WarningStatus: "Warning 1",
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
	var (
		buf      = new(bytes.Buffer)
		item     statusItem
		haspulse bool
	)

	for _, item = range err.StatusItems {
		fmt.Fprintf(buf, "Dominos %s (%d)\n", item.Code, err.Status)
	}
	for _, item = range err.Order.StatusItems {
		haspulse = item.PulseText != ""
		fmt.Fprint(buf, "    ")
		if !haspulse {
			switch err.Order.Status {
			case WarningStatus:
				fmt.Fprintf(buf, "Warning ")
			case FailureStatus:
				fmt.Fprintf(buf, "Failure ")
			}
		}

		if item.Code != "" {
			fmt.Fprintf(buf, "Code: '%s'", item.Code)
		}
		if item.Message != "" {
			fmt.Fprintf(buf, ":\n        %s\n", item.Message)
		} else if haspulse {
			fmt.Fprintf(buf, "    PulseCode %d: %s", item.PulseCode, item.PulseText)
		} else {
			fmt.Fprint(buf, "\n")
		}
	}
	return buf.String()
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
	return e.Status == WarningStatus
}

// IsOk will tell you if the error returned does not contain any fatal errors or
// warnings from Dominos' servers. Will return true if the error is nil.
func IsOk(err error) bool {
	if err == nil {
		return true
	}
	e, ok := isDominosErr(err)
	if !ok {
		return false
	}
	return e.Status == OkStatus
}

func isDominosErr(err error) (*DominosError, bool) {
	e, ok := err.(*DominosError)
	if !ok {
		return nil, false
	}
	return e, true
}

// because i want my errs.Pair function but i dont want to add it as a
// dependancy to the dawg package in case i ever want to separate them.
func errpair(first, second error) error {
	if first == nil || second == nil {
		if first != nil { // should check the first error first
			return first
		}
		return second
	}
	return &errorpair{first, second}
}

type errorpair struct {
	e1, e2 error
}

func (e *errorpair) Error() string {
	return fmt.Sprintf("error 1. %s\nerror 2. %s", e.e1.Error(), e.e2.Error())
}

func eatint(n int, e error) error {
	return e
}
