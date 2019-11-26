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
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/harrybrwn/apizza/cmd/command"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
)

// Logger for the cmd package
var Logger = &lumberjack.Logger{
	Filename:   "",
	MaxSize:    25,  // megabytes
	MaxBackups: 10,  // number of spare files
	MaxAge:     365, //days
	Compress:   false,
}

const enableLog = true

// Execute runs the root command
func Execute() {
	app := NewApp(os.Stdout)
	err := app.Init()
	if err != nil {
		errs.Handle(err, "Internal Error", 1)
	}

	if enableLog {
		Logger.Filename = filepath.Join(config.Folder(), "logs", "dev.log")
		log.SetOutput(Logger)
	}

	defer func() {
		errs.Handle(app.Cleanup(), "Internal Error", 1)
	}()

	cmd := app.Cmd()
	cmd.AddCommand(
		newCartCmd(app).Cmd(),
		command.NewConfigCmd(app).Cmd(),
		NewMenuCmd(app).Cmd(),
		newOrderCmd(app).Cmd(),
	)

	errs.Handle(cmd.Execute(), "Error", 1)
}
