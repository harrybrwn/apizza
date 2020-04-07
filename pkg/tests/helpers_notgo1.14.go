// +build !go1.14

package tests

import "testing"

func initHelpers(t *testing.T) {
	currentTest = &struct {
		name string
		t    *testing.T
	}{
		name: t.Name(),
		t:    t,
	}
}
