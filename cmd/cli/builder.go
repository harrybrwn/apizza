package cli

import (
	"io"

	"github.com/harrybrwn/apizza/cmd/opts"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
)

// Builder defines an interface for an object that builds commands.
type Builder interface {
	CommandBuilder
	DBBuilder
	StateBuilder
	AddressBuilder
	Output() io.Writer
}

// CommandBuilder defines an interface for building commands.
type CommandBuilder interface {
	Build(use, short string, r Runner) *Command
}

// DBBuilder is a cli builder that can give away a database.
type DBBuilder interface {
	DB() *cache.DataBase
}

// ConfigBuilder is a cli builder that can give away a config struct.
type ConfigBuilder interface {
	Config() *Config
}

// AddressBuilder is a builder interface that should be able to get an
// address.
type AddressBuilder interface {
	Address() dawg.Address
}

// StateBuilder defines a cli builder that has control over the
// program state, whether that is from the config file or the global
// command line options.
type StateBuilder interface {
	ConfigBuilder
	GlobalOptions() *opts.CliFlags
}

// AddrDBBuilder is an anddress-builder and a db-builder.
type AddrDBBuilder interface {
	CommandBuilder
	AddressBuilder
	DBBuilder
}
