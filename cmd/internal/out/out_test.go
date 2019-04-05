package out

import (
	"testing"

	"github.com/harrybrwn/apizza/pkg/tests"
)

func TestFormatLine(t *testing.T) {
	exp := []string{
		"The menu command will show the dominos menu. To show a subdivition of the menu, ",
		"give an item or category to the --category and --item flags or give them as an ",
		"argument to the command itself.",
	}
	s := `The menu command will show the dominos menu. To show a subdivition of the menu, give an item or category to the --category and --item flags or give them as an argument to the command itself.`
	for i, line := range FormatLine(s, 80) {
		if exp[i] != line {
			t.Error("wrong line format")
		}
	}

	expected := `The menu command will show the dominos menu. To show a subdivition of the menu, 
    give an item or category to the --category and --item flags or give them as an 
    argument to the command itself.`
	tests.Compare(t, FormatLineIndent(s, 80, 4), expected)
}
