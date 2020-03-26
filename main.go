package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Discorgi command
type discorgiFetcher struct {
	names  []string
	help   string
	fetch  func(string) (string, error)
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
	botToken := flag.String("bot-token", "", "Discord bot token")
	steamToken := flag.String("steam-token", "", "Steam token")
	giphyToken := flag.String("giphy-token", "", "Giphy token")
	flag.Parse()

	if *botToken == "" {
		log.Fatal("Error: Missing bot-token command flag")
	} else if *steamToken == "" {
		log.Fatal("Error: Missing steam-token command flag")
	} else if *giphyToken == "" {
		log.Fatal("Error: Missing giphy-token command flag")
	}

	// Fetch steam games
	games, err := getSteamGames(*steamToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Fetched steam games")

	// Ticker to update steam games every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	gamesMutex := sync.Mutex{}
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Print("Updated steam games... ")
				gamesMutex.Lock()
				games, err = getSteamGames(*steamToken)
				if err != nil {
					log.Fatal(err)
				}
				gamesMutex.Unlock()
				fmt.Println("Complete")
			}
		}
	}()

	// Create discord session
	discord, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatal("Error creating discord session:", err)
	}

	// add commands to discorgi
	commands := []discorgiFetcher{
		discorgiFetcher{
			names: []string{"gif"},
			help:  "gif [search terms]",
			fetch: func(searchTerm string) (string, error) {
				searchTerm = strings.ReplaceAll(searchTerm, " ", "+")
				url := fmt.Sprintf("https://api.giphy.com/v1/gifs/search?api_key=%v&q=%v&limit=1&offset=0&rating=R&lang=en", *giphyToken, searchTerm)

				resp, err := http.Get(url)
				if err != nil {
					return "Woof! Something went wrong!", err
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "Woof! Something went wrong!", err
				}

				var container gifContainer
				err = json.Unmarshal(body, &container)
				if err != nil {
					return "Woof! Something went wrong!", err
				}

				if len(container.Data) < 1 {
					return "Woof! Can't sniff out the perfect gif.", nil
				}
				return container.Data[0].URL, err
			}},
		discorgiFetcher{
			names: []string{"steam"},
			help:  "steam [game name]",
			fetch: func(name string) (string, error) {
				// Default value
				url := "Sorry, couldn't sniff that one out 🔍"

				// Make a copy of current game cache
				gamesMutex.Lock()
				gamesInstance := games
				gamesMutex.Unlock()

				// Range over games to see if game exists
				for _, game := range gamesInstance {
					if name == strings.ToLower(game.Name) {
						url = fmt.Sprintf("https://store.steampowered.com/app/%v", game.SteamID)
						break
					}
				}
				return url, nil
			}},
		discorgiFetcher{
			names: []string{"define"},
			help:  "define [search terms]",
			fetch: func(searchTerm string) (string, error) {
				url := fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%v", searchTerm)

				resp, err := http.Get(url)
				if err != nil {
					return "Woof! Something went wrong!", err
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "Woof! Something went wrong!", err
				}

				var definitions definitionList
				err = json.Unmarshal(body, &definitions)
				if err != nil {
					return "Woof! Something went wrong!", err
				}

				if len(definitions.List) < 1 {
					return "Woof! Can't sniff it out on Urban Dictionary.", nil
				}
				def := strings.ReplaceAll(definitions.List[0].Definition, "\n", "\n> ")
				ex := strings.ReplaceAll(definitions.List[0].Example, "\n", "\n> ")
				return fmt.Sprintf("The Urban Dictionary defines %v as\n> %v\n_Example_:\n> %v", searchTerm, def, ex), nil
			}},
		discorgiFetcher{
			names:  []string{"who's a good boy", "whos a good boy", "whose a good boy"},
			help:   "who's a good boy",
			noArgs: true,
			fetch: func(s string) (string, error) {
				return "https://gfycat.com/femininedefiantgiantschnauzer-corgi-puppy-dog", nil
			}},
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
					msg, err = cmd.fetch("")
					if err != nil {
						fmt.Println("ERROR: ", err)
					}
					break
				} else if len(input) >= len(name)+2 && input[:len(name)+1] == name+" " { // Check if args exist

					msg, err = cmd.fetch(input[len(name)+1:])
					if err != nil {
						fmt.Println("ERROR: ", err)
					}
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
		log.Fatal("Error opening connection: ", err)
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

// Fetch games via steam api
func getSteamGames(steamAPIKey string) ([]steamGame, error) {
	url := fmt.Sprintf("http://api.steampowered.com/ISteamApps/GetAppList/v0002/?key=%v&format=json", steamAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var list steamResponse
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}

	return list.Applist.Apps, nil
}
