package config

import (
	"os"
	"os/exec"
)

// Edit will edit the config file using the editor set by $EDITOR
// or will default to use vim.
func Edit() error {
	cfg.changed = true
	return edit(cfg.file)
}

// EditFile will find an editor and run it to open the file given.
func EditFile(name string) error {
	return edit(name)
}

func edit(f string) error {
	editor := os.Getenv("Editor")
	if editor == "" {
		editor = DefaultEditor
	}

	cmd := exec.Command(editor, f)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
