package cmd

import (
	"apizza/dawg"
	"apizza/pkg/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute() {
	builder := cliBuilder{root: newApizzaCmd()}

	err := config.SetConfig(".apizza", cfg)
	if err != nil {
		handle(err)
	}

	err = initDatabase()
	if err != nil {
		handle(err)
	}

	addr = builder.getAddress()

	defer func() {
		err = db.Close()
		if err != nil {
			handle(err)
		}
		err = config.Save()
		if err != nil {
			handle(err)
		}
	}()

	_, err = builder.exec()
	if err != nil {
		handle(err)
	}
}

func handle(e error) { fmt.Println(e); os.Exit(1) }

type cliCommand interface {
	command() *cobra.Command
	AddCmd(...cliCommand)
	run(*cobra.Command, []string) error
}

type basecmd struct {
	cmd *cobra.Command
}

func (bc *basecmd) command() *cobra.Command {
	return bc.cmd
}

func (bc *basecmd) AddCmd(cmds ...cliCommand) {
	for _, cmd := range cmds {
		bc.cmd.AddCommand(cmd.command())
	}
}

func (bc *basecmd) run(cmd *cobra.Command, args []string) error {
	return bc.cmd.Usage()
}

type runFunc func(*cobra.Command, []string) error

func newVerboseBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{cmd: &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  f,
	}}
	if f == nil {
		// f = base.run // i guess this works too, cool
		base.cmd.RunE = base.run
	}
	return base
}

func newBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{cmd: &cobra.Command{
		Use:           use,
		Short:         short,
		RunE:          f,
		SilenceErrors: true,
		SilenceUsage:  true,
	}}
	if f == nil {
		// f = base.run // i guess this works too, cool
		base.cmd.RunE = base.run
	}

	return base
}

type cliBuilder struct {
	root     cliCommand
	commands []cliCommand
}

type commandBuilder interface {
	exec()
	add(cliCommand)
}

func (b *cliBuilder) exec() (*cobra.Command, error) {
	b.root.AddCmd(
		newOrderCommand(),
		newMenuCmd(),
		newConfigCmd(),
	)
	return b.root.command().ExecuteC()
}

func (b *cliBuilder) add(cmd cliCommand) {
	b.commands = append(b.commands, cmd)
}

func (b *cliBuilder) getAddress() *dawg.Address {
	addrStr := b.root.(*apizzaCmd).address
	if addrStr == "" {
		return &dawg.Address{
			Street: cfg.Address.Street,
			City:   cfg.Address.City,
			State:  cfg.Address.State,
			Zip:    cfg.Address.Zip,
		}
	}
	return dawg.ParseAddress(addrStr)
}
