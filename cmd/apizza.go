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
	fp "path/filepath"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/command"
	"github.com/harrybrwn/apizza/pkg/config"
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
		NewAddAddressCmd(builder, os.Stdin).Cmd(),
	}
}

// ErrMsg is not actually an error but it is my way of
// containing an error with a message and an exit code.
type ErrMsg struct {
	Msg  string
	Code int
	Err  error
}

func senderr(e error, msg string, code int) *ErrMsg {
	if e == nil {
		return nil
	}
	return &ErrMsg{Msg: msg, Code: code, Err: e}
}

// Execute runs the root command
func Execute(args []string, dir string) (msg *ErrMsg) {
	app := NewApp(os.Stdout)
	err := app.Init(dir)
	if err != nil {
		return senderr(err, "Internal Error", 1)
	}

	if enableLog {
		Logger.Filename = fp.Join(config.Folder(), "logs", "dev.log")
		log.SetOutput(Logger)
	}

	defer func() {
		errmsg := senderr(app.Cleanup(), "Internal Error", 1)
		if errmsg != nil {
			// if we always set it the the return value will always
			// be the same as errmsg
			msg = errmsg
		}
	}()

	cmd := app.Cmd()
	cmd.SetArgs(args)
	cmd.AddCommand(AllCommands(app)...)
	return senderr(cmd.Execute(), "Error", 1)
}

var test = false

func yesOrNo(in *os.File, msg string) bool {
	var res string
	fmt.Printf("%s ", msg)
	_, err := fmt.Fscan(in, &res)
	if err != nil {
		return false
	}

	switch strings.ToLower(res) {
	case "y", "yes", "si":
		return true
	}
	return false
}
