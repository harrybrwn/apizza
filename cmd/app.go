package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/cmd/opts"
	"github.com/harrybrwn/apizza/dawg"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/harrybrwn/apizza/pkg/config"
	"github.com/harrybrwn/apizza/pkg/errs"
	"github.com/spf13/cobra"
)

// App is the main app, the root command, and the Builder.
type App struct {
	base.CliCommand    // this is also the root command
	client.StoreFinder // for app.store()

	db   *cache.DataBase
	conf *base.Config
	addr *obj.Address
	logf *os.File
	log  *log.Logger

	opts opts.RootFlags

	// persistant flags
	storeLocation bool
	logfile       string
}

// NewApp creates a new app for the main cli.
func NewApp(out io.Writer) *App {
	app := &App{
		db:   nil,
		conf: &base.Config{},
		opts: opts.RootFlags{},
	}
	app.CliCommand = base.NewCommand("apizza", "Dominos pizza from the command line.", app.Run)
	app.StoreFinder = client.NewStoreGetterFunc(app.getService, app.Address)
	app.SetOutput(out)
	return app
}

// CreateApp from a pre-created database and config.
func CreateApp(db *cache.DataBase, conf *base.Config, out io.Writer) *App {
	app := NewApp(out)
	app.db = db
	app.conf = conf
	return app
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

	a.initflags()
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
func (a *App) Cleanup() (err error) {
	return errs.Pair(a.db.Close(), config.Save())
}

// Log to the logging file
func (a *App) Log(v ...interface{}) {
	log.Print(v...)
}

// Execute the root command.
func (a *App) Execute() error {
	a.initflags()
	return a.Cmd().Execute()
}

func (a *App) getService() string {
	if len(a.opts.Service) == 0 {
		return a.conf.Service
	}
	return a.opts.Service
}

var _ base.Builder = (*App)(nil)

// Run the app.
func (a *App) Run(cmd *cobra.Command, args []string) (err error) {
	if a.opts.Openlogs {
		editor := os.Getenv("EDITOR")
		c := exec.Command(editor, filepath.Join(config.Folder(), "logs", "dev.log"))
		c.Stdin = os.Stdin
		c.Stdout = a.Output()
		return c.Run()
	}
	if a.opts.ClearCache {
		err = a.db.Close()
		a.Printf("removing %s\n", a.db.Path())
		return errs.Pair(err, os.Remove(a.db.Path()))
	}
	if a.storeLocation {
		store := a.Store()
		a.Println(store.Address)
		a.Printf("\n")
		a.Println("Store id:", store.ID)
		a.Printf("Coordinates: %s, %s\n",
			store.StoreCoords["StoreLatitude"],
			store.StoreCoords["StoreLongitude"],
		)
		return nil
	}
	if a.opts.Dumpdb {
		data, err := a.db.Map()
		if err != nil {
			return err
		}
		log.Println("dumping database to stdout")
		fmt.Print("{")
		for k, v := range data {
			fmt.Printf("\"%s\":%v,", k, string(v))
		}
		fmt.Print("}")
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
	a.opts.Install(flags, persistflags)

	flags.BoolVarP(&a.storeLocation, "store-location", "L", false, "show the location of the nearest store")
	persistflags.StringVar(&a.logfile, "log", "", "set a log file (found in ~/.apizza/logs)")

	persistflags.BoolVar(&test, "test", false, "testing flag (for development)")
	persistflags.BoolVar(&reset, "reset", false, "reset the program (for development)")
	persistflags.MarkHidden("test")
	persistflags.MarkHidden("reset")
}

func (a *App) prerun(*cobra.Command, []string) (err error) {
	if a.opts.ResetMenu {
		err = a.DB().Delete("menu")
	}
	var (
		e    error
		file string
	)
	if a.opts.Address != "" {
		parsed, err := dawg.ParseAddress(a.opts.Address)
		if err != nil {
			return err
		}
		a.conf.Address = *obj.FromAddress(parsed)
	}

	if a.opts.Service != "" {
		if !(a.opts.Service == dawg.Delivery || a.opts.Service == dawg.Carryout) {
			return dawg.ErrBadService
		}
		a.conf.Service = a.opts.Service
	}

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
