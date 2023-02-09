// Quotebot X by Joseph DeLuca
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DeLucaJ/quotebot/internal/botdata"
	"github.com/bwmarrin/discordgo"
)

// location of configuration file
const configfile string = "./config.json"

// internal struct for configuration management
type BotConfig struct {
	DiscordToken string
	DBuri        string
}

// Used for general error checking and panicing
func checkError(err error, message string) {
	if err != nil {
		fmt.Println(message + err.Error())
		log.Fatal(err)
	}
}

// Retrieves the Bot DiscordToken
// Will eventually get the JSON configuration for quotebot
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

func ready(session *discordgo.Session, event *discordgo.Ready) {
	err := session.UpdateGameStatus(0, "q!")
	if err != nil {
		fmt.Println("Error updated Bot Status")
	}
}

func guildCreateEvent(bm botdata.Manager) func(*discordgo.Session, *discordgo.GuildCreate) {
	return func(session *discordgo.Session, event *discordgo.GuildCreate) {
		if event.Guild.Unavailable {
			return
		}

		if !bm.GuildExists(*event.Guild) {
			bm.AddGuild(*event.Guild)
		}
		fmt.Println("Login: ", event.Guild.Name)

		for _, channel := range event.Guild.Channels {
			if channel.ID == event.Guild.ID {
				_, _ = session.ChannelMessageSend(channel.ID, "QuoteBot is ready! Type !quote")
				return
			}
		}
	}
}

func messageCreateEvent(bm botdata.Manager) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(session *discordgo.Session, message *discordgo.MessageCreate) {
		// Check to see if the author is this bot
		if message.Author.ID == session.State.User.ID {
			return
		}

		// check if its a command

		// parse arguments

		// construct result

		// send message
	}
}

func main() {
	// INITIALIZATION ---------------------------------------------------------
	// Store the application configuration
	botconfig := getConfig(configfile)

	// Starts the data manager for the bot
	botManager := botdata.Start(botconfig.DBuri)
	// defers the graceful shutdown of the data manager
	defer botManager.Shutdown()

	// Initialize the Discord Bot
	ds, err := discordgo.New("Bot " + botconfig.DiscordToken)
	checkError(err, "Error creating Discord Session: ")

	// EVENT HANDLING ---------------------------------------------------------
	// Define Handlers for discord events.
	messageCreate := messageCreateEvent(botManager)
	guildCreate := guildCreateEvent(botManager)

	// Attach Handlers to the discord session
	ds.AddHandler(ready)
	ds.AddHandler(messageCreate)
	ds.AddHandler(guildCreate)

	// START SESSION ----------------------------------------------------------
	// Open Discord session
	err = ds.Open()
	checkError(err, "Error opening Discord session: ")
	// Defers a call to Close the Discord Session
	defer ds.Close()

	// Start Message
	fmt.Println("Welcome to Quotebot X. Press CTRL+C to exit.")

	// Creates Signal Interupt Channels
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
