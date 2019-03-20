package tests

import (
	"reflect"
	"runtime"
	"testing"
)

// Runner is a structure that runs a list of tests with a setup functon
// and a teardown function
type Runner struct {
	Setup    func()
	Teardown func()
	Tests    []func(*testing.T)
	T        *testing.T
}

// Run runs all of the Runner's tests
func (r *Runner) Run() {
	r.Setup()
	defer r.Teardown()

	for _, test := range r.Tests {
		r.T.Run(testName(test), test)
	}
}

// AddTest adds any number of test functions as arguments.
func (r *Runner) AddTest(funcs ...func(*testing.T)) {
	r.Tests = append(r.Tests, funcs...)
}

func testName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
