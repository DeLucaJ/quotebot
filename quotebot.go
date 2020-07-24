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

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "github.com/DeLucaJ/quotebot/internal/types"
)

const tfile string = "./data/token.txt"
const uri = "mongodb://localhost:27017"

// Retrieves the Bot Token
func getToken(file string) string {
	t, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading token")
	}
	return string(t)
}

func main() {

	// Store the application token
	var token string = getToken(tfile)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defers the context cancel
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	// Defer a disconnect from the database
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// Initialize the Discord Bot
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord Session: " + err.Error())
		return
	}
	defer ds.Close() // Defers a call to Close the Discord Session

	// Define Handlers for discord events.

	err = ds.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: " + err.Error())
		/* ds.Close()
		return */
	}

	fmt.Println("Welcome to Quotebot X. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
