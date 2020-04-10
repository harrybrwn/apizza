package internal

import "errors"

var (
	// ErrNoAddress is the error found when the cli could no find an address
	ErrNoAddress = errors.New("no address found. (see 'apizza address' or 'apizza config')")

	// ErrNoOrderName is the error raised when the is no order name given to the
	// cart or the order commands.
	ErrNoOrderName = errors.New("No order name... use '--name=<order name>' or give name as an argument")
)
