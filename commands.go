package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

var quoteRandom = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "random",
	Description: "send a random quote",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "amount",
			Description: "the number of quotes to send (max 20)",
		},
	},
}

var quoteAdd = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "add",
	Description: "adds a quote to the quote database",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "speaker",
			Description: "the server user that spoke the quote",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "content",
			Description: "the content of the quote",
			Required:    true,
		},
	},
}

var quoteBy = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionSubCommand,
	Name:        "by",
	Description: "sends a random quote by the given user",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "speaker",
			Description: "the speaker of the quote",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "amount",
			Description: "the number of quotes to send (max 20)",
		},
	},
}

var quoteSlashCommands = discordgo.ApplicationCommand{
	Type:        discordgo.ChatApplicationCommand,
	Name:        "quote",
	Description: "A command for managing and displaying quotes from QuoteBot",
	Options: []*discordgo.ApplicationCommandOption{
		&quoteRandom,
		&quoteAdd,
		&quoteBy,
	},
}

var quoteThisMessageCommand = discordgo.ApplicationCommand{
	Type:        discordgo.MessageApplicationCommand,
	Name:        "quote-this",
	Description: "Turns the selected message into a quote",
}

var allCommands = []*discordgo.ApplicationCommand{
	&quoteSlashCommands,
	&quoteThisMessageCommand,
}

var commandHandlers = map[string]func(session *discordgo.Session, icEvent *discordgo.InteractionCreate){}

func registerAllCommands(session *discordgo.Session, guildID string) []string {
	log.Println("Registering commands...")
	registeredCommandIDs := make([]string, len(allCommands))
	for index, command := range allCommands {
		command, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, command)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v\n", command.Name, err)
		}
		registeredCommandIDs[index] = command.ID
	}
	return registeredCommandIDs
}

// Need to be able to cache registered command IDs by server
func removeAllCommands(session *discordgo.Session, guildID string, registeredCommandIDs []string) {
	log.Println("Removing commands...")
	for _, commandID := range registeredCommandIDs {
		err := session.ApplicationCommandDelete(session.State.User.ID, guildID, commandID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", commandID, err)
		}
	}
}

func interactionHandler(session *discordgo.Session, icEvent *discordgo.InteractionCreate) {
	if handler, ok := commandHandlers[icEvent.ApplicationCommandData().Name]; ok {
		handler(session, icEvent)
	}
}
