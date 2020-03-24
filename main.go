package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Discorgi command
type discorgiFetcher struct {
	names  []string
	help   string
	fetch  func(string) string
	noArgs bool
}

//// Structs use in commands ////
type steamResponse struct {
	Applist struct {
		Apps []steamGame `json:"apps"`
	} `json:"applist"`
}

type steamGame struct {
	SteamID int    `json:"appid"`
	Name    string `json:"name"`
}

type gifContainer struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

type definitionList struct {
	List []struct {
		Definition string `json:"definition"`
		Example    string `json:"example"`
	} `json:"list"`
}

func main() {
	// Fetch config
	config, err := getConfig("config.json")
	if err != nil {
		log.Fatal("Error opening config.json:", err)
	}
	fmt.Println("Fetched config")

	// Fetch steam games
	games := getSteamGames(config["steam-token"])

	// Ticker to update steam games every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Print("Updated steam games... ")
				games = getSteamGames(config["steam-token"])
				fmt.Println("Complete")
			}
		}
	}()
	fmt.Println("Fetched steam games")

	// Create discord session
	discord, err := discordgo.New("Bot " + config["bot-token"])
	if err != nil {
		fmt.Println("Error creating discord session:", err)
		return
	}

	// add commands to discorgi
	commands := []discorgiFetcher{
		discorgiFetcher{
			names: []string{"gif"},
			help:  "gif [search terms]",
			fetch: func(searchTerm string) string {
				searchTerm = strings.ReplaceAll(searchTerm, " ", "+")
				url := fmt.Sprintf("https://api.giphy.com/v1/gifs/search?api_key=%v&q=%v&limit=1&offset=0&rating=R&lang=en", config["giphy-token"], searchTerm)

				resp, err := http.Get(url)
				if err != nil {
					fmt.Println("gif get request error: ", err)
					return "Woof! Something went wrong!"
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("gif body decord error: ", err)
					return "Woof! Something went wrong!"
				}

				var container gifContainer
				json.Unmarshal(body, &container)

				if len(container.Data) < 1 {
					return "Woof! Can't sniff out the perfect gif."
				}
				return container.Data[0].URL
			}},
		discorgiFetcher{
			names: []string{"steam"},
			help:  "steam [game name]",
			fetch: func(name string) string {
				url := "Sorry, couldn't sniff that one out 🔍"

				for _, game := range games {
					if name == strings.ToLower(game.Name) {
						url = fmt.Sprintf("https://store.steampowered.com/app/%v", game.SteamID)
					}
				}
				return url
			}},
		discorgiFetcher{
			names: []string{"define"},
			help:  "define [search terms]",
			fetch: func(searchTerm string) string {
				url := fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%v", searchTerm)

				resp, err := http.Get(url)
				if err != nil {
					fmt.Println("urbanDictionary get request error: ", err)
					return "Woof! Something went wrong!"
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println("urbanDictionary body decord error: ", err)
					return "Woof! Something went wrong!"
				}

				var definitions definitionList
				json.Unmarshal(body, &definitions)

				if len(definitions.List) < 1 {
					return "Woof! Can't sniff it out on Urban Dictionary."
				}
				def := strings.ReplaceAll(definitions.List[0].Definition, "\n", "\n> ")
				ex := strings.ReplaceAll(definitions.List[0].Example, "\n", "\n> ")
				return fmt.Sprintf("The Urban Dictionary defines %v as\n> %v\n_Example_:\n> %v", searchTerm, def, ex)
			}},
		discorgiFetcher{
			names:  []string{"who's a good boy", "whos a good boy", "whose a good boy"},
			help:   "who's a good boy",
			noArgs: true,
			fetch:  func(s string) string { return "https://gfycat.com/femininedefiantgiantschnauzer-corgi-puppy-dog" }},
	}

	// Register handler
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		content := strings.ToLower(m.Content)

		// Ignore bot's messages
		if m.Author.ID == s.State.User.ID {
			return
		}

		if len(content) < 9 || content[:9] != "discorgi " {
			return
		}
		fmt.Printf("Woof! Message received! %v said, \"%v\"\n", m.Author, m.Content)

		input := content[9:]

		if input == "help" || input == "halp" {
			msg := "To ask me to do something, just say `discorgi [command goes here]`\n You use the following commands:\n```\n"
			for _, cmd := range commands {
				msg += cmd.help + "\n"
			}
			msg += "```"
			s.ChannelMessageSend(m.ChannelID, msg)
		}

		msg := "I haven't learned that trick yet! 🐕"

		for _, cmd := range commands {
			for _, name := range cmd.names {
				// No arguments to command
				if cmd.noArgs && name == input {
					msg = cmd.fetch("")
					break
				} else if len(input) >= len(name)+2 && input[:len(name)+1] == name+" " { // Check if args exist

					msg = cmd.fetch(input[len(name)+1:])
					break
				}
			}
			if msg != "I haven't learned that trick yet! 🐕" {
				break
			}
		}
		s.ChannelMessageSend(m.ChannelID, msg)
	})

	// Begin listening
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}

	// Close at end of session
	defer discord.Close()

	// Wait for SIGNAL to end program
	fmt.Println("Discorgi is patiently listening. Press CTRL-C to exit")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	select {
	case <-sc:
		fmt.Println("\nExiting...")
	}
}

//// Helper Functions ////

// Fetch config from json
func getConfig(path string) (map[string]string, error) {
	configJSON, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configJSON.Close()

	var config map[string]string
	jsonBytes, _ := ioutil.ReadAll(configJSON)
	json.Unmarshal([]byte(jsonBytes), &config)

	return config, nil
}

// Fetch games via steam api
func getSteamGames(steamAPIKey string) []steamGame {
	url := fmt.Sprintf("http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=%v&format=json", steamAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("fetch steam game get request error: ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("fetch steam game body decode error: ", err)
	}

	var list steamResponse
	json.Unmarshal(body, &list)
	return list.Applist.Apps
}
