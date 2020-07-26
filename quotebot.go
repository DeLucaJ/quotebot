// Quotebot X
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "json"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "go.mongodb.org/mongo-driver/bson"
	// "github.com/DeLucaJ/quotebot/internal/types"
)

const tfile string = "./data/token.txt"
const uri = "mongodb://localhost:27017"

// Retrieves the Bot Token
func getConfig(file string) string {
	t, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading token")
	}
	return string(t)
}

func checkError(err error, message string) {
	if err != nil {
		fmt.Println(message + err.Error())
		panic(err)
	}
}

func main() {

	// Store the application token
	var token string = getConfig(tfile)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defers the context cancel
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {

	}

	// Defer a disconnect from the database
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

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
