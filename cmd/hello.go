package cmd

import (
	"fmt"

	"github.com/harrybrwn/apizza/cmd/internal/base"
	"github.com/spf13/cobra"
)

type helloCmd struct {
	*base.Command
}

func (c *helloCmd) Run(cmd *cobra.Command, args []string) error {
	fmt.Fprintf(c.Output(), "hello?\n")
	return nil
}

func newHelloCmd(b base.Builder) base.CliCommand {
	c := &helloCmd{}
	c.Command = base.NewCommand(
		"hello",
		"say hello",
		c.Run,
	)
	c.Cmd().Hidden = true

	// c.Command = b.Build("hello", "say hello", c)
	return c.Command
}
