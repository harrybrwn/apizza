// Copyright © 2019 Harrison Brown harrybrown98@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/harrybrwn/apizza/pkg/config"
)

var (
	menuUpdateTime = 12 * time.Hour

	// Logger for the cmd package
	Logger = &lumberjack.Logger{
		Filename:   "",
		MaxSize:    25,  // megabytes
		MaxBackups: 10,  // number of spare files
		MaxAge:     365, //days
		Compress:   false,
	}
)

const enableLog = true

// Execute runs the root command
func Execute() {
	app := NewApp(os.Stdout)
	err := app.Init()
	if err != nil {
		handle(err, "Internal Error", 1)
	}

	if enableLog {
		Logger.Filename = filepath.Join(config.Folder(), "logs", "dev.log")
		log.SetOutput(Logger)
	}

	defer func() {
		handle(app.Cleanup(), "Internal Error", 1)
	}()

	cmd := app.Cmd()
	cmd.AddCommand(
		newCartCmd(app).Addcmd(
			newAddOrderCmd(app),
		).Cmd(),
		newConfigCmd(app).Addcmd(
			newConfigSet(),
			newConfigGet(),
		).Cmd(),
		newMenuCmd(app).Cmd(),
		newOrderCmd(app).Cmd(),
	)

	handle(cmd.Execute(), "Error", 1)
}

func handle(e error, msg string, code int) {
	if e == nil {
		return
	}
	w := io.MultiWriter(Logger, os.Stderr)
	fmt.Fprintf(w, "%s: %s\n", msg, e)
	os.Exit(code)
}
