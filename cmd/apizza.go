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
	"strings"
)

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

// func newApizzaCmd(app *App) *cobra.Command {
// 	return base.NewCommand(
// 		"apizza", "Dominos pizza from the command line.", app.Run,
// 	).Cmd()
// }

// type rootopts struct {
// 	address string
// 	service string

// 	clearCache bool
// 	resetMenu  bool

// 	// developer opts
// 	openlogs bool
// 	dumpdb   bool
// }

// func (opts *rootopts) install(flags *pflag.FlagSet, persistflags *pflag.FlagSet) {
// 	flags.BoolVar(&opts.clearCache, "clear-cache", false, "delete the database")
// 	persistflags.BoolVar(&opts.resetMenu, "delete-menu", false, "delete the menu stored in cache")

// 	persistflags.StringVar(&opts.address, "address", opts.address, "use a specific address")
// 	persistflags.StringVar(&opts.service, "service", opts.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

// 	persistflags.BoolVar(&test, "test", false, "testing flag (for development)")
// 	persistflags.BoolVar(&reset, "reset", false, "reset the program (for development)")
// 	persistflags.MarkHidden("test")
// 	persistflags.MarkHidden("reset")

// 	flags.BoolVar(&opts.openlogs, "open-logs", false, "open the log file")
// 	flags.MarkHidden("open-logs")
// 	flags.BoolVar(&opts.dumpdb, "dump-db", opts.dumpdb, "dump the database to stdout as json")
// 	flags.MarkHidden("dump-db")
// }
