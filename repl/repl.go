package repl

import (
	"fmt"
	"os"
	"strings"
)

func StartRepl() {
	fmt.Print("Welcome to the Pokedex!\n")
	fmt.Print("Usage:\n")
	for _, desc := range supportedCommands {
		fmt.Printf("%s: %s\n", desc.name, desc.description)
	}
	commandExit()
}

func CleanInput(text string) []string {
	words := strings.Fields(text)
	return words
}

func commandExit() error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Print("Display a help message")
	return nil
}

func commandMap() error {
	
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func getCommands() map[string]cliCommand {
	return supportedCommands
}

var supportedCommands = map[string]cliCommand{
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
	"help": {
		name:        "help",
		description: "Display a help message",
		callback:    commandHelp,
	},
	"map": {
		name:        "map",
		description: "Displays locations of pokemons",
		callback:    commandMap,
	},
}
