package opts

import "github.com/spf13/pflag"

// RootFlags for the root apizza command.
type RootFlags struct {
	Address string
	Service string

	ClearCache bool
	ResetMenu  bool

	// developer opts
	Openlogs bool
	Dumpdb   bool
}

// Install the RootFlags
func (rf *RootFlags) Install(flags *pflag.FlagSet, persistflags *pflag.FlagSet) {
	flags.BoolVar(&rf.ClearCache, "clear-cache", false, "delete the database")
	persistflags.BoolVar(&rf.ResetMenu, "delete-menu", false, "delete the menu stored in cache")

	persistflags.StringVar(&rf.Address, "address", rf.Address, "use a specific address")
	persistflags.StringVar(&rf.Service, "service", rf.Service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	flags.BoolVar(&rf.Openlogs, "open-logs", false, "open the log file")
	flags.MarkHidden("open-logs")
	flags.BoolVar(&rf.Dumpdb, "dump-db", rf.Dumpdb, "dump the database to stdout as json")
	flags.MarkHidden("dump-db")
}
