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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/command"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
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

// AllCommands returns a list of all the Commands.
func AllCommands(builder cli.Builder) []*cobra.Command {
	return []*cobra.Command{
		NewCartCmd(builder).Cmd(),
		command.NewConfigCmd(builder).Cmd(),
		NewMenuCmd(builder).Cmd(),
		NewOrderCmd(builder).Cmd(),
	}
}

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
	cmd.AddCommand(AllCommands(app)...)
	errs.Handle(cmd.Execute(), "Error", 1)
}

var test = false
var reset = false

func yesOrNo(msg string) bool {
	var in string
	fmt.Printf("%s ", msg)
	fmt.Scan(&in)

	switch strings.ToLower(in) {
	case "y", "yes", "si":
		return true
	}
	return false
}
