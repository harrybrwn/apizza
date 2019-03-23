package base

import (
	"testing"
)

func TestCommand(t *testing.T) {
	c := &Command{}
	if c == nil {
		t.Error("what?")
	}
}
