package commands

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
		db:  b.DB(),
		in:  in,
		new: false,
	}
	c.CliCommand = b.Build("address", "Add a new named address to the internal storage.", c)
	cmd := c.Cmd()
	cmd.Aliases = []string{"addr"}
	cmd.Flags().BoolVarP(&c.new, "new", "n", c.new, "add a new address")
	cmd.Flags().StringVarP(&c.delete, "delete", "d", "", "delete an address")
	return c
}

type addAddressCmd struct {
	cli.CliCommand

	db     *cache.DataBase
	in     io.Reader
	new    bool
	delete string
}

func (a *addAddressCmd) Run(cmd *cobra.Command, args []string) error {
	if a.new {
		return a.newAddress()
	}
	if a.delete != "" {
		db := a.db.WithBucket("addresses")
		return db.Delete(a.delete)
	}

	m, err := a.db.WithBucket("addresses").Map()
	if err != nil {
		return err
	}

	var addr *obj.Address
	for key, val := range m {
		addr, err = obj.FromGob(val)
		if err != nil {
			return err
		}

		a.Printf("%s:\n  %s\n", key, obj.AddressFmtIndent(addr, 2))
	}
	return nil
}

type reader struct {
	scanner *bufio.Reader
}

func (a *addAddressCmd) newAddress() error {
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
	raw, err := obj.AsGob(&addr)
	if err != nil {
		return err
	}
	return a.db.WithBucket("addresses").Put(name, raw)
}

func (r *reader) readline() (string, error) {
	lineone, err := r.scanner.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(lineone, "\n \t\r"), nil
}
