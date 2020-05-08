// +build go1.14

package cmdtest

import "github.com/harrybrwn/apizza/pkg/tests"

// CleanUp is a noop for go1.14
func (tr *TestRecorder) CleanUp() {}

func (tr *TestRecorder) init() {
	tr.t.Cleanup(func() {
		tr.Recorder.CleanUp()
		tests.ResetHelpers()
	})
}
