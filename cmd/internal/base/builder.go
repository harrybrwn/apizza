package base

import (
	"io"

	"github.com/harrybrwn/apizza/pkg/cache"
)

// Builder defines an interface for an object that builds commands.
type Builder interface {
	Build(use, short string, r Runner) *Command
	Output() io.Writer
	DBBuilder
	ConfigBuilder
}

// DBBuilder is a cli builder that can give away a database.
type DBBuilder interface {
	DB() *cache.DataBase
}

// ConfigBuilder is a cli builder that can give away a config struct.
type ConfigBuilder interface {
	Config() *Config
}
