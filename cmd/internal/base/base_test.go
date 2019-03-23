package base

import (
	"fmt"
	"testing"
)

func TestCommand(t *testing.T) {
	c := &Command{}
	t.Error("what")
	fmt.Printf("%+v\n", c)
}
