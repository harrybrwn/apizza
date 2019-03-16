package cmd

import (
	"bytes"
	"fmt"
	"testing"
)

func testOrderNew(t *testing.T) {
	b := newBuilder()
	order := b.newOrderCommand().(*orderCommand)
	new := b.newNewOrderCmd()

	new.command().Flags().Parse([]string{"--name=testorder", "--products=12SCMEATZA"})
	err := new.run(new.command(), []string{})
	if err != nil {
		t.Error(err)
	}

	buf := &bytes.Buffer{}
	order.output = buf
	args := []string{}
	if err := order.run(order.command(), args); err != nil {
		t.Error(err)
	}

	expected := `Your Orders:
  testorder
`
	if string(buf.Bytes()) != expected {
		t.Error("wrong output from apizza order")
		fmt.Println("got this:", string(buf.Bytes()))
		fmt.Println("expected this:", expected)
	}
	buf.Reset()
	if err := order.run(order.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}

	expected = `testorder
  Products:
    12SCMEATZA - quantity: 1, options: map[]
  StoreID: 4336
  Method:  Carryout
  Address: {Street:1600 Pennsylvania Ave NW StreetNum: City:Washington DC State: Zip:20500 AddrType: StreetName:}
`
	if string(buf.Bytes()) != expected {
		t.Error("wrong output from apizza order")
		fmt.Println("got this:", string(buf.Bytes()))
		fmt.Println("expected this:", expected)
	}
}
