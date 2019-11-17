package errs

import "bytes"

import "fmt"

// Append errors together.
func Append(err error, erlist ...error) error {
	if erlist == nil {
		return err
	} else if res, ok := err.(*linearError); ok {
		return res.append(erlist)
	}

	n := len(erlist)
	res := newline(err, n+1)
	k := 1
	for i := 0; i < n; i++ {
		if e, ok := erlist[i].(*linearError); ok {
			res.append(e.errList) // flatten any other linear errors found
		} else if e := erlist[i]; e != nil {
			res.errList[k] = e
			k++
		}
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

	e1, e1ok := first.(*linearError)
	e2, e2ok := second.(*linearError)
	if e1ok && !e2ok {
		return e1.add(second)
	}
	if !e1ok && e2ok {
		return e2.add(first)
	}
	if e1ok && e2ok {
		return e1.append(e2.errList)
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

func (le *linearError) Error() string {
	buf := new(bytes.Buffer)
	buf.WriteString("Errors:\n")
	le.write(buf, 1)
	return buf.String()
}

func (le *linearError) write(buf *bytes.Buffer, start int) int {
	var (
		i int
		e error
	)
	for i, e = range le.errList {
		if line, ok := e.(*linearError); ok {
			start += line.write(buf, i)
		} else {
			fmt.Fprintf(buf, "  %d. ", i+start)
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}
	}
	return i + start
}
