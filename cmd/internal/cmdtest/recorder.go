package cmdtest

import (
	"bytes"
	"io"
	"math/rand"
	"strings"
	"testing"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"
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

// TODO:
//   - give the inner config an actual temp file and delete it in
//     the CleanUp function. (need to get rid of global cfg var first)

var services = []string{dawg.Carryout, dawg.Delivery}

// NewRecorder create a new command recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		DataBase: TempDB(),
		Out:      new(bytes.Buffer),
		Conf: &base.Config{
			Name:    "Apizza TestRecorder",
			Service: services[rand.Intn(2)],
			Address: *TestAddress(),
		},
	}
}

// DB will return the internal database.
func (r *Recorder) DB() *cache.DataBase {
	return r.DataBase
}

// Config will return the config struct.
func (r *Recorder) Config() config.Config {
	return r.Conf
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

// ToApp returns the arguments needed to create a cmd.App.
func (r *Recorder) ToApp() (*cache.DataBase, *base.Config, io.Writer) {
	return r.DB(), r.Conf.(*base.Config), r.Output()
}

// CleanUp will cleanup all the the Recorder tempfiles and free all resources.
func (r *Recorder) CleanUp() {
	if err := r.DataBase.Destroy(); err != nil {
		panic(err)
	}
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

// Compare the recorder output with a string
func (r *Recorder) Compare(t *testing.T, expected string) {
	tests.Compare(t, r.Out.String(), expected)
}
