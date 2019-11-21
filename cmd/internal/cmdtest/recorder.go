package cmdtest

import (
	"bytes"
	"io"
	"strings"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/harrybrwn/apizza/pkg/tests"
)

// Recorder is a mock command builder.
type Recorder struct {
	DataBase *cache.DataBase
	Conf     config.Config
	Out      *bytes.Buffer
}

// NewRecorder create a new command recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		DataBase: must(cache.GetDB(tests.NamedTempFile("test", "apizza_test.db"))),
		Out:      new(bytes.Buffer),
		Conf:     new(base.Config),
	}
}

// DB will return the internal database.
func (r *Recorder) DB() *cache.DataBase {
	return r.DataBase
}

// Config will return the config struct.
func (r *Recorder) Config() config.Config {
	return nil
}

// Output returns the reqorder's output.
func (r *Recorder) Output() io.Writer {
	return r.Out
}

// Build a command.
func (r *Recorder) Build(use, short string, run base.Runner) *base.Command {
	c := base.NewCommand(use, short, run.Run)
	c.SetOutput(r.Output())
	return c
}

var _ base.Builder = (*Recorder)(nil)

func must(db *cache.DataBase, e error) *cache.DataBase {
	if e != nil {
		panic(e)
	}
	return db
}

// Clear will clear all data stored by the recorder. This includes reseting
// the output buffer, opening a fresh database, and resetting the config.
func (r *Recorder) Clear() (err error) {
	r.ClearBuf()
	return r.FreshDB()
}

// ClearBuf will reset the internal output buffer.
func (r *Recorder) ClearBuf() {
	r.Out.Reset()
}

// FreshDB will close the old database, delete it, and open a fresh one.
func (r *Recorder) FreshDB() error {
	var err2 error
	err1 := r.DataBase.Destroy()
	r.DataBase, err2 = cache.GetDB(tests.NamedTempFile("test", "apizza_test.db"))
	return errs.Pair(err1, err2)
}

// Contains will return true if s is contained within the output buffer
// of the Recorder.
func (r *Recorder) Contains(s string) bool {
	return strings.Contains(r.Out.String(), s)
}

// StrEq compares a string with the recorder output buffer.
func (r *Recorder) StrEq(s string) bool {
	return r.Out.String() == s
}
