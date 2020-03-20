package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	config, err := getConfig("config.json")
	discord, err := discordgo.New("Bot " + config["bot-token"])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(discord)
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
