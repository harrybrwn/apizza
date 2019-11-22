package data

import (
	"encoding/json"
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
	mc := &menuCache{
		m:        nil,
		db:       db,
		getstore: store,
	}
	mc.Updater = cache.NewUpdater(decay, mc.getCachedMenu, mc.cacheNewMenu)
	return mc
}

type menuCache struct {
	cache.Updater
	m        *dawg.Menu
	db       cache.Storage
	getstore func() *dawg.Store
}

func (mc *menuCache) Menu() *dawg.Menu {
	if mc.m != nil {
		return mc.m
	}
	return nil
}

func (mc *menuCache) cacheNewMenu() error {
	var e1, e2 error
	var raw []byte
	mc.m, e1 = mc.getstore().Menu()
	log.Println("caching another menu")
	raw, e2 = json.Marshal(mc.m)
	return errs.Append(e1, e2, mc.db.Put("menu", raw))
}

func (mc *menuCache) getCachedMenu() error {
	if mc.m == nil {
		mc.m = new(dawg.Menu)
		raw, err := mc.db.Get("menu")
		if raw == nil {
			return mc.cacheNewMenu()
		}
		err = errs.Pair(err, json.Unmarshal(raw, mc.m))
		if err != nil {
			return err
		}
		if mc.m.ID != mc.getstore().ID {
			return mc.cacheNewMenu()
		}
	}
	return nil
}

var _ MenuCacher = (*menuCache)(nil)
