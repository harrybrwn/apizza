package data

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/errs"
)

// MenuCacher defines an interface that retrieves, caches, and stores
// menu timestamps.
type MenuCacher interface {
	cache.Updater
	Menu() *dawg.Menu
}

// NewMenuCacher creates a new MenuCacher.
func NewMenuCacher(
	decay time.Duration,
	db cache.Storage,
	store func() *dawg.Store,
) MenuCacher {
	// use gob to cache the menu in binary format
	return NewGobMenuCacher(decay, db, store)
}

func init() {
	gob.Register([]interface{}{})
}

// Encoder is an interface that defines objects that
// are able to Encode and interface.
type Encoder interface {
	Encode(interface{}) error
}

// Decoder is an inteface that defines objects
// that can decode and inteface.
type Decoder interface {
	Decode(interface{}) error
}

type generalMenuCacher struct {
	cache.Updater
	m        *dawg.Menu
	db       cache.Storage
	getstore func() *dawg.Store

	newEncoder func(io.Writer) Encoder
	newDecoder func(io.Reader) Decoder
}

// NewJSONMenuCacher will create a new MenuCacher that stores the
// menu as json.
func NewJSONMenuCacher(
	decay time.Duration,
	db cache.Storage,
	store func() *dawg.Store,
) MenuCacher {
	mc := &generalMenuCacher{
		m:          nil,
		db:         db,
		getstore:   store,
		newEncoder: func(w io.Writer) Encoder { return json.NewEncoder(w) },
		newDecoder: func(r io.Reader) Decoder { return json.NewDecoder(r) },
	}
	mc.Updater = cache.NewUpdater(decay, mc.cacheNewMenu, mc.getCachedMenu)
	return mc
}

// NewGobMenuCacher will create a MenuCacher that will store the menu
// in a binary format using the "encoding/gob" package.
func NewGobMenuCacher(
	decay time.Duration,
	db cache.Storage,
	store func() *dawg.Store,
) MenuCacher {
	mc := &generalMenuCacher{
		m:          nil,
		db:         db,
		getstore:   store,
		newEncoder: func(w io.Writer) Encoder { return gob.NewEncoder(w) },
		newDecoder: func(r io.Reader) Decoder { return gob.NewDecoder(r) },
	}
	mc.Updater = cache.NewUpdater(decay, mc.cacheNewMenu, mc.getCachedMenu)
	return mc
}

func (mc *generalMenuCacher) Menu() *dawg.Menu {
	if mc.m != nil {
		return mc.m
	}
	return nil
}

func (mc *generalMenuCacher) cacheNewMenu() error {
	var e1, e2 error
	mc.m, e1 = mc.getstore().Menu()
	log.Println("caching another menu")

	buf := &bytes.Buffer{}
	e2 = mc.newEncoder(buf).Encode(mc.m)
	return errs.Append(e1, e2, mc.db.Put("menu", buf.Bytes()))
}

func (mc *generalMenuCacher) getCachedMenu() error {
	if mc.m == nil {
		mc.m = new(dawg.Menu)
		raw, err := mc.db.Get("menu")
		if raw == nil {
			return mc.cacheNewMenu()
		}

		dec := mc.newDecoder(bytes.NewBuffer(raw))
		err = errs.Pair(err, dec.Decode(mc.m))
		if err != nil {
			return err
		}

		if mc.m.ID != mc.getstore().ID {
			return mc.cacheNewMenu()
		}
	}
	return nil
}

var _ MenuCacher = (*generalMenuCacher)(nil)
