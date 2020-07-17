// Quotebot X
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const tfile string = "./data/token.txt"

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

	// Initialize the Bot
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord Session: " + err.Error())
		return
	}

	// Handlers for discord events.

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

	ds.Close()
}
