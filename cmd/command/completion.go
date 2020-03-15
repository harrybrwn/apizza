package command

import (
	"errors"

	"github.com/spf13/cobra"
)

// CompletionCmd is the completion command.
var CompletionCmd = &cobra.Command{
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
		if len(args) < 1 {
			if err = root.GenBashCompletion(out); err != nil {
				return err
			}
			return nil
		}

		if args[0] == "zsh" {
			return root.GenZshCompletion(out)
		} else if args[0] == "ps" || args[0] == "powershell" {
			return root.GenPowerShellCompletion(out)
		} else if args[0] == "bash" {
			return root.GenBashCompletion(out)
		}
		return errors.New("unknown shell type")
	},
	ValidArgs: []string{"zsh", "bash", "ps", "powershell"},
	Aliases:   []string{"comp"},
}
