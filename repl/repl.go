package repl

import "strings"

func CleanInput(text string) []string {
	words := strings.Fields(text)
	return words
}
