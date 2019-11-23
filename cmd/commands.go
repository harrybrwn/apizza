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
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/config"
)

var menuUpdateTime = 12 * time.Hour

// Execute runs the root command
func Execute() {
	app := NewApp(os.Stdout)
	err := app.Init()
	if err != nil {
		handle(err, "Internal Error", 1)
	}
	// cfg = app.conf

	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join(config.Folder(), "logs", "dev.log"),
		MaxSize:    25,  // megabytes
		MaxBackups: 10,  // number of spare files
		MaxAge:     365, //days
		Compress:   false,
	})

	defer func() {
		handle(app.Cleanup(), "Internal Error", 1)
	}()

	handle(app.exec(), "Error", 1)
}

func handle(e error, msg string, code int) {
	if e == nil {
		return
	}
	log.Printf("(Failure) %s: %s\n", msg, e)
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, e)
	os.Exit(code)
}

// StoreFinder is a mixin that allows for efficient caching and retrival of
// store structs.
type StoreFinder interface {
	Store() *dawg.Store
}

// storegetter is meant to be a mixin for any struct that needs to be able to
// get a store.
type storegetter struct {
	getaddr   func() dawg.Address
	getmethod func() string
	dstore    *dawg.Store
}

// NewStoreGetter will create a new storefinder.
func NewStoreGetter(builder base.Builder) StoreFinder {
	return &storegetter{
		getmethod: func() string {
			return builder.Config().Service
		},
		getaddr: builder.Address,
		dstore:  nil,
	}
}

func newStoreGetter(service func() string, addr func() dawg.Address) StoreFinder {
	return &storegetter{
		getmethod: service,
		getaddr:   addr,
		dstore:    nil,
	}
}

func (s *storegetter) Store() *dawg.Store {
	if s.dstore == nil {
		var err error
		var address = s.getaddr()
		s.dstore, err = dawg.NearestStore(address, s.getmethod())
		if err != nil {
			handle(err, "Store Find Error", 1) // will exit
		}
	}
	return s.dstore
}
