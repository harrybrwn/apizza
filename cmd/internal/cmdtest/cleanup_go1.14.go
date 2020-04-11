// +build go1.14

package cmdtest

// CleanUp is a noop for go1.14
func (tr *TestRecorder) CleanUp() {}

func (tr *TestRecorder) init() {
	tr.t.Cleanup(tr.Recorder.CleanUp)
}
