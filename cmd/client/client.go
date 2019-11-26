package client

import (
	"time"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/data"
)

// Client defines an interface which interacts with the dominos api.
type Client interface {
	StoreFinder
	data.MenuCacher
}

// FromBuilder creates a dominos Client from a cli Builder
func FromBuilder(b base.Builder, menuDecay time.Duration) Client {
	finder := NewStoreGetter(b)
	return &client{
		StoreFinder: finder,
		MenuCacher:  data.NewMenuCacher(menuDecay, b.DB(), finder.Store),
	}
}

type client struct {
	StoreFinder
	data.MenuCacher
}
