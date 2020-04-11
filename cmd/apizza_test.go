package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	fp "path/filepath"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/cmdtest"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestRunner(t *testing.T) {
	tests.InitHelpers(t)
	app := CreateApp(cmdtest.TempDB(), &cli.Config{}, nil)
	builder := cmdtest.NewRecorder()
	builder.ConfigSetup([]byte(cmdtest.TestConfigjson))

	tsts := []func(*testing.T){
		cli.WithCmds(testOrderNew, NewCartCmd(builder), newAddOrderCmd(builder)),
		cli.WithCmds(testAddOrder, NewCartCmd(builder), newAddOrderCmd(builder)),
		cli.WithCmds(testOrderNewErr, newAddOrderCmd(builder)),
		cli.WithCmds(testOrderRunAdd, NewCartCmd(builder)),
		withCartCmd(builder, testOrderPriceOutput),
		withCartCmd(builder, testAddToppings),
		withCartCmd(builder, testOrderRunDelete),
		withAppCmd(testAppRootCmdRun, app),
	}

	for i, tst := range tsts {
		t.Run(fmt.Sprintf("Test %d", i), tst)
	}

	builder.CleanUp()
	tests.Check(app.db.Destroy())

	msg := senderr(errs.New("this is an error"), "error message", 4)
	if msg.Code != 4 {
		t.Error("wrong code")
	}
}

func testAppRootCmdRun(t *testing.T, buf *bytes.Buffer, a *App) {
	a.Cmd().ParseFlags([]string{})
	tests.Check(a.Run(a.Cmd(), []string{}))
	if buf.String() != a.Cmd().UsageString() {
		t.Error("wrong output")
	}

	test = true
	tests.Check(a.Run(a.Cmd(), []string{}))
	test = false
	buf.Reset()

	tests.Check(a.prerun(a.Cmd(), []string{}))
	tests.Check(a.postrun(a.Cmd(), []string{}))

	if len(a.Cmd().Commands()) != 0 {
		t.Error("should not have commands yet")
	}
	tests.Check(a.Cmd().Execute())
}

func TestAppResetFlag(t *testing.T) {
	tests.InitHelpers(t)
	r := cmdtest.NewRecorder()
	a := CreateApp(r.ToApp())
	r.ConfigSetup([]byte(cmdtest.TestConfigjson))

	a.gOpts.ClearCache = true
	test = false
	tests.Check(a.Run(a.Cmd(), []string{}))
	if _, err := os.Stat(a.DB().Path()); os.IsExist(err) {
		t.Error("database should not exist")
	} else if !os.IsNotExist(err) {
		t.Error("database should not exist")
	}
	r.Compare(t, fmt.Sprintf("removing %s\n", a.DB().Path()))
	r.ClearBuf()
}

func TestAppStoreFinder(t *testing.T) {
	r := cmdtest.NewRecorder()
	defer r.CleanUp()
	a := CreateApp(r.ToApp())

	store := a.Store()
	if store == nil {
		t.Error("what")
	}
}

func setupTests() {
	// config.SetNonFileConfig(cfg) // don't want it to over ride the file on disk
	// check(json.Unmarshal([]byte(testconfigjson), cfg), "json")
}

func teardownTests() {}

func withAppCmd(f func(*testing.T, *bytes.Buffer, *App), c cli.CliCommand) func(*testing.T) {
	return func(t *testing.T) {
		cmd, ok := c.(*App)
		if !ok {
			t.Error("not an *App")
		}
		buf := new(bytes.Buffer)
		cmd.SetOutput(buf)
		f(t, buf, cmd)
	}
}

func withCartCmd(
	b cli.Builder,
	f func(*cartCmd, *bytes.Buffer, *testing.T),
) func(*testing.T) {
	return func(t *testing.T) {
		cart := NewCartCmd(b).(*cartCmd)
		buf := &bytes.Buffer{}
		cart.SetOutput(buf)

		f(cart, buf, t)
	}
}

func check(e error, msg string) {
	if e != nil {
		fmt.Printf("test setup failed: %s - %s\n", e, msg)
		os.Exit(1)
	}
}

func TestExecute(t *testing.T) {
	tests.InitHelpers(t)
	var (
		exp    string
		err    error
		buf    *bytes.Buffer
		errmsg *ErrMsg
	)

	tt := []struct {
		args    []string
		exp     string
		outfunc func() string
		test    func(*testing.T)
		cleanup bool
	}{
		{args: []string{"config", "-f"}, outfunc: func() string { return fmt.Sprintf("%s\n", config.File()) }},
		{args: []string{"--delete-menu", "config", "-d"}, outfunc: func() string { return config.Folder() + "\n" }},
		{args: []string{"--service=Delivery", "config", "-f"}, outfunc: func() string { return config.File() + "\n" }},
		{args: []string{"--log=log.txt", "config", "-d"}, outfunc: func() string { return config.Folder() + "\n" },
			test: func(t *testing.T) {
				logfile := fp.Join(config.Folder(), "logs", "log.txt")
				if _, err = os.Stat(logfile); os.IsNotExist(err) {
					t.Error("file should exist")
				}
				log.Print("hello")
				data, _ := ioutil.ReadFile(logfile)
				if !strings.HasSuffix(strings.Trim(string(data), "\n\t "), "hello") {
					t.Error("logfile should end with the last message")
				}
			}},
		{args: []string{
			"config", "set",
			"address.street='1600 Pennsylvania Ave NW'",
			"address.cityname=Washington",
			"address.state=DC",
			"address.zipcode=20500"},
			exp: "", outfunc: nil,
			test: func(t *testing.T) {
				if config.GetString("address.zipcode") != "20500" {
					t.Error("wrong zip")
				}
				if config.GetString("address.state") != "DC" {
					t.Error("wrong state")
				}
			},
		},
		{args: []string{"cart"}, exp: "No_orders_saved.\n"},
		{args: []string{"cart", "new", "testorder", "-p=12SCREEN"}, exp: ""},
		{args: []string{"cart"}, exp: "Your Orders:\n  testorder\n"},
		// {args: []string{"-L"}, exp: "1300 L St Nw\nWashington, DC 20005\nALL Credit Card orders must have Credit Card and ID present at the Time of Delivery or Pick-up\n\nStore id: 4336\nCoordinates: 38.9036, -77.03\n"},
		{args: []string{"-L"}, exp: "1300 L St Nw\nWashington, DC 20005\nPlease consider tipping your driver for awesome service!!!\n\nStore id: 4336\nCoordinates: 38.9036, -77.03\n"},
		{args: []string{"config", "-d"}, outfunc: func() string { return config.Folder() + "\n" }, cleanup: true},
	}

	for i, tc := range tt {
		buf, err = tests.CaptureOutput(func() {
			errmsg = Execute(tc.args, ".config/apizza/.tests")
		})
		tests.Check(err)
		if errmsg != nil {
			t.Error(errmsg.Msg, errmsg.Err, tc.args)
		}
		if tc.outfunc != nil {
			exp = tc.outfunc()
		} else {
			exp = tc.exp
		}
		if tc.test != nil {
			t.Run(fmt.Sprintf("Exec test: %d", i), tc.test)
		}
		tests.Compare(t, buf.String(), exp)
		config.Save()
		if tc.cleanup {
			os.RemoveAll(config.Folder())
		}
	}
}

func TestYesOrNo(t *testing.T) {
	tests.InitHelpers(t)
	var res bool = false
	f, err := ioutil.TempFile("", "")
	tests.Fatal(err)
	_, err = f.Write([]byte("yes"))
	tests.Fatal(err)
	_, err = f.Seek(0, os.SEEK_SET)
	tests.Fatal(err)
	if yesOrNo(f, "this is a message") {
		res = true
	}
	if !res {
		t.Error("should have been yes")
	}

	tests.Check(f.Close())
	f, err = ioutil.TempFile("", "")
	tests.Check(err)
	_, err = f.Write([]byte("no"))
	tests.Check(err)
	_, err = f.Seek(0, os.SEEK_SET)
	tests.Check(err)
	res = false
	if yesOrNo(f, "msg") {
		res = true
	}
	if res {
		t.Error("should have gotten a no")
	}
	tests.Check(f.Close())
}
