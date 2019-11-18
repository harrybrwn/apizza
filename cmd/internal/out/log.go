package out

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
)

// DefaultLogFile is the default logging file.
var (
	DefaultLogFile       = "dev.log"
	MaxLogSize     int64 = 4000
	TruncateLines        = 10
)

// Logdir is where the logs go
const Logdir = "logs"

// LogFile returns a log file.
func LogFile() (f *os.File, err error) {
	path := filepath.Join(config.Folder(), Logdir)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		if err = os.Mkdir(path, 0755); err != nil {
			return nil, err
		}
	}
	fname := filepath.Join(path, DefaultLogFile)
	f, err = os.OpenFile(fname, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return f, err
	}

	if stat.Size() > MaxLogSize {
		return reset(f, TruncateLines)
	}

	return f, nil
}

func reset(f *os.File, n int) (*os.File, error) {
	name := f.Name()
	data, err := trimLines(f, n)
	if err != nil {
		return nil, err
	}

	f, err = os.OpenFile(name, os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	return f, errs.EatInt(f.Write(data))
}

func trimLines(f *os.File, n int) ([]byte, error) {
	nlCount := 0
	for nlCount < n {
		if peek(f) == '\n' {
			nlCount++
		}
	}
	b, err := ioutil.ReadAll(f)
	return b, errs.Pair(err, f.Close())
}

func peek(f *os.File) byte {
	b := make([]byte, 1)
	_, err := f.Read(b)
	if err != nil {
		return 0
	}
	return b[0]
}
