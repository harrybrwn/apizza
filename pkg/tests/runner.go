package tests

import (
	"io"
	"reflect"
	"runtime"
	"testing"
)

// Runner is a structure that runs a list of tests with a setup functon
// and a teardown function
type Runner struct {
	Setup    func()
	Teardown func()
	T        *testing.T
	tests    []testing.InternalTest
}

// Run runs all of the Runner's tests
func (r *Runner) Run() int {
	r.Setup()
	defer r.Teardown()

	if r.T != nil {
		for _, test := range r.tests {
			r.T.Run(test.Name, test.F)
		}
		return 0
	}

	return testing.MainStart(matchString(matchStr), r.tests, nil, nil).Run()
}

// AddTest adds any number of test functions as arguments.
func (r *Runner) AddTest(funcs ...func(*testing.T)) {
	for _, f := range funcs {
		r.tests = append(r.tests, testing.InternalTest{
			Name: testName(f),
			F:    f,
		})
	}
}

func matchStr(pat, str string) (bool, error) { return true, nil }

func testName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// matchString and its methods are from the Go standard library.
type matchString func(pat, str string) (bool, error)

func (f matchString) MatchString(pat, str string) (bool, error) { return f(pat, str) }

// func (f matchString) StartCPUProfile(w io.Writer) error {return errors.New("testing: unexpected use of func Main")}
func (f matchString) StartCPUProfile(w io.Writer) error { return nil }
func (f matchString) StopCPUProfile()                   {}

// func (f matchString) WriteProfileTo(string, io.Writer, int) error {return errors.New("testing: unexpected use of func Main")}
func (f matchString) WriteProfileTo(string, io.Writer, int) error { return nil }
func (f matchString) ImportPath() string                          { return "" }
func (f matchString) StartTestLog(io.Writer)                      {}

// func (f matchString) StopTestLog() error {return errors.New("testing: unexpected use of func Main")}
func (f matchString) StopTestLog() error { return nil }
