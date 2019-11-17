package errs

import "fmt"

// New returns a new basic error.
func New(msg interface{}) error {
	return &basicError{msg}
}

type basicError struct {
	msg interface{}
}

func (e *basicError) Error() string {
	return fmt.Sprintf("%v", e.msg)
}
