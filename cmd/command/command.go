package command

import (
	"errors"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/harrybrwn/apizza/cmd/internal/data"
	"github.com/spf13/cobra"
)

// NewCompletionCmd creates a new command for shell completion.
func NewCompletionCmd(b cli.Builder) *cobra.Command {
	var (
		listOrders    bool
		listAddresses bool
	)

	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|powershell]",
		Short: "Generate bash, zsh, or powershell completion",
		Long: `Generate bash, zsh, or powershell completion
just add '. <(apizza completion <shell name>)' to you .bashrc or .zshrc
note: for zsh you will need to run 'compdef _apizza apizza'`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			root := cmd.Root()
			out := cmd.OutOrStdout()

			if listOrders {
				orders := data.ListOrders(b.DB())
				cmd.Print(strings.Join(orders, " "))
				return nil
			}
			if listAddresses {
				m, err := b.DB().WithBucket("addresses").Map()
				if err != nil {
					return err
				}
				keys := make([]string, 0, len(m))
				for key := range m {
					keys = append(keys, key)
				}
				cmd.Print(strings.Join(keys, " "))
				return nil
			}
			if len(args) == 0 {
				return errors.New("no shell type given")
			}
			switch args[0] {
			case "zsh":
				return root.GenZshCompletion(out)
			case "ps", "powershell":
				return root.GenPowerShellCompletion(out)
			case "bash":
				return root.GenBashCompletion(out)
			}
			return errors.New("unknown shell type")
		},
		ValidArgs: []string{"zsh", "bash", "ps", "powershell"},
		Aliases:   []string{"comp"},
	}

	flg := cmd.Flags()
	flg.BoolVar(&listOrders, "list-orders", false, "")
	flg.BoolVar(&listAddresses, "list-addresses", false, "")
	flg.MarkHidden("list-orders")
	flg.MarkHidden("list-addresses")
	return cmd
}
