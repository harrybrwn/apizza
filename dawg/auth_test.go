package dawg

import (
	"testing"
)

func TestBadCreds(t *testing.T) {
	tok, err := gettoken("no", "and no")
	if err == nil {
		t.Error("expected an error")
	}
	if tok != nil {
		t.Error("expected nil token")
	}

	tok, err = gettoken("", "")
	if err == nil {
		t.Error("expected an error")
	}
	if tok != nil {
		t.Error("expected nil token")
	}

	tok, err = gettoken("5uup;hrg];ht8bijer$u9tot", "hurieahgr9[0249eingurivja")
	if err == nil {
		t.Error("wow i accidentlly cracked someone's password")
	}
}
