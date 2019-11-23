package tests

// import (
// 	"errors"
// 	"io"
// 	"reflect"
// 	"runtime"
// 	"testing"
// )

// // TestingFunc represents a test function.
// type TestingFunc func(*testing.T)

// // Runner is a structure that runs a list of tests with a setup functon
// // and a teardown function
// type Runner struct {
// 	Setup    func()
// 	Teardown func()
// 	T        *testing.T
// 	tests    []testing.InternalTest
// }

// // Run runs all of the Runner's tests
// func (r *Runner) Run() int {
// 	if r.Setup != nil {
// 		r.Setup()
// 	}
// 	if r.Teardown != nil {
// 		defer r.Teardown()
// 	}

// 	if r.T != nil {
// 		for _, test := range r.tests {
// 			r.T.Run(test.Name, test.F)
// 		}
// 		return 0
// 	}

// 	return testing.MainStart(matchString(matchStr), r.tests, nil, nil).Run()
// }

// // NewRunner returns a new runner
// func NewRunner(t *testing.T, setup, teardown func()) *Runner {
// 	return &Runner{T: t, Setup: setup, Teardown: teardown}
// }

// // AddTest adds any number of test functions as arguments.
// func (r *Runner) AddTest(funcs ...TestingFunc) {
// 	for _, f := range funcs {
// 		r.tests = append(r.tests, testing.InternalTest{
// 			Name: testName(f),
// 			F:    f,
// 		})
// 	}
// }

// // Reset will remove all of the tests stored in the Runner.
// func (r *Runner) Reset() {
// 	r.tests = []testing.InternalTest{}
// }

// // Wrap will wrap a test in a start function and an end function. start and end
// // will be run as part of the test
// func (r *Runner) Wrap(test func(*testing.T), start, end func() error) func(*testing.T) {
// 	return func(t *testing.T) {
// 		if err := start(); err != nil {
// 			t.Errorf("Wrapped function %s (start): %s", testName(test), err)
// 		}

// 		test(t)

// 		if err := end(); err != nil {
// 			t.Errorf("Wrapped function %s (end): %s", testName(test), err)
// 		}
// 	}
// }

// func matchStr(pat, str string) (bool, error) { return true, nil }

// func testName(f interface{}) string {
// 	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
// }

// var errMatchStr = errors.New("tests: unexpected use of tests.Runner")

// // matchString and its methods are from the Go standard library.
// type matchString func(pat, str string) (bool, error)

// func (f matchString) MatchString(pat, str string) (bool, error)   { return f(pat, str) }
// func (f matchString) StartCPUProfile(w io.Writer) error           { return errMatchStr }
// func (f matchString) StopCPUProfile()                             {}
// func (f matchString) WriteProfileTo(string, io.Writer, int) error { return errMatchStr }
// func (f matchString) ImportPath() string                          { return "" }
// func (f matchString) StartTestLog(io.Writer)                      {}
// func (f matchString) StopTestLog() error                          { return errMatchStr }
