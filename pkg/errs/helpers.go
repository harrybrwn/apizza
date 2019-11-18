package errs

// EatInt will eat an int and return the error. This is good if you want
// to chain calls to an io.Writer or io.Reader.
func EatInt(n int, e error) error {
	return e
}
