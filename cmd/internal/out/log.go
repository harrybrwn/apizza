package out

import (
	"os"
	"path/filepath"

	"github.com/harrybrwn/apizza/pkg/config"
)

// DefaultLogFile is the default logging file.
var DefaultLogFile = "dev.log"

// Logdir is where the logs go
const Logdir = "logs"

// LogFile returns a log file.
func LogFile() (*os.File, error) {
	path := filepath.Join(config.Folder(), Logdir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.Mkdir(path, 0755); err != nil {
			return nil, err
		}
	}
	fname := filepath.Join(path, DefaultLogFile)
	return os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0666)
}
