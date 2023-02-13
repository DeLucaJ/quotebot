// QuoteBot X by Joseph DeLuca
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/DeLucaJ/quotebot/internal/botdata"
	"github.com/bwmarrin/discordgo"
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

func checkErrorBenign(err error, message string) {
	if err != nil {
		fmt.Println(message + err.Error())
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

func quoteToMessage(session *discordgo.Session, quote botdata.Quote) discordgo.MessageSend {
	footer := discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Submitted by: %s", quote.Submitter.Name),
	}

	speakerInfo, _ := session.User(quote.Speaker.DiscordID)

	thumbnail := discordgo.MessageEmbedThumbnail{
		URL: speakerInfo.AvatarURL(""),
	}

	embed := discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Color:       speakerInfo.AccentColor,
		Title:       fmt.Sprintf("%s", quote.Speaker.Name),
		Description: fmt.Sprintf("\"%s\"", quote.Content),
		Footer:      &footer,
		Thumbnail:   &thumbnail,
		Timestamp:   quote.CreatedAt.Format(time.RFC3339),
	}

	messagePackage := discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{&embed},
	}
	return messagePackage
}

func quoteToMessageWithContent(session *discordgo.Session, quote botdata.Quote, content string) discordgo.MessageSend {
	quoteEmbed := quoteToMessage(session, quote)
	quoteEmbed.Content = content
	return quoteEmbed
}

func sendQuoteMessage(channelID string, session *discordgo.Session, quote botdata.Quote) {
	channel, err := session.State.Channel(channelID)
	checkError(err, "Error accessing Channel: ")
	// Replace this with message send complex
	quoteEmbed := quoteToMessage(session, quote)
	_, err = session.ChannelMessageSendComplex(channel.ID, &quoteEmbed)
	checkErrorBenign(err, "Error sending quote message: ")
}

func sendRandomQuote(bm botdata.Manager, session *discordgo.Session, message *discordgo.MessageCreate) {
	quote := bm.ChooseRandomQuote(message.GuildID)
	sendQuoteMessage(message.ChannelID, session, quote)
}

func sendRandomQuoteByUser(bm botdata.Manager, session *discordgo.Session, message *discordgo.MessageCreate) {
	quote := bm.ChooseRandomQuoteBySpeaker(message.Mentions[0].ID, message.GuildID)
	sendQuoteMessage(message.ChannelID, session, quote)
}

func addQuote(bm botdata.Manager, commandConfig CommandConfig, arguments []string, session *discordgo.Session, message *discordgo.MessageCreate) {
	channel, err := session.State.Channel(message.ChannelID)
	checkError(err, "Error accessing Channel: ")

	if len(arguments) < 2 {
		delivery := fmt.Sprintf("Missing arguments for Add.\n Use command \"%s%s <Mention User> <Quote>\".",
			commandConfig.Prefix,
			commandConfig.Add)
		_, err = session.ChannelMessageSend(channel.ID, delivery)
		checkErrorBenign(err, "Error sending add quote missing arg message: ")
		return
	}

	speaker := message.Mentions[0]
	submitter := message.Author
	content := strings.Trim(strings.Join(arguments[1:], " "), `'"`)
	quote := bm.AddQuote(content, *speaker, *submitter, channel.GuildID)

	quoteEmbed := quoteToMessageWithContent(session, quote, "Thank you for the new quote!")

	_, err = session.ChannelMessageSendComplex(channel.ID, &quoteEmbed)
	checkErrorBenign(err, "Error sending add quote message: ")
}

func ready(session *discordgo.Session, _ *discordgo.Ready) {
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
				_, _ = session.ChannelMessageSend(channel.ID, "QuoteBot is ready! Type q!")
				return
			}
		}
	}
}

func messageCreateEvent(bm botdata.Manager, commandConfig CommandConfig) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(session *discordgo.Session, message *discordgo.MessageCreate) {
		// Check to see if the author is this bot
		if message.Author.ID == session.State.User.ID {
			return
		}

		// check if it's a command
		if strings.HasPrefix(message.Content, commandConfig.Prefix) {
			// Get the current discord Channel
			channel, err := session.State.Channel(message.ChannelID)
			checkError(err, "Error accessing Channel: ")

			// remove the prefix and delineate the string into separate words
			content := strings.TrimPrefix(message.Content, commandConfig.Prefix)
			arguments := strings.Split(content, " ")

			// if there are arguments, parse them
			if len(arguments) > 0 {
				switch arguments[0] {
				case commandConfig.Add:
					addQuote(bm, commandConfig, arguments[1:], session, message)
					break
				case commandConfig.By:
					sendRandomQuoteByUser(bm, session, message)
					break
				default:
					sendRandomQuote(bm, session, message)
				}
			} else {
				sendRandomQuote(bm, session, message)
			}
			err = session.ChannelMessageDelete(channel.ID, message.ID)
			checkErrorBenign(err, "Error deleting message: ")
			return
		}

	}
}

func main() {
	// INITIALIZATION ---------------------------------------------------------
	// Store the application configuration
	botConfig := getConfig(configFile)

	// Starts the data manager for the bot
	botManager := botdata.Start(botConfig.ConnectionString)
	// defers the graceful shutdown of the data manager
	defer botManager.Shutdown()

	// Initialize the Discord Bot
	ds, err := discordgo.New("Bot " + botConfig.DiscordToken)
	checkError(err, "Error creating Discord Session: ")

	// EVENT HANDLING ---------------------------------------------------------
	// Define Handlers for discord events.
	messageCreate := messageCreateEvent(botManager, botConfig.CommandConfig)
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
	defer func(ds *discordgo.Session) {
		err := ds.Close()
		checkError(err, "Error closing Discord session: ")
	}(ds)

	// Start Message
	fmt.Println("Welcome to QuoteBot X. Press CTRL+C to exit.")

	// Creates Signal Interrupt Channels
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
