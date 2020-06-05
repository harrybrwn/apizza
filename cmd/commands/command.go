package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/harrybrwn/apizza/cmd/cli"
	"github.com/spf13/cobra"
)

// Color toggles output color
//
// TODO: this is shitty code FIXME!!!
var Color = true

// NewCompletionCmd creates a new command for shell completion.
func NewCompletionCmd(b cli.Builder) *cobra.Command {
	var validArgs = []string{"zsh", "bash", "ps", "powershell", "fish"}

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

			if len(args) == 0 {
				return fmt.Errorf(
					"no shell type given; (expected %s, or %s)",
					strings.Join(validArgs[:len(validArgs)-1], ", "),
					validArgs[len(validArgs)-1])
			}
			switch args[0] {
			case "zsh":
				return root.GenZshCompletion(out)
			case "ps", "powershell":
				return root.GenPowerShellCompletion(out)
			case "bash":
				return root.GenBashCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, false)
			}
			return errors.New("unknown shell type")
		},
		ValidArgs: validArgs,
		Aliases:   []string{"comp"},
	}
	return cmd
}
