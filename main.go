// Copyright © 2020 Harrison Brown harrybrown98@gmail.com
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

package main

import (
	"os"
	fp "path/filepath"

	"github.com/harrybrwn/apizza/cmd"
	"github.com/harrybrwn/apizza/pkg/errs"
)

var xdgConfigHome = os.Getenv("XDG_CONFIG_HOME")

func main() {
	configDir := os.Getenv("APIZZA_CONFIG")
	if configDir == "" {
		if xdgConfigHome != "" {
			configDir = fp.Join(xdgConfigHome, "apizza")
		} else {
			configDir = ".config/apizza"
		}
	}

	err := cmd.Execute(os.Args[1:], configDir)
	if err != nil {
		errs.StopNow(err.Err, err.Msg, err.Code)
	}
}
