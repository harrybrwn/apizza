package errs

import (
	"bytes"
	"fmt"
)

// Append errors together.
func Append(err error, errlist ...error) error {
	if errlist == nil {
		return err
	}

	var (
		check bool
		res   *linearError
	)

	if res, check = err.(*linearError); !check {
		res = &linearError{}
		res.errList = append(res.errList, err)
	}
	check = true
	for _, e := range errlist {
		if e != nil {
			res.errList = append(res.errList, e)
			check = false
		}
	}
	if check {
		return nil
	}
	return res
}

// Pair two errors together. Returns nil if both errors are nil. If one of the errors
// given is nil, then Pair will return the one that is not. Otherwise
// a linearError containing both errors is returned.
func Pair(first, second error) error {
	if first == nil || second == nil {
		if second != nil {
			return second
		}
		return first
	}
	return &linearError{[]error{first, second}}
}

type linearError struct {
	errList []error
}

func newline(e error, n int) *linearError {
	err := &linearError{errList: make([]error, n)}
	if e != nil {
		err.errList[0] = e
	}
	return err
}

func (le *linearError) append(ers []error) *linearError {
	for _, e := range ers {
		if e != nil {
			le.errList = append(le.errList, e)
		}
	}
	return le
}

func (le *linearError) add(e error) *linearError {
	le.errList = append(le.errList, e)
	return le
}

// Error for a linearError will print out all of its errors as a list.
func (le *linearError) Error() string {
	var (
		buf  = new(bytes.Buffer)
		list = []error{}
	)

	le.flatten(&list)
	buf.WriteString("Errors:\n")
	for i, e := range list {
		fmt.Fprintf(buf, "  %d. ", i+1)
		buf.WriteString(e.Error())
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (le *linearError) flatten(arr *[]error) {
	for _, e := range le.errList {
		if list, ok := e.(*linearError); ok {
			list.flatten(arr)
		} else if e != nil {
			*arr = append(*arr, e)
		}
	}
}
