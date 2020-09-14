// Quotebot X
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "json"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "github.com/DeLucaJ/quotebot/internal/data"
)

const tfile string = "./data/token.txt"
const uri = "mongodb://localhost:27017"

// Used for general error checking and panicing
func checkError(err error, message string) {
	if err != nil {
		fmt.Println(message + err.Error())
		log.Fatal(err)
	}
}

// Retrieves the Bot Token
// Will eventually get the JSON configuration for quotebot
func getConfig(file string) string {
	t, err := ioutil.ReadFile(file)
	checkError(err, "Error reading discord token: ")
	return string(t)
}

func main() {

	// Store the application token
	var token string = getConfig(tfile)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defers the context cancel
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	checkError(err, "Error Conecting Mongo Client: ")

	// Defer a disconnect from the database
	defer func() {
		checkError(client.Disconnect(ctx), "Error disconnecting from mongo client: ")
	}()

	// Initialize the Database
	/* database := client.Database("quotedb")
	guildCollection := database.Collection("guilds")
	userCollection := database.Collection("users")
	quoteCollection := database.Collection("quotes") */

	// Initialize the Discord Bot
	ds, err := discordgo.New("Bot " + token)
	checkError(err, "Error creating Discord Session: ")

	// Define Handlers for discord events.

	// Open Discord session
	err = ds.Open()
	checkError(err, "Error opening Discord session: ")
	// Defers a call to Close the Discord Session
	defer ds.Close()

	// Creates Signal Interupt Channels
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Start Message
	fmt.Println("Welcome to Quotebot X. Press CTRL+C to exit.")
}
