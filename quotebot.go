// QuoteBot X by Joseph DeLuca
package main

import (
	"encoding/json"
	"fmt"
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// location of configuration file
const configFile string = "./config.json"

type CommandConfig struct {
	Prefix string `json:"prefix"`
	Add    string `json:"add"`
	By     string `json:"by"`
}

// BotConfig internal struct for configuration management
type BotConfig struct {
	DiscordToken     string        `json:"discord-token"`
	ConnectionString string        `json:"connection-string"`
	CommandConfig    CommandConfig `json:"command-config"`
}

// Used for general error checking and panicking
func checkError(err error, message string) {
	if err != nil {
		fmt.Println(message + err.Error())
		log.Fatal(err)
	}
}

// Retrieves the Bot DiscordToken
// Will eventually get the JSON configuration for quoteBot
func getConfig(file string) BotConfig {
	// read the fileData of file
	fileData, err := os.ReadFile(file)
	checkError(err, "Error reading discord token: ")

	//turn config.json into a BotConfig struct
	var configuration BotConfig
	err = json.Unmarshal(fileData, &configuration)
	checkError(err, "Error unmarshalling json: ")

	return configuration
}

func ready(session *discordgo.Session, _ *discordgo.Ready) {
	err := session.UpdateGameStatus(0, "q!")
	if err != nil {
		fmt.Println("Error updated Bot Status")
	}
}

func guildCreateHandler(bm data.Manager, commandMap map[string][]string) func(*discordgo.Session, *discordgo.GuildCreate) {
	return func(session *discordgo.Session, event *discordgo.GuildCreate) {
		if event.Guild.Unavailable {
			return
		}

		if !bm.GuildExists(*event.Guild) {
			bm.AddGuild(*event.Guild)
		}
		fmt.Println("Login: ", event.Guild.Name)

		commandMap[event.Guild.ID] = registerAllCommands(session, event.Guild.ID)

		for _, channel := range event.Guild.Channels {
			if channel.ID == event.Guild.ID {
				_, _ = session.ChannelMessageSend(channel.ID, "QuoteBot is ready! Type q!")
				return
			}
		}
	}
}

func main() {
	// INITIALIZATION ---------------------------------------------------------
	// Store the application configuration
	botConfig := getConfig(configFile)

	// Starts the data manager for the bot
	botManager := data.Start(botConfig.ConnectionString)
	// defers the graceful shutdown of the data manager
	defer botManager.Shutdown()

	// Initialize the Discord Bot
	session, err := discordgo.New("Bot " + botConfig.DiscordToken)
	checkError(err, "Error creating Discord Session: ")

	// a map of registered commands by server ID
	var commandMap = make(map[string][]string)

	// EVENT HANDLING ---------------------------------------------------------
	// Define Handlers for discord events.
	guildCreate := guildCreateHandler(botManager, commandMap)
	interactionCreate := interactionCreateHandler(botManager)

	// Attach Handlers to the discord session
	session.AddHandler(ready)
	session.AddHandler(guildCreate)
	session.AddHandler(interactionCreate)

	// START SESSION ----------------------------------------------------------
	// Open Discord session
	err = session.Open()
	checkError(err, "Error opening Discord session: ")

	// Defers a call to Close the Discord Session
	defer func(ds *discordgo.Session) {
		for guildID, commandIDs := range commandMap {
			removeAllCommands(ds, guildID, commandIDs)
		}

		err := ds.Close()
		checkError(err, "Error closing Discord session: ")
	}(session)

	// Start Message
	fmt.Println("Welcome to QuoteBot X. Press CTRL+C to exit.")

	// Creates Signal Interrupt Channels
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
