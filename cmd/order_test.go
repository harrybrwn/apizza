package cmd

import (
	"testing"
)

func testOrderNew(t *testing.T) {
	order := newOrderCommand()
	new := newBuilder().newNewOrderCmd()

	new.command().Flags().Parse([]string{"--name=testorder", "--products=12SCMEATZA"})
	err := new.run(new.command(), []string{})
	if err != nil {
		t.Error(err)
	}

	// args := []string{"testorder"}
	args := []string{}
	if err := order.run(order.command(), args); err != nil {
		t.Error(err)
	}
	if err := order.run(order.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}
}
