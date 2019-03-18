package cmd

import (
	"bytes"
	"fmt"
	"testing"
)

func testOrderNew(t *testing.T) {
	b := newBuilder()
	cart := b.newCartCmd().(*cartCmd)
	add := b.newAddOrderCmd().(*addOrderCmd)

	buf := &bytes.Buffer{}
	add.setOutput(buf)
	cart.setOutput(buf)

	add.command().ParseFlags([]string{"--name=testorder", "--products=12SCMEATZA"})
	err := add.run(add.command(), []string{})
	if err != nil {
		t.Error(err)
	}
	buf.Reset()

	if err := cart.run(cart.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}

	expected := `testorder
  Products:
    12SCMEATZA - quantity: 1, options: map[]
  StoreID: 4336
  Method:  Carryout
  Address: &{StreetLineOne:1600 Pennsylvania Ave NW StreetNum:1600 CityName:Washington DC State: Zipcode:20500 AddrType: StreetName:Pennsylvania Ave NW}
`
	if string(buf.Bytes()) != expected {
		t.Error("wrong output from apizza order")
		fmt.Println("got this:", string(buf.Bytes()))
		fmt.Println("expected this:", expected)
	}
}

func testOrderNewErr(t *testing.T) {
	new := newBuilder().newAddOrderCmd().(*addOrderCmd)
	if err := new.run(new.command(), []string{}); err == nil {
		t.Error("expected error")
	}
}

func testOrderRunAdd(t *testing.T) {
	cart := newBuilder().newCartCmd().(*cartCmd)

	buf := &bytes.Buffer{}
	cart.setOutput(buf)

	if err := cart.run(cart.command(), []string{}); err != nil {
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

	cart.command().ParseFlags([]string{"--add", "W08PBNLW,W08PPLNW"})
	if err := cart.run(cart.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "updated order successfully saved.\n" {
		t.Error("wrong ouput message")
		fmt.Println("expected:", "updated order successfully saved.")
		fmt.Println("got:", string(buf.Bytes()))
	}
}

func testOrderPriceOutput(t *testing.T) {
	cart := newBuilder().newCartCmd().(*cartCmd)

	buf := &bytes.Buffer{}
	cart.setOutput(buf)

	cart.price = true
	if err := cart.run(cart.command(), []string{"testorder"}); err != nil {
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
  Address: &{StreetLineOne:1600 Pennsylvania Ave NW StreetNum:1600 CityName:Washington DC State: Zipcode:20500 AddrType: StreetName:Pennsylvania Ave NW}
`
	if string(buf.Bytes()) != expected {
		t.Error("unexpected price output")
	}

	if err := cart.run(cart.command(), []string{"to-many", "args"}); err == nil {
		t.Error("expected error")
	}
}

func testOrderRunDelete(t *testing.T) {
	cart := newBuilder().newCartCmd().(*cartCmd)

	buf := &bytes.Buffer{}
	cart.setOutput(buf)

	cart.delete = true
	if err := cart.run(cart.command(), []string{"testorder"}); err != nil {
		t.Error(err)
	}
	if string(buf.Bytes()) != "testorder successfully deleted.\n" {
		t.Error("wrong output message")
		fmt.Println("got:", string(buf.Bytes()))
	}
	cart.delete = false
	buf.Reset()

	cart.command().ParseFlags([]string{})
	if err := cart.run(cart.command(), []string{}); err != nil {
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

	if err := cart.run(cart.command(), []string{"not_a_real_order"}); err == nil {
		t.Error("expected error")
	}
}
