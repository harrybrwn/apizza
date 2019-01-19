package config

import (
	"testing"
	"os"
)

func TestSetConfig(t *testing.T) {
	path := "C:\\Users\\harry\\.testconfig"
	type Cnfg struct {
		Test string `config:"test" default:"\"this is a test config file\""`
		Msg string `config:"msg" default:"\"this should have been deleted, please remove it\""`
		More struct {
	        One string `config:"one"`
	        Two string `config:"two"`
	    } `config:"more"`
	}
	var c = &Cnfg{}
	if Get(c, "msg") != Get(c, "Msg") != c.Msg {
		t.Error("these function calls and variables should be the same")
	}
	cfg.More.One = "hey is this shit workin"
	if Get(c, "more.one")
	config.Set(c, key, val)
	err := SetConfig(".testconfig", c)
	if err != nil { t.Error(err) }



	os.Remove(path)
}
