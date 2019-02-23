package cmd

import (
	"github.com/spf13/cobra"
)

// TODO: re-write Execute func with the new cli builder system... ex.
// func Execute() {
// 	builder := cliBuilder{}
// }

type cliCommand interface {
	command() *cobra.Command
	addCommands(...cliCommand)
	run(*cobra.Command, []string) error
}

type cliBuilder struct {
	root     cliCommand
	commands []cliCommand
}

type builder interface {
	build()
	add(cliCommand)
}

func (b *cliBuilder) build() {
	b.addAll()
	b.root.addCommands(b.commands...)
}

func (b *cliBuilder) add(cmd cliCommand) {
	b.commands = append(b.commands, cmd)
}

func (b *cliBuilder) addAll() {
	// menually add all commands here with new<cmd name> functions
	b.commands = []cliCommand{}
}

// func (b *cliBuilder) newApizzaCmd() cliCommand {
// 	return newApizzaCmd(cobra.Command{
// 		Use:   "apizza",
// 		Short: "Dominos pizza from the command line.",
// 		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
// 			return config.Save()
// 		},
// 	})
// }
