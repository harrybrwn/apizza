package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
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
	cli.CliCommand     // this is also the root command
	client.StoreFinder // for app.store()

	db   *cache.DataBase
	conf *cli.Config
	addr *obj.Address
	logf *os.File

	// global apizza options
	gOpts opts.CliFlags

	// root specific opts
	opts opts.ApizzaFlags
}

// NewApp creates a new app for the main cli.
func NewApp(out io.Writer) *App {
	app := &App{
		db:    nil,
		conf:  &cli.Config{},
		gOpts: opts.CliFlags{},
		opts:  opts.ApizzaFlags{},
	}
	app.CliCommand = cli.NewCommand("apizza", "Dominos pizza from the command line.", app.Run)
	app.StoreFinder = client.NewStoreGetterFunc(app.getService, app.Address)
	app.SetOutput(out)
	return app
}

// CreateApp from a pre-created database and config.
func CreateApp(db *cache.DataBase, conf *cli.Config, out io.Writer) *App {
	app := NewApp(out)
	app.db = db
	app.conf = conf
	return app
}

// Init wil setup the app.
func (a *App) Init() error {
	if a.conf == nil {
		a.conf = &cli.Config{}
	}
	a.initflags()
	return errs.Pair(a.SetConfig(".apizza"), a.InitDB())
}

// SetConfig for the the app
func (a *App) SetConfig(dir string) error {
	return config.SetConfig(dir, a.conf)
}

// InitDB for the app.
func (a *App) InitDB() (err error) {
	a.db, err = data.NewDatabase()
	return
}

// DB returns the database
func (a *App) DB() *cache.DataBase {
	return a.db
}

// Build builds commands.
func (a *App) Build(use, short string, r cli.Runner) *cli.Command {
	return cli.NewCommand(use, short, r.Run)
}

// Config returns the config struct.
func (a *App) Config() *cli.Config {
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
	if len(a.gOpts.Service) == 0 {
		return a.conf.Service
	}
	return a.gOpts.Service
}

var _ cli.Builder = (*App)(nil)

// Run the app.
func (a *App) Run(cmd *cobra.Command, args []string) (err error) {
	if a.opts.Openlogs {
		editor := os.Getenv("EDITOR")
		c := exec.Command(editor, filepath.Join(config.Folder(), "logs", "dev.log"))
		c.Stdin = os.Stdin
		c.Stdout = a.Output()
		return c.Run()
	}
	if a.gOpts.ClearCache {
		a.Printf("removing %s\n", a.db.Path())
		return a.db.Destroy()
	}
	if a.opts.StoreLocation {
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

	a.gOpts.Install(persistflags)
	a.opts.Install(flags)

	persistflags.BoolVar(&test, "test", false, "testing flag (for development)")
	persistflags.MarkHidden("test")
}

func (a *App) prerun(*cobra.Command, []string) (err error) {
	if a.gOpts.ResetMenu {
		err = a.DB().Delete("menu")
	}
	var e error
	if a.gOpts.Address != "" {
		parsed, err := dawg.ParseAddress(a.gOpts.Address)
		if err != nil {
			return err
		}
		a.conf.Address = *obj.FromAddress(parsed)
	}

	if a.gOpts.Service != "" {
		if !(a.gOpts.Service == dawg.Delivery || a.gOpts.Service == dawg.Carryout) {
			return dawg.ErrBadService
		}
		a.conf.Service = a.gOpts.Service
	}

	if a.gOpts.LogFile != "" {
		a.logf, e = os.Create(logfile(a.gOpts.LogFile))
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
