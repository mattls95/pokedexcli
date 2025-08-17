package repl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

func StartRepl() {
	fmt.Print("Welcome to the Pokedex!\n")
	fmt.Print("Usage:\n")
	for _, desc := range supportedCommands {
		fmt.Printf("%s: %s\n", desc.name, desc.description)
	}
	reader := bufio.NewScanner(os.Stdin)
	cfg := new(config)
	cfg = &config{
		pokecache: *pokecache.NewCache((5 * time.Second)),
	}
	cfg.pokemons = make(map[string]Pokemon)
	for {
		fmt.Print("Pokedex >")
		reader.Scan()
		words := CleanInput(reader.Text())
		cfg.args = words
		switch words[0] {
		case "exit":
			commandExit(cfg)
			continue
		case "help":
			commandHelp(cfg)
			continue
		case "map":
			commandMap(cfg)
			continue
		case "mapb":
			commandMapb(cfg)
			continue
		case "explore":
			commandExplore(cfg)
			continue
		case "catch":
			commandCatch(cfg)
			continue
		}
	}

}

func CleanInput(text string) []string {
	words := strings.Fields(text)
	return words
}

func commandExit(config *config) error {
	fmt.Print("\nClosing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *config) error {
	fmt.Print("Display a help message\n")
	return nil
}

func commandPokedex(config *config) error {
	for _, name := range config.pokemons {
		fmt.Printf("- %s", name)
	}
	return nil
}

func commandInspect(config *config) error {
	if len(config.args) < 2 {
		fmt.Println("Please provide a pokemon to inspect, e.g., 'pikachu'")
		return nil
	}
	pokemon, exists := config.pokemons[config.args[1]]
	if !exists {
		fmt.Println("you have not caught that pokemon")
		return nil
	}
	fmt.Printf("Name: %s", pokemon.Name)
	fmt.Printf("Height: %d", pokemon.Height)
	fmt.Printf("Weight: %d", pokemon.Weight)
	for _, stat := range pokemon.Stats {
		fmt.Println("Stats:")
		fmt.Printf("  -hp: %d\n", stat.Hp)
		fmt.Printf("  -attack: %d\n", stat.Attack)
		fmt.Printf("  -defense: %d\n", stat.Defense)
		fmt.Printf("  -special-attack: %d\n", stat.SpecialAttack)
		fmt.Printf("  -special-defense: %d\n", stat.SpecialDefense)
		fmt.Printf("  -speed: %d\n", stat.Speed)
	}
	fmt.Println("Types:")
	for _, typ := range pokemon.Types {
		fmt.Printf("  - normal: %s\n", typ.Normal)
		fmt.Printf("  - flying: %s\n", typ.Flying)
	}
	return nil
}

func commandCatch(config *config) error {
	if len(config.args) < 2 {
		fmt.Println("Please provide a pokemon to catch, e.g., 'pikachu'")
		return nil
	}
	pokemon := config.args[1]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)
	fullURL := "https://pokeapi.co/api/v2/pokemon" + "/" + pokemon

	res, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("get failed")
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if res.StatusCode > 299 {
		return fmt.Errorf("get failed")
	}
	if err != nil {
		return fmt.Errorf("get failed")
	}
	data := RespPokemon{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("get failed")
	}
	randomNumber := rand.Intn(data.Pokemon.BaseExperince + 1)
	chance := 50
	var catched bool
	if randomNumber < chance-data.Pokemon.BaseExperince/2 {
		catched = true
	} else {
		catched = false
	}
	if catched {
		config.pokemons[pokemon] = Pokemon{
			Name: data.Pokemon.Name,
		}
		fmt.Println("Pokemon catched")
	} else {
		fmt.Println("Pokemon got away")
	}
	return nil
}

func commandExplore(config *config) error {
	if len(config.args) < 2 {
		fmt.Println("Please provide a location area to explore, e.g., 'explore pastoria-city-area'")
		return nil
	}
	fullURL := "https://pokeapi.co/api/v2/location-area" + "/" + config.args[1]
	res, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("get failed")
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if res.StatusCode > 299 {
		return fmt.Errorf("get failed")
	}
	if err != nil {
		return fmt.Errorf("get failed")
	}
	data := RespPokemons{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("get failed")
	}

	for _, name := range data.PokemonEncounters {
		fmt.Printf("%s\n", string(name.Pokemon.Name))
	}
	return nil
}

func commandMap(config *config) error {
	var res *http.Response
	var err error
	if config.next == "" {
		res, err = http.Get("https://pokeapi.co/api/v2/location-area")
		if config.next == "" {
			res, err = http.Get("https://pokeapi.co/api/v2/location-area")
			if err != nil {
				return fmt.Errorf("get failed")
			}
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if res.StatusCode > 299 {
				return fmt.Errorf("get failed")
			}
			if err != nil {
				return fmt.Errorf("get failed")
			}
			data := RespShallowLocations{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				return fmt.Errorf("get failed")
			}
			if data.Next != nil {
				config.next = *data.Next
			}
			if data.Previous != nil {
				config.previous = *data.Previous
			}
			for _, name := range data.Results {
				fmt.Printf("%s\n", string(name.Name))
			}
			config.pokecache.Add("https://pokeapi.co/api/v2/location-area", body)
			return nil
		}
	} else {
		data, exists := config.pokecache.Get(config.next)
		if exists {
			dataExists := RespShallowLocations{}
			err = json.Unmarshal(data, &dataExists)
			if err != nil {
				return fmt.Errorf("get failed")
			}
			if dataExists.Next != nil {
				config.next = *dataExists.Next
			}
			if dataExists.Previous != nil {
				config.previous = *dataExists.Previous
			}

			for _, name := range dataExists.Results {
				fmt.Printf("%s\n", string(name.Name))
			}
			return nil
		} else {
			res, err = http.Get(config.next)
			if err != nil {
				return fmt.Errorf("get failed")
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)

			if res.StatusCode > 299 {
				return fmt.Errorf("get failed")
			}
			if err != nil {
				return fmt.Errorf("get failed")
			}
			config.pokecache.Add(config.next, body)
			data := RespShallowLocations{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				return fmt.Errorf("get failed")
			}
			if data.Next != nil {
				config.next = *data.Next
			}
			if data.Previous != nil {
				config.previous = *data.Previous
			}

			for _, name := range data.Results {
				fmt.Printf("%s\n", string(name.Name))
			}
		}
	}
	return nil
}

func commandMapb(config *config) error {
	var res *http.Response
	var err error
	if config.previous == "" {
		fmt.Println("you're on the first page")
		return nil
	} else {
		data, exists := config.pokecache.Get(config.previous)
		if exists {
			dataExists := RespShallowLocations{}
			err = json.Unmarshal(data, &dataExists)
			if err != nil {
				return fmt.Errorf("get failed")
			}
			if dataExists.Next != nil {
				config.next = *dataExists.Next
			}
			if dataExists.Previous != nil {
				config.previous = *dataExists.Previous
			}
			for _, name := range dataExists.Results {
				fmt.Printf("%s\n", string(name.Name))
			}
		}
		res, err = http.Get(config.previous)
	}
	if err != nil {
		return fmt.Errorf("get failed")
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if res.StatusCode > 299 {
		return fmt.Errorf("get failed")
	}
	if err != nil {
		return fmt.Errorf("get failed")
	}
	data := RespShallowLocations{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return fmt.Errorf("get failed")
	}
	if data.Next != nil {
		config.next = *data.Next
	}
	if data.Previous != nil {
		config.previous = *data.Previous
	}

	for _, name := range data.Results {
		fmt.Printf("%s\n", string(name.Name))
	}
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	next      string
	previous  string
	pokecache pokecache.Cache
	pokemons  map[string]Pokemon
	args      []string
}

type RespShallowLocations struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type RespPokemons struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Weight int    `json:"weight"`
	Stats  []struct {
		Hp             int `json:"hp"`
		Attack         int `json:"attack"`
		Defense        int `json:"defense"`
		SpecialAttack  int `json:"special-attack"`
		SpecialDefense int `json:"special-defense"`
		Speed          int `json:"spedd"`
	}
	Types []struct {
		Normal string `json:"normal"`
		Flying string `json:"flying"`
	}
}

type RespPokemon struct {
	Pokemon struct {
		Name          string `json:"name"`
		BaseExperince int    `json:"base_experience"`
	} `json:"pokemon"`
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
	"mapb": {
		name:        "mapb",
		description: "Displays locations of pokemons",
		callback:    commandMapb,
	},
	"explore": {
		name:        "explore",
		description: "Displays pokemons",
		callback:    commandExplore,
	},
	"catch": {
		name:        "explore",
		description: "Catch pokemons",
		callback:    commandCatch,
	},
	"inspect": {
		name:        "inspect",
		description: "Inspect pokemons",
		callback:    commandInspect,
	},
	"pokedex": {
		name:        "pokedex",
		description: "Prints pokemons",
		callback:    commandPokedex,
	},
}
