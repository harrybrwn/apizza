// +build !go1.14

package cmdtest

// CleanUp cleans up all the TestRecorder's allocated recourses
func (tr *TestRecorder) CleanUp() {
	tr.r.CleanUp()
}

// init is a noop for builds below 1.14
func (tr *TestRecorder) init() {}
