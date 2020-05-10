package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	fp "path/filepath"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/client"
	"github.com/harrybrwn/apizza/cmd/internal"
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

	// root specific options
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
	cmd := app.Cmd()
	cmd.PersistentPreRunE = app.prerun
	cmd.PostRunE = app.postrun
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
func (a *App) Init(dir string) error {
	if a.conf == nil {
		a.conf = &cli.Config{}
	}
	a.initflags()
	return errs.Pair(a.SetConfig(dir), a.InitDB())
}

// SetConfig for the the app
func (a *App) SetConfig(dir string) error {
	return config.SetConfig(dir, a.conf)
}

// InitDB for the app.
func (a *App) InitDB() (err error) {
	a.db, err = data.OpenDatabase()
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
	if a.conf.DefaultAddressName != "" {
		addr, err := a.getDBAddress(a.conf.DefaultAddressName)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Warning: could not find an address named '%s'\n",
				a.conf.DefaultAddressName)
			if obj.AddrIsEmpty(&a.conf.Address) {
				errs.StopNow(internal.ErrNoAddress, "Error", 1)
			}
			return &a.conf.Address
		}
		a.addr = addr
		return addr
	}
	return &a.conf.Address
}

// GlobalOptions returns the variables for the app's global flags
func (a *App) GlobalOptions() *opts.CliFlags {
	return &a.gOpts
}

// Cleanup cleans everything up.
func (a *App) Cleanup() (err error) {
	return errs.Pair(a.db.Close(), config.Save())
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
		c := exec.Command(editor, fp.Join(config.Folder(), "logs", "dev.log"))
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
		a.Println("\nStore id:", store.ID)
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
		// First look in the database as if the flag was a named address.
		// Else check if the flag is a parsable address.
		if a.db.WithBucket("addresses").Exists(a.gOpts.Address) {
			newaddr, err := a.getDBAddress(a.gOpts.Address)
			if err != nil {
				return err
			}
			a.addr = newaddr
		} else {
			parsed, err := dawg.ParseAddress(a.gOpts.Address)
			if err != nil {
				return err
			}
			a.addr = obj.FromAddress(parsed)
		}
	}

	if a.gOpts.Service != "" {
		if !(a.gOpts.Service == dawg.Delivery || a.gOpts.Service == dawg.Carryout) {
			return dawg.ErrBadService
		}
		// BUG: setting the config field will implicitly change the config file
		a.conf.Service = a.gOpts.Service
	}

	if a.gOpts.LogFile != "" {
		dir := fp.Join(config.Folder(), "logs")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0700)
			if err != nil {
				return err
			}
		}
		a.logf, e = os.Create(fp.Join(dir, a.gOpts.LogFile))
		log.SetOutput(a.logf)
	}
	return errs.Pair(err, e)
}

func (a *App) getDBAddress(key string) (*obj.Address, error) {
	raw, err := a.db.WithBucket("addresses").Get(key)
	if err != nil {
		return nil, err
	}
	return obj.FromGob(raw)
}

func (a *App) postrun(*cobra.Command, []string) (err error) {
	if a.logf != nil {
		return a.logf.Close()
	}
	return nil
}

func logfile(name string) string {
	return fp.Join(config.Folder(), "logs", name)
}
