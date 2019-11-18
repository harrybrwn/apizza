package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/cmd/internal/out"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/spf13/cobra"
)

// App is the main app, the root command, and the Builder.
type App struct {
	base.CliCommand
	storefinder // for app.store()

	db   *cache.DataBase
	conf *Config
	addr *obj.Address
	logf *os.File
	log  *log.Logger

	// temporary compatability stuff
	builder *cliBuilder

	// util flags
	logfile string

	// root command flags
	address    string
	service    string
	storeID    string
	clearCache bool
	resetMenu  bool

	// persistant flags
	storeLocation bool
}

func newapp(db *cache.DataBase, conf *Config, logs io.Writer) *App {
	app := &App{
		CliCommand: nil,
		db:         db,
		conf:       conf,
		addr:       &conf.Address,
		log:        log.New(logs, "", 0),

		// default flag values...
		address:    "",
		service:    "",
		clearCache: false,
	}

	app.CliCommand = base.NewCommand("apizza", "Dominos pizza from the command line.", app.Run)
	app.SetOutput(os.Stdout)
	app.storefinder = newStoreGetter(
		func() string {
			if len(app.service) == 0 {
				return conf.Service
			}
			return app.service
		},
		func() dawg.Address {
			if app.addr == nil {
				return &conf.Address
			}
			return app.addr
		},
	)

	app.builder = &cliBuilder{
		db:   db,
		addr: &conf.Address,
		root: app,
	}
	return app
}

// DB returns the database
func (a *App) DB() *cache.DataBase {
	return a.db
}

// Build builds commands.
func (a *App) Build(use, short string, r base.Runner) *base.Command {
	return base.NewCommand(use, short, r.Run)
}

// Config returns the config struct.
func (a *App) Config() config.Config {
	return a.conf
}

// Log to the logging file
func (a *App) Log(v ...interface{}) {
	a.log.Print(v...)
}

func (a *App) exec() error {
	a.initflags()
	a.Addcmd(
		newCartCmd(a.builder).Addcmd(
			newAddOrderCmd(a.builder),
		),
		newConfigCmd().Addcmd(
			newConfigSet(),
			newConfigGet(),
		),
		a.builder.newMenuCmd(),
		newOrderCmd(),
		newDumpCmd(a),
	)
	return a.Cmd().Execute()
}

var _ base.Builder = (*App)(nil)

// Run the app.
func (a *App) Run(cmd *cobra.Command, args []string) (err error) {
	if a.clearCache {
		err = a.db.Close()
		a.Printf("removing %s\n", a.db.Path())
		return errs.Pair(err, os.Remove(a.db.Path()))
	}
	if a.storeLocation {
		a.Println(a.store().Address)
		a.Printf("\n")
		a.Println("Store id:", a.store().ID)
		a.Printf("Coordinates: %s, %s\n",
			a.store().StoreCoords["StoreLatitude"],
			a.store().StoreCoords["StoreLongitude"],
		)
		return nil
	}
	return cmd.Usage()
}

func (a *App) initflags() {
	a.Cmd().PersistentPreRunE = a.prerun
	a.Cmd().PostRunE = a.postrun
	a.Cmd().PersistentFlags().StringVar(&a.logfile, "log", "", "set a log file")

	a.Flags().BoolVar(&a.clearCache, "clear-cache", false, "delete the database")
	a.Cmd().PersistentFlags().BoolVar(&a.resetMenu, "delete-menu", false, "delete the menu stored in cache")

	a.Cmd().PersistentFlags().StringVar(&a.address, "address", a.address, "use a specific address")
	a.Cmd().PersistentFlags().StringVar(&a.service, "service", a.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	a.Cmd().PersistentFlags().BoolVar(&test, "test", false, "testing flag (for development)")
	a.Cmd().PersistentFlags().BoolVar(&reset, "reset", false, "reset the program (for development)")
	a.Cmd().PersistentFlags().MarkHidden("test")
	a.Cmd().PersistentFlags().MarkHidden("reset")

	a.Flags().BoolVarP(&a.storeLocation, "store-location", "L", false, "show the location of the nearest store")
}

func (a *App) prerun(*cobra.Command, []string) (err error) {
	if a.resetMenu {
		err = a.DB().Delete("menu")
	}
	var (
		e    error
		file string
	)

	if a.logfile != "" {
		file = a.logfile
		a.logf, e = os.Create(logfile(file))
		a.log = log.New(a.logf, "", 0)
		log.SetOutput(a.logf)
	}
	return errs.Pair(err, e)
}

func (a *App) postrun(*cobra.Command, []string) (err error) {
	if a.logf != nil {
		return a.logf.Close()
	}
	return nil
}

func logfile(name string) string {
	return filepath.Join(config.Folder(), out.Logdir, name)
}
