package cmd

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/spf13/cobra"
)

// App is the main app, the root command, and the Builder.
type App struct {
	base.CliCommand // this is also the root command
	storefinder     // for app.store()

	db   *cache.DataBase
	conf *base.Config
	addr *obj.Address
	logf *os.File
	log  *log.Logger

	// root command flags
	address    string
	service    string
	storeID    string
	clearCache bool
	resetMenu  bool
	openlogs   bool

	// persistant flags
	storeLocation bool
	logfile       string
}

func newapp(db *cache.DataBase, conf *base.Config, out io.Writer) *App {
	app := &App{
		CliCommand: nil,
		db:         db,
		conf:       conf,
		addr:       &conf.Address,

		// default flag values...
		address:    "",
		service:    "",
		clearCache: false,
	}

	app.CliCommand = base.NewCommand("apizza", "Dominos pizza from the command line.", app.Run)
	app.storefinder = newStoreGetter(
		func() string {
			if len(app.service) == 0 {
				return conf.Service
			}
			return app.service
		},
		app.Address,
	)
	app.SetOutput(out)
	return app
}

// NewApp creates a new app for the main cli.
func NewApp(out io.Writer) (*App, error) {
	app := &App{
		db:   nil,
		conf: &base.Config{},
	}
	if err := app.Init(); err != nil {
		return nil, err
	}
	app.CliCommand = base.NewCommand("apizza", "Dominos pizza from the command line.", app.Run)
	app.storefinder = newStoreGetter(
		func() string {
			if len(app.service) == 0 {
				return app.conf.Service
			}
			return app.service
		},
		app.Address,
	)
	app.SetOutput(out)
	return app, nil
}

// Init wil setup the app.
func (a *App) Init() error {
	var err error
	if a.conf == nil {
		a.conf = &base.Config{}
	}
	if err = config.SetConfig(".apizza", a.conf); err != nil {
		return err
	}

	if a.db, err = data.NewDatabase(); err != nil {
		return err
	}
	return err
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
func (a *App) Config() *base.Config {
	return a.conf
}

// Address returns the address.
func (a *App) Address() dawg.Address {
	if a.addr != nil {
		return a.addr
	}
	return &a.conf.Address
}

// Cleanup cleans everything up.
func (a *App) Cleanup() error {
	return errs.Pair(a.db.Close(), config.Save())
}

// Log to the logging file
func (a *App) Log(v ...interface{}) {
	log.Print(v...)
}

func (a *App) exec() error {
	a.initflags()
	a.Addcmd(
		newCartCmd(a).Addcmd(
			newAddOrderCmd(a),
		),
		newConfigCmd(a).Addcmd(
			newConfigSet(),
			newConfigGet(),
		),
		newMenuCmd(a),
		newOrderCmd(a),
	)
	return a.Cmd().Execute()
}

var _ base.Builder = (*App)(nil)

// Run the app.
func (a *App) Run(cmd *cobra.Command, args []string) (err error) {
	if a.openlogs {
		editor := os.Getenv("EDITOR")
		c := exec.Command(editor, filepath.Join(config.Folder(), "logs", "dev.log"))
		c.Stdin = os.Stdin
		c.Stdout = a.Output()
		return c.Run()
	}
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
	cmd := a.Cmd()
	flags := cmd.Flags()
	persistflags := cmd.PersistentFlags()

	cmd.PersistentPreRunE = a.prerun
	cmd.PostRunE = a.postrun

	flags.BoolVar(&a.clearCache, "clear-cache", false, "delete the database")
	flags.BoolVarP(&a.storeLocation, "store-location", "L", false, "show the location of the nearest store")
	persistflags.BoolVar(&a.resetMenu, "delete-menu", false, "delete the menu stored in cache")

	persistflags.StringVar(&a.logfile, "log", "", "set a log file (found in ~/.apizza/logs)")
	persistflags.StringVar(&a.address, "address", a.address, "use a specific address")
	persistflags.StringVar(&a.service, "service", a.service, "select a Dominos service, either 'Delivery' or 'Carryout'")

	persistflags.BoolVar(&test, "test", false, "testing flag (for development)")
	persistflags.BoolVar(&reset, "reset", false, "reset the program (for development)")
	persistflags.MarkHidden("test")
	persistflags.MarkHidden("reset")

	flags.BoolVar(&a.openlogs, "open-logs", false, "open the log file")
	flags.MarkHidden("open-logs")
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
	return filepath.Join(config.Folder(), "logs", name)
}
