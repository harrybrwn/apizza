package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/obj"
	"github.com/harrybrwn/apizza/pkg/cache"
	"github.com/spf13/cobra"
)

// NewAddAddressCmd creates the 'add-address' command.
func NewAddAddressCmd(b cli.Builder, in io.Reader) cli.CliCommand {
	c := &addAddressCmd{
		db: b.DB(),
		in: in,
	}
	c.CliCommand = b.Build("add-address", "Add a new named address to the internal storage.", c)
	cmd := c.Cmd()
	cmd.Aliases = []string{"add-addr", "addaddr", "aa"}
	cmd.Hidden = true
	return c
}

type addAddressCmd struct {
	cli.CliCommand

	db cache.Storage
	in io.Reader
}

func (a *addAddressCmd) Run(cmd *cobra.Command, args []string) error {
	r := reader{bufio.NewReader(a.in)}
	addr := obj.Address{}

	a.Printf("Address Name: ")
	name, err := r.readline()
	if err != nil {
		return err
	}
	a.Printf("Street Address: ")
	addr.Street, err = r.readline()
	if err != nil {
		return err
	}
	a.Printf("City: ")
	addr.CityName, err = r.readline()
	if err != nil {
		return err
	}
	a.Printf("State Code: ")
	addr.State, err = r.readline()
	if err != nil {
		return err
	}
	a.Printf("Zipcode: ")
	addr.Zipcode, err = r.readline()
	if err != nil {
		return err
	}

	fmt.Print(name, ":\n", addr, "\n")
	return nil
}

type reader struct {
	scanner *bufio.Reader
}

func (r *reader) readline() (string, error) {
	lineone, err := r.scanner.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(lineone, "\n \t\r"), nil
}
