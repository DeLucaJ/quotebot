// Quotebot X
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DeLucaJ/quotebot/internal/botdata"
	"github.com/bwmarrin/discordgo"
)

const configfile string = "./config.json"

type config struct {
	Token string
	DBuri string
}

// Used for general error checking and panicing
func checkError(err error, message string) {
	if err != nil {
		fmt.Println(message + err.Error())
		log.Fatal(err)
	}
}

// Retrieves the Bot Token
// Will eventually get the JSON configuration for quotebot
func getConfig(file string) config {
	// read the fildata of configfile
	filedata, err := ioutil.ReadFile(file)
	checkError(err, "Error reading discord token: ")

	//turn config.json into a config struct
	var configuration config
	err = json.Unmarshal(filedata, &configuration)
	if err != nil {
		fmt.Println("Error unmarshalling json: ", err)
	}

	return configuration
}

func main() {
	// Store the application configuration
	botconfig := getConfig(configfile)

	// Starts the data manager for the bot
	dataManager := botdata.Start(botconfig.DBuri)
	// defers the graceful shutdown of the data manager
	defer dataManager.Shutdown()

	// Initialize the Discord Bot
	ds, err := discordgo.New("Bot " + botconfig.Token)
	checkError(err, "Error creating Discord Session: ")

	// EVENT HANDLING ---------------------------------------------------------
	// Define Handlers for discord events.
	messageCreate := dataManager.MessageCreate
	guildCreate := dataManager.GuildCreate

	// Attach Handlers to the discord session
	ds.AddHandler(messageCreate)
	ds.AddHandler(guildCreate)
	//-------------------------------------------------------------------------

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
