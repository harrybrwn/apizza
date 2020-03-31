package command

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/harrybrwn/apizza/pkg/tests"
)

var testconfigjson = `
{
	"name":"joe","email":"nojoe@mail.com", "phone":"1231231234",
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
phone: "1231231234"
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
	r.ConfigSetup([]byte(testconfigjson))
	defer func() {
		r.CleanUp()
		// config.SetNonFileConfig(cfg) // for test compatability
	}()
	// check(json.Unmarshal([]byte(testconfigjson), r.Config()), "json")
	if err := json.Unmarshal([]byte(testconfigjson), r.Config()); err != nil {
		t.Fatal(err)
	}

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
}

func TestConfigCmd(t *testing.T) {
	r := cmdtest.NewRecorder()
	c := NewConfigCmd(r).(*configCmd)
	r.ConfigSetup([]byte(testconfigjson))
	defer func() {
		r.CleanUp()
		// config.SetNonFileConfig(cfg) // for test compatability
	}()

	c.file = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	c.file = false
	r.Compare(t, "\n")
	r.ClearBuf()
	c.dir = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	r.Compare(t, "\n")
	r.ClearBuf()

	err := json.Unmarshal([]byte(testconfigjson), r.Config())
	if err != nil {
		t.Error(err)
	}
	c.dir = false
	c.getall = true
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	r.Compare(t, testConfigOutput)
	r.ClearBuf()
	c.getall = false
	cmdUseage := c.Cmd().UsageString()
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	r.Compare(t, cmdUseage)
	r.ClearBuf()
	if err := c.Run(c.Cmd(), []string{}); err != nil {
		t.Error(err)
	}
	r.Compare(t, c.Cmd().UsageString())
}

func TestConfigEdit(t *testing.T) {
	r := cmdtest.NewRecorder()
	c := NewConfigCmd(r).(*configCmd)
	err := config.SetConfig(".config/apizza/tests", r.Conf)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = errs.Pair(r.DB().Destroy(), os.RemoveAll(config.Folder()))
		if err != nil {
			t.Error()
		}
		// config.SetNonFileConfig(cfg) // for test compatability
	}()

	os.Setenv("EDITOR", "cat")
	c.edit = true
	exp := `{
    "Name": "",
    "Email": "",
    "Phone": "",
    "Address": {
        "Street": "",
        "CityName": "",
        "State": "",
        "Zipcode": ""
    },
    "Card": {
        "Number": "",
        "Expiration": ""
    },
    "Service": "Delivery"
}`
	t.Run("edit output", func(t *testing.T) {
		if os.Getenv("TRAVIS") == "true" {
			// for some reason, 'cat' in travis gives no output
			t.Skip()
		}
		tests.CompareOutput(t, exp, func() {
			if err = c.Run(c.Cmd(), []string{}); err != nil {
				t.Error(err)
			}
		})
	})
	c.edit = false

	err = json.Unmarshal([]byte(testconfigjson), r.Config())
	if err != nil {
		t.Error(err)
	}
	a := config.Get("address")
	if a == nil {
		t.Error("should not be nil")
	}
	addr, ok := a.(obj.Address)
	if !ok {
		t.Error("this should be an obj.Address")
	}
	if addr.City() != "Washington DC" {
		t.Error("bad config address city")
	}
}

func TestConfigGet(t *testing.T) {
	conf := &cli.Config{}
	config.SetNonFileConfig(conf) // don't want it to over ride the file on disk
	if err := json.Unmarshal([]byte(testconfigjson), conf); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}

	args := []string{"email", "name"}
	err := get(args, buf)
	if err != nil {
		t.Error(err)
	}
	if err := configGetCmd.RunE(configGetCmd, []string{"email", "name"}); err != nil {
		t.Error(err)
	}
	tests.Compare(t, buf.String(), "nojoe@mail.com\njoe\n")
	buf.Reset()

	args = []string{}
	err = get(args, buf)
	if err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}

	args = []string{"nonExistantKey"}
	err = get(args, buf)
	if err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}

	if err := configGetCmd.RunE(configGetCmd, []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}
	if err := configGetCmd.RunE(configGetCmd, []string{"nonExistantKey"}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "cannot find nonExistantKey" {
		t.Error("wrong error message, got:", err.Error())
	}
}

func TestConfigSet(t *testing.T) {
	// c := newConfigSet() //.(*configSetCmd)
	conf := &cli.Config{}
	config.SetNonFileConfig(conf) // don't want it to over ride the file on disk
	if err := json.Unmarshal([]byte(cmdtest.TestConfigjson), conf); err != nil {
		t.Fatal(err)
	}

	if err := configSetCmd.RunE(configSetCmd, []string{"name=someNameOtherThanJoe"}); err != nil {
		t.Error(err)
	}
	if config.GetString("name") != "someNameOtherThanJoe" {
		t.Error("did not set the name correctly")
	}
	if err := configSetCmd.RunE(configSetCmd, []string{}); err == nil {
		t.Error("expected error")
	} else if err.Error() != "no variable given" {
		t.Error("wrong error message, got:", err.Error())
	}
	if err := configSetCmd.RunE(configSetCmd, []string{"nonExistantKey=someValue"}); err == nil {
		t.Error("expected error")
	}
	if err := configSetCmd.RunE(configSetCmd, []string{"badformat"}); err == nil {
		t.Error(err)
	} else if err.Error() != "use '<key>=<value>' format (no spaces), use <key>='-' to set as empty" {
		t.Error("wrong error message, got:", err.Error())
	}
}
