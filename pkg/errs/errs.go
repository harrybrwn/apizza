package errs

import (
	"fmt"
	"log"
	"os"
)

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

// Handle errors and exit.
func Handle(e error, msg string, exitcode int) {
	if e == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, e)
	log.Printf("%s: '%s'\n", msg, e)
	os.Exit(exitcode)
}
