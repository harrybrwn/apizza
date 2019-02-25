package cmd

import (
	"apizza/pkg/config"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute() {
	builder := cliBuilder{root: newApizzaCmd()}

	err := initDatabase()
	if err != nil {
		handle(err)
	}

	err = config.SetConfig(".apizza", cfg)
	if err != nil {
		handle(err)
	}

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

func newBaseCommand(use, short string, f runFunc) *basecmd {
	return &basecmd{cmd: &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  f,
	}}
}

func newSilentBaseCommand(use, short string, f runFunc) *basecmd {
	base := &basecmd{cmd: &cobra.Command{
		Use:           use,
		Short:         short,
		RunE:          f,
		SilenceErrors: true,
		SilenceUsage:  true,
	}}

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
