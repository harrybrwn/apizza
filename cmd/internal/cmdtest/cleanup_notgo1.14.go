// +build !go1.14

package cmdtest

import "github.com/harrybrwn/apizza/pkg/tests"

// CleanUp cleans up all the TestRecorder's allocated recourses
func (tr *TestRecorder) CleanUp() {
	tr.Recorder.CleanUp()
	tests.ResetHelpers()
}

// init is a noop for builds below 1.14
func (tr *TestRecorder) init() {}
