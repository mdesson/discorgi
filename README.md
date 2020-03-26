# discorgi
Discord bot. Fetches things for friends. Is good boy.

## Running Discorgi

The current version uses CLI flags to run:

```
bot-token // discord bot's secret token
steam-token // steam api token
giphy-token // giphy api token

// to run
go run main.go -bot-token=TOKEN -steam-token=TOKEN -giphy-token=TOKEN
```
## Discorgi Commands

Discorgi is a loyal little bot who loves to fetch things. 

All commands are case-insensitive and you can get a list with `discorgi help` or `discorgi halp`.

Currently he can:

* `discorgi gif search terms go here`: Searches for gifs on giphy and displays the first result
* `discorgi define search terms go here`: Grabs the first urban dictionary definition it can find
* `discorgi steam game name goes here`: Search for a steam game of the exact name as the search term
* `discorgi who's a good boy`: Discorgi is

## Extending Discorgi

Discorgi is pretty easy to extend. Just add a `discorgiFetcher` to the existing slice in the main method. Discorgi will take care of the rest for you, and your command will automatically be added to the help command.

* `names []string`: This is the name of the command, and what will be used to trigger the action. Define multiple if you want include mispellings or shortcuts. Remember, discorgi is case insensitive.
* `help string`: Text displaying how to use the command, such as `gif [search terms]`
* `fetch func(string) (string, error)`: Function that returns the string to be displayed in discord.
*	`noArgs bool`: Flag to signify that this command takes no arguments


