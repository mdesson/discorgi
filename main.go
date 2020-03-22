package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

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
		Definition  string        `json:"definition"`
		Permalink   string        `json:"permalink"`
		ThumbsUp    int           `json:"thumbs_up"`
		SoundUrls   []interface{} `json:"sound_urls"`
		Author      string        `json:"author"`
		Word        string        `json:"word"`
		Defid       int           `json:"defid"`
		CurrentVote string        `json:"current_vote"`
		WrittenOn   time.Time     `json:"written_on"`
		Example     string        `json:"example"`
		ThumbsDown  int           `json:"thumbs_down"`
	} `json:"list"`
}

func main() {
	// TODO: set up a ticker to periodically refresh steam game cache
	// Fetch steam games
	games, err := getSteamGames("steamgames.json")
	if err != nil {
		fmt.Println("Error opening steamgames.json:", err)
		return
	}

	fmt.Println("Fetched steam games")

	// Fetch config
	config, err := getConfig("config.json")
	if err != nil {
		fmt.Println("Error opening config.json:", err)
		return
	}

	fmt.Println("Fetched config")

	// Create discord session
	discord, err := discordgo.New("Bot " + config["bot-token"])
	if err != nil {
		fmt.Println("Error creating discord session:", err)
		return
	}

	// Register handlers
	discord.AddHandler(discorgiListener(games, config))

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

	// select {
	// case <-sc:

	// }
	<-sc
	fmt.Println("\nExiting...")
}

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

func getSteamGames(path string) ([]steamGame, error) {
	gameJSON, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer gameJSON.Close()

	jsonBytes, _ := ioutil.ReadAll(gameJSON)
	games := make([]steamGame, 0)
	json.Unmarshal([]byte(jsonBytes), &games)

	return games, nil
}

func discorgiListener(games []steamGame, config map[string]string) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// TODO: Make a struct for commands with properies for name, help, and a function returning a string
		commands := []string{"help", "steam [game name]", "gif [search term]", "define [search term]", "who's a good boy"}
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
			msg := "To ask me to do something, just say `discorgi [command goes here]`\n You use the following commands:\n```\n" + strings.Join(commands, "\n") + "\n```"
			s.ChannelMessageSend(m.ChannelID, msg)
		}
		if input == "whos a good boy" || input == "who's a good boy" {
			s.ChannelMessageSend(m.ChannelID, "https://gfycat.com/femininedefiantgiantschnauzer-corgi-puppy-dog")
		}

		// If message is ping, reply with pong
		if len(input) >= 7 && input[:6] == "steam " {
			msg := fetchGame(input[6:], games)
			s.ChannelMessageSend(m.ChannelID, msg)
		}

		if len(input) >= 5 && input[:4] == "gif " {
			msg := fetchGif(input[4:], config["giphy-token"])
			s.ChannelMessageSend(m.ChannelID, msg)
		}

		if len(input) >= 8 && input[:7] == "define " {
			msg := fetchDefinition(input[7:])
			s.ChannelMessageSend(m.ChannelID, msg)
		}
	}

}

func fetchGame(name string, games []steamGame) string {
	url := "Sorry, couldn't sniff that one out 🔍"

	for _, game := range games {
		if name == strings.ToLower(game.Name) {
			url = fmt.Sprintf("https://store.steampowered.com/app/%v", game.SteamID)
		}
	}

	return url
}

func fetchGif(searchTerm, giphyToken string) string {
	searchTerm = strings.ReplaceAll(searchTerm, " ", "+")
	url := fmt.Sprintf("https://api.giphy.com/v1/gifs/search?api_key=%v&q=%v&limit=1&offset=0&rating=R&lang=en", giphyToken, searchTerm)

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
}

func fetchDefinition(searchTerm string) string {
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
	return fmt.Sprintf("The Urban Dictionary defines %v as\n> %v", searchTerm, definitions.List[0].Definition)
}
