package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Quote - a simple data strucutre for quote storage
type Quote struct {
	Speaker string
	Text    string
}

const token string = "filler"
const qfile string = "./data/quotes.json"

var qslice []Quote
var qmap map[string][]string

func load() {
	qjson, err := ioutil.ReadFile(qfile)
	if err != nil {
		fmt.Println(err.Error())
		qslice = make([]Quote, 0)
		return
	}

	err = json.Unmarshal(qjson, &qslice)
	if err != nil {
		fmt.Println(err.Error())
		qslice = make([]Quote, 0)
		return
	}

	qmap = make(map[string][]string)
	for _, quote := range qslice {
		_, present := qmap[quote.Speaker]
		if !present {
			qmap[quote.Speaker] = make([]string, 0)
		}
		qmap[quote.Speaker] = append(qmap[quote.Speaker], quote.Text)
	}
}

func save() {
	qjson, err := json.MarshalIndent(qslice, "", "\t")
	if err != nil {
		fmt.Println("Error when encoding quote list: " + err.Error())
		return
	}

	err = ioutil.WriteFile(qfile, qjson, 0644)
	if err != nil {
		fmt.Println("Error when saving quote list: " + err.Error())
	}
}

func add(quote *Quote) {
	qslice = append(qslice, *quote)

	_, present := qmap[quote.Speaker]
	if !present {
		qmap[quote.Speaker] = make([]string, 0)
	}
	qmap[quote.Speaker] = append(qmap[quote.Speaker], quote.Text)

	save()
}

func createMap() map[string]string {
	return make(map[string]string)
}

func chooseByName(name string) string {
	ql, ok := qmap[name]
	if !ok {
		return ""
	}

	return ql[rand.Intn(len(ql))]
}

func choose() (string, string) {
	q := qslice[rand.Intn(len(qslice))]
	return q.Speaker, q.Text
}

func ready(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateStatus(0, "!quote")
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	if strings.HasPrefix(message.Content, "!quote") {

		channel, err := session.State.Channel(message.ChannelID)
		if err != nil {
			fmt.Println("Error accessing Channel: " + message.ChannelID)
			return
		}

		args := strings.Split(message.Content, " ")[1:]

		delivery := ""
		fmt.Println(args)

		// reoganize
		if len(args) > 0 {
			switch args[0] {
			case "add":
				if len(args) < 3 {
					delivery = "Missing Arguments for Add.\nUse command \"!quote add <Name> <Quote>\"."
					break
				}
				nquote := &Quote{Speaker: args[1], Text: strings.Join(args[2:], " ")}
				add(nquote)
				delivery = fmt.Sprintf("Added Quote:\n\t%s: \"%s\"", nquote.Speaker, nquote.Text)
			case "by":
				if len(args) != 2 {
					delivery = "Invalid args for Quote By.\nUse command \"!quote by <Name>\"."
					break
				}
				speaker, text := args[1], chooseByName(args[1])
				delivery = fmt.Sprintf("%s: \"%s\"", speaker, text)
			default:
				speaker, text := choose()
				delivery = fmt.Sprintf("%s: \"%s\"", speaker, text)
			}
		} else {
			speaker, text := choose()
			delivery = fmt.Sprintf("%s: \"%s\"", speaker, text)
		}

		session.ChannelMessageDelete(channel.ID, message.ID)
		session.ChannelMessageSend(channel.ID, delivery)
		fmt.Printf("#%s (%s)\n", channel.Name, delivery)
		return
	}
}

func guildCreate(session *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = session.ChannelMessageSend(channel.ID, "QuoteBot is ready! Type !quote")
			return
		}
	}
}

func main() {

	load()

	// initialize bot
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord Session: " + err.Error())
		return
	}

	// ready - called when discord sends ready event
	ds.AddHandler(ready)

	// messageCreate - everytime a message is sent in a channel the bot can access
	ds.AddHandler(messageCreate)

	// guildCreate - everytime a serer is joined
	ds.AddHandler(guildCreate)

	err = ds.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err.Error())
	}

	fmt.Println("QuoteBot is now running. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	ds.Close()
}
