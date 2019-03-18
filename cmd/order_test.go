package cmd

import (
	"bytes"
	"fmt"
	"testing"
)

func testOrderNew(t *testing.T) {
	b := newBuilder()
	order := b.newOrderCommand().(*orderCommand)
	new := b.newNewOrderCmd().(*newOrderCmd)

	buf := &bytes.Buffer{}
	new.setOutput(buf)
	order.setOutput(buf)

	new.command().ParseFlags([]string{"--name=testorder", "--products=12SCMEATZA"})
	err := new.run(new.command(), []string{})
	if err != nil {
		t.Error(err)
	}
	buf.Reset()

	if err := order.run(order.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}

	expected := `testorder
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

func testOrderNewErr(t *testing.T) {
	new := newBuilder().newNewOrderCmd().(*newOrderCmd)
	if err := new.run(new.command(), []string{}); err == nil {
		t.Error("expected error")
	}
}

func testOrderRunAdd(t *testing.T) {
	order := newBuilder().newOrderCommand().(*orderCommand)

	buf := &bytes.Buffer{}
	order.setOutput(buf)

	if err := order.run(order.command(), []string{}); err != nil {
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

	order.command().ParseFlags([]string{"--add", "W08PBNLW,W08PPLNW"})
	if err := order.run(order.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "updated order successfully saved.\n" {
		t.Error("wrong ouput message")
		fmt.Println("expected:", "updated order successfully saved.")
		fmt.Println("got:", string(buf.Bytes()))
	}
}

func testOrderPriceOutput(t *testing.T) {
	order := newBuilder().newOrderCommand().(*orderCommand)

	buf := &bytes.Buffer{}
	order.setOutput(buf)

	order.price = true
	if err := order.run(order.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}

	expected := `testorder
  Price: 34.070000
  Products:
    12SCMEATZA - quantity: 1, options: map[]
    W08PBNLW - quantity: 1, options: map[]
    W08PPLNW - quantity: 1, options: map[]
  StoreID: 4336
  Method:  Carryout
  Address: {Street:1600 Pennsylvania Ave NW StreetNum: City:Washington DC State: Zip:20500 AddrType: StreetName:}
`
	if string(buf.Bytes()) != expected {
		t.Error("unexpected price output")
	}

	if err := order.run(order.command(), []string{"to-many", "args"}); err == nil {
		t.Error("expected error")
	}
}

func testOrderRunDelete(t *testing.T) {
	order := newBuilder().newOrderCommand().(*orderCommand)

	buf := &bytes.Buffer{}
	order.setOutput(buf)

	order.delete = true
	if err := order.run(order.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "testorder successfully deleted.\n" {
		t.Error("wrong output message")
		fmt.Println("got:", string(buf.Bytes()))
	}
	order.delete = false
	buf.Reset()

	order.command().ParseFlags([]string{})
	if err := order.run(order.command(), []string{}); err != nil {
		t.Error(err)
	}
	expected := `No orders saved.
`
	if string(buf.Bytes()) != expected {
		t.Error("wrong output")
		fmt.Println("expected:", expected)
		fmt.Println("got:", string(buf.Bytes()))
	}
	buf.Reset()

	if err := order.run(order.command(), []string{"not_a_real_order"}); err == nil {
		t.Error("expected error")
	}
}
