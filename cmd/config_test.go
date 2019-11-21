package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/tests"
)

var testconfigjson = `
{
	"name":"joe","email":"nojoe@mail.com",
	"address":{
		"street":"1600 Pennsylvania Ave NW",
		"cityName":"Washington DC",
		"state":"","zipcode":"20500"
	},
	"card":{"number":"","expiration":"","cvv":""},
	"service":"Carryout"
}`

var testConfigOutput = `name: "joe"
email: "nojoe@mail.com"
address:
  street: "1600 Pennsylvania Ave NW"
  cityname: "Washington DC"
  state: ""
  zipcode: "20500"
card:
  number: ""
  expiration: ""
service: "Carryout"
`

func TestConfigStruct(t *testing.T) {
	r := cmdtest.NewRecorder()
	defer r.CleanUp()
	r.ConfigSetup()
	check(json.Unmarshal([]byte(testconfigjson), r.Config()), "json")

	if r.Config().Get("name").(string) != "joe" {
		t.Error("wrong value")
	}
	if err := r.Config().Set("name", "not joe"); err != nil {
		t.Error(err)
	}
	if r.Config().Get("Name").(string) != "not joe" {
		t.Error("wrong value")
	}
	if err := r.Config().Set("name", "joe"); err != nil {
		t.Error(err)
	}
	config.SetNonFileConfig(cfg) // reset the global config for compatability
}

func testConfigCmd(t *testing.T) {
	r := cmdtest.NewRecorder()
	c := newConfigCmd(r).(*configCmd)
	buf := &bytes.Buffer{}
	c.SetOutput(buf)
	c.file = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	c.file = false
	tests.Compare(t, string(buf.Bytes()), "\n")
	buf.Reset()
	c.dir = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), "\n")
	c.dir = false
	buf.Reset()
	c.getall = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), testConfigOutput)
	c.getall = false
	buf.Reset()
	cmdUseage := c.Cmd().UsageString()
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), cmdUseage)
	buf.Reset()
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), c.Cmd().UsageString())
}

func testConfigGet(t *testing.T) {
	c := newConfigGet()
	buf := &bytes.Buffer{}
	c.SetOutput(buf)
	if err := c.Run(c.Cmd(), []string{"email", "name"}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, string(buf.Bytes()), "nojoe@mail.com\njoe\n")
	buf.Reset()
	if err := c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}
	if err := c.Run(c.Cmd(), []string{"nonExistantKey"}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func testConfigSet(t *testing.T) {
	c := newConfigSet() //.(*configSetCmd)
	if err := c.Run(c.Cmd(), []string{"name=someNameOtherThanJoe"}); err != nil {
		t.Error(err)
	}
	if cfg.Name != "someNameOtherThanJoe" {
		t.Error("did not set the name correctly")
	}
	if err := c.Run(c.Cmd(), []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}
	if err := c.Run(c.Cmd(), []string{"nonExistantKey=someValue"}); err == nil {
		t.Error("expected error")
	}
	if err := c.Run(c.Cmd(), []string{"badformat"}); err == nil {
		t.Error(err)
	} else if err.Error() != "use '<key>=<value>' format (no spaces), use <key>='-' to set as empty" {
		t.Error("wrong error message, got:", err.Error())
	}
}
