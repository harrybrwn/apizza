package dawg_test

import (
	"testing"

	"github.com/harrybrwn/apizza/dawg"
)

func TestURLParams(t *testing.T) {
	expected := []string{
		"a=what&b=7&c=false",
		"b=7&c=false&a=what",
		"c=false&a=what&b=7",
	}
	p := dawg.Params{"a": "what", "b": 7, "c": false}
	enc := p.Encode()
	func() {
		for _, expec := range expected {
			if enc == expec {
				return
			}
		}
		t.Error("bad url encoding")
	}()
	p = dawg.Params{"byteobj": []byte("data")}
	if p.Encode() != "byteobj=data" {
		t.Error("bad encoding for bytes")
	}
	t.Run("bad Param type", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic")
			}
		}()
		type test struct{ a string }
		p = dawg.Params{"struct": test{"no"}}
		p.Encode()
	})
	p = nil
	enc = p.Encode()
	if enc != "" {
		t.Error("should be empty string")
	}
}
