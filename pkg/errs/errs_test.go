package errs

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestBasicError(t *testing.T) {
	e := New("this is an error")
	if e.Error() != "this is an error" {
		t.Error("bad error message from basic error")
	}
	Handle(nil, "should be nil", 1)
}

func TestLinearErrorsPair(t *testing.T) {
	e := Pair(nil, nil)
	if e != nil {
		t.Error("expected nil error from a pair of nil errors")
	}

	terr := errors.New("this is a test error")
	e = Pair(nil, terr)
	if e != terr {
		t.Error("Pair should return an arg if the other is nil: arg1 failed")
	}
	terr2 := errors.New("second test error")
	e = Pair(terr2, nil)
	if e != terr2 {
		t.Error("Pair should return an arg if the other is nil: arg2 failed")
	}

	erra := errors.New("error a")
	errb := errors.New("error b")
	tt := []struct{ a, b error }{
		{nil, nil},
		{nil, errb},
		{erra, nil},
	}
	for _, tc := range tt {
		p := Pair(tc.a, tc.b)
		if _, ok := p.(*linearError); ok {
			t.Error("theses test cases should not have returned a linear error")
		}
		if tc.a == nil && p != tc.b {
			t.Error("Pair(nil, err) should return err")
		}
		if tc.b == nil && p != tc.a {
			t.Error("Pair(err, nil) should return err")
		}
	}
	if erra == nil || errb == nil {
		t.Fatal("neither should not be nil")
	}

	p := Pair(erra, errb)
	if p == nil {
		t.Error("Pair(err, err) should not return nil")
	}
	pair, ok := p.(*linearError)
	if !ok {
		t.Error("Pair(err, err) should return a linearError")
	}
	if pair.errList[0] != erra {
		t.Error("pair did not save first error")
	}
	if pair.errList[1] != errb {
		t.Error("pari did not save second error")
	}
	emsg := pair.Error()
	exp := "Errors:\n  1. error a\n  2. error b\n"
	if emsg != exp {
		t.Errorf("wrong error msg; got: %s, want: %s", emsg, exp)
	}

	p = Pair(pair, New("last"))
	p = Pair(New("first"), p)
	if pair, ok = p.(*linearError); !ok {
		t.Error("should be linear")
	}
	if len(pair.errList) != 2 {
		t.Error("wrong number of errors; got:", len(pair.errList))
	}
	exp = `Errors:
  1. first
  2. error a
  3. error b
  4. last
`
	if p.Error() != exp {
		t.Error("wrong output:\n", p.Error())
	}

	inner := Pair(Pair(New("b"), New("c")), New("d"))
	exp = "Errors:\n  1. b\n  2. c\n  3. d\n"
	if inner.Error() != exp {
		t.Error("bad output:\n", inner.Error())
	}

	p = Pair(New("a"), inner)
	exp = "Errors:\n  1. a\n  2. b\n  3. c\n  4. d\n"
	if p.Error() != exp {
		t.Error("bad output:\n", p.Error())
	}
}

func TestLinearError(t *testing.T) {
	e := Append(New("err 1"))
	if e == nil {
		t.Error("should not return nil")
	}
	if _, ok := e.(*linearError); ok {
		t.Error("Append on a single error should return ")
	}
	if e.Error() != "err 1" {
		t.Error("Append on a single error should return that one error")
	}

	e = Append(e, New("err 2"), New("err 3"))
	if _, ok := e.(*linearError); !ok {
		t.Error("should be a linearError")
	}
	msg := e.Error()
	exp := "Errors:\n  1. err 1\n  2. err 2\n  3. err 3\n"
	if msg != exp {
		t.Errorf("wrong error message; got:\n%s,\nwant:\n%s", msg, exp)
	}
	e = Append(e, New("batch 2"), New("batch 2"))
	if e == nil {
		t.Error("should not be nil")
	}
	msg = e.Error()
	exp += "  4. batch 2\n  5. batch 2\n"
	if msg != exp {
		t.Errorf("wrong error message;\ngot:\n%s\nwant:\n%s", msg, exp)
	}

	e = Pair(e, Pair(New("pair1"), New("pair2")))
	if _, ok := e.(*linearError); !ok {
		t.Error("should be a linearError")
	}

	p := Pair(Pair(New("one"), New("two")), Pair(New("three"), New("four")))
	exp = "Errors:\n  1. one\n  2. two\n  3. three\n  4. four\n"
	if p.Error() != exp {
		t.Errorf("wrong error message\ngot:\n%s\nwant:\n%s", p.Error(), exp)
	}

	p = Append(nil, New("one"), New("two"))
	exp = "Errors:\n  1. one\n  2. two\n"
	if p.Error() != exp {
		t.Error("bad output:\n", p.Error())
	}
}

func TestBadAppendandPair(t *testing.T) {
	e := Append(nil, nil, nil)
	if e != nil {
		t.Error("an error list with all nil should return nil")
		fmt.Println(e)
	}
	exp := "Errors:\n  1. 1\n  2. 2\n  3. 3\n"

	e = Append(New("1"), Pair(New("2"), New("3")))
	if e.Error() != exp {
		t.Error("wrong output")
	}

	e = Append(nil, New("1"), nil, New("2"), New("3"), nil)
	if e.Error() != exp {
		t.Errorf("wrong output; got: %s\nwant: %s", e.Error(), exp)
	}
}

func TestEating(t *testing.T) {
	f := func() (int, error) { return 3, New("eats") }
	e := EatInt(f())
	if e.Error() != "eats" {
		t.Error("EatInt spat out the wrong error")
	}
}

func TestPrintStack(t *testing.T) {
	buf := new(bytes.Buffer)
	func() {
		func() {
			stackFrame(buf, 0)
		}()
	}()
	lines := strings.Split(buf.String(), "\n")
	if len(lines) < 8 {
		t.Error("too few stack frames")
	}
}

func BenchmarkAppend(b *testing.B) {
	var (
		i int
		e error
	)
	b.Run("linearError on left", func(b *testing.B) {
		var list = make([]error, b.N)
		for i = 0; i < b.N; i++ {
			list[i] = New(fmt.Sprintf("error %d", i))
		}
		b.ResetTimer()
		for i = 0; i < b.N; i++ {
			e = Append(e, list[i]) // linearError on the left
		}
	})
	b.Run("linearError on right", func(b *testing.B) {
		var list = make([]error, b.N)
		for i = 0; i < b.N; i++ {
			list[i] = New(fmt.Sprintf("error %d", i))
		}
		b.ResetTimer()
		for i = 0; i < b.N; i++ {
			e = Append(list[i], e) // linearError on the right
		}
	})
}

func BenchmarkPair(b *testing.B) {
	b.Run("both errors", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := Pair(New("one"), New("two")); err == nil {
				b.Error("expected error")
			}
		}
	})
	b.Run("both nil", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := Pair(nil, nil); err != nil {
				b.Error("expected nil")
			}
		}
	})
}
