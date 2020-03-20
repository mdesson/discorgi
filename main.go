package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// Fetch config
	config, err := getConfig("config.json")
	if err != nil {
		fmt.Println("Error opening config.json:", err)
		return
	}

	// Create discord session
	discord, err := discordgo.New("Bot " + config["bot-token"])
	if err != nil {
		fmt.Println("Error creating discord session:", err)
		return
	}

	// Register handler
	discord.AddHandler(messageCreate)

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

func getConfig(path string) (map[string]string, error) {
	configJSON, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer configJSON.Close()

	var config map[string]string
	jsonBytes, _ := ioutil.ReadAll(configJSON)
	json.Unmarshal([]byte(jsonBytes), &config)

	return config, nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Woof! Message received!")

	// Ignore bot's messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If message is ping, reply with pong
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}
