package opts

import "github.com/spf13/pflag"

// CliFlags for the root apizza command.
type CliFlags struct {
	Address string
	Service string

	ClearCache bool
	ResetMenu  bool
	LogFile    string
}

// Install the RootFlags
func (rf *CliFlags) Install(persistflags *pflag.FlagSet) {
	persistflags.BoolVar(&rf.ClearCache, "clear-cache", false, "delete the database")
	persistflags.BoolVar(&rf.ResetMenu, "delete-menu", false, "delete the menu stored in cache")
	persistflags.StringVar(&rf.LogFile, "log", "", "set a log file (found in ~/.apizza/logs)")

	persistflags.StringVar(&rf.Address, "address", rf.Address, "use a specific address")
	persistflags.StringVar(&rf.Service, "service", rf.Service, "select a Dominos service, either 'Delivery' or 'Carryout'")
}

// ApizzaFlags that are not persistant.
type ApizzaFlags struct {
	StoreLocation bool

	// developer opts
	Openlogs bool
	Dumpdb   bool
}

// Install the apizza flags
func (af *ApizzaFlags) Install(flags *pflag.FlagSet) {
	flags.BoolVarP(&af.StoreLocation, "store-location", "L", false, "show the location of the nearest store")

	flags.BoolVar(&af.Openlogs, "open-logs", false, "open the log file")
	flags.MarkHidden("open-logs")
	flags.BoolVar(&af.Dumpdb, "dump-db", false, "dump the database to stdout as json")
	flags.MarkHidden("dump-db")
}