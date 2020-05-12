package errs

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

// Eat will throw away the first value and return the error given.
//
// 	err := Eat(w.Write([]byte("hello?")))
func Eat(v interface{}, e error) error {
	return e
}

// EatInt will eat an int and return the error. This is good if you want
// to chain calls to an io.Writer or io.Reader.
func EatInt(n int, e error) error {
	return e
}

// PrintStack will print out the current stack trace.
func PrintStack() {
	stackFrame(os.Stderr, 3)
}

// credit: https://www.komu.engineer/blogs/golang-stacktrace/golang-stacktrace
func stackFrame(w io.Writer, depth int) {
	buf := make([]uintptr, 100)
	n := runtime.Callers(depth, buf[:])
	stack := buf[:n]
	frames := runtime.CallersFrames(stack)

	for i := 0; i < n; i++ {
		frame, _ := frames.Next()
		fmt.Fprintf(w, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
	}
}
