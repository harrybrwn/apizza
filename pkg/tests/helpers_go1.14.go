// +build go1.14

package tests

import "testing"

func initHelpers(t *testing.T) {
	t.Cleanup(func() {
		currentTest = nil
		PrintErrType = false
	})
	currentTest = &struct {
		name string
		t    *testing.T
	}{
		name: t.Name(),
		t:    t,
	}
}
