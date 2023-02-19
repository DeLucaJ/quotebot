package main

import (
	"fmt"
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/DeLucaJ/quotebot/internal/migration"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

const maxAmount = 10
const minAmount = 1

func quoteSlashCommandHandler(manager data.Manager, session *discordgo.Session, icEvent *discordgo.InteractionCreate) {
	options := icEvent.ApplicationCommandData().Options

	switch options[0].Name {
	case "random":
		quoteRandomHandler(manager, session, icEvent.Interaction, options[0])
	case "add":
		quoteAddHandler(manager, session, icEvent.Interaction, options[0])
	case "by":
		quoteByHandler(manager, session, icEvent.Interaction, options[0])
	}
}

func quoteRandomHandler(manager data.Manager, session *discordgo.Session, interaction *discordgo.Interaction, optionData *discordgo.ApplicationCommandInteractionDataOption) {
	optionMap := makeOptionMap(optionData.Options)

	amount := minAmount

	if amountOption, ok := optionMap["amount"]; ok {
		amount = clampAmount(int(amountOption.IntValue()))
	}
	quotes := manager.GetNRandomQuotes(interaction.GuildID, amount)

	response := multiQuoteResponse(session, quotes)
	err := session.InteractionRespond(interaction, &response)
	if err != nil {
		log.Panicf("Unable to send response: %v", err)
	}
}

func quoteByHandler(manager data.Manager, session *discordgo.Session, interaction *discordgo.Interaction, optionData *discordgo.ApplicationCommandInteractionDataOption) {
	optionMap := makeOptionMap(optionData.Options)

	amount := minAmount
	var speaker *discordgo.User

	if speakerOption, ok := optionMap["speaker"]; ok {
		speaker = speakerOption.UserValue(session)
	}

	if amountOption, ok := optionMap["amount"]; ok {
		amount = clampAmount(int(amountOption.IntValue()))
	}

	quotes := manager.GetNRandomQuotesBySpeaker(speaker.ID, interaction.GuildID, amount)

	log.Println(quotes)

	response := multiQuoteResponse(session, quotes)

	err := session.InteractionRespond(interaction, &response)
	if err != nil {
		log.Panicf("Unable to send response: %v", err)
	}
}

func quoteAddHandler(manager data.Manager, session *discordgo.Session, interaction *discordgo.Interaction, optionData *discordgo.ApplicationCommandInteractionDataOption) {
	optionMap := makeOptionMap(optionData.Options)

	submitter := interaction.Member.User
	var speaker *discordgo.User
	var content string

	if speakerOption, ok := optionMap["speaker"]; ok {
		speaker = speakerOption.UserValue(session)
	}

	if contentOption, ok := optionMap["content"]; ok {
		content = strings.Trim(contentOption.StringValue(), " ")
	}

	quote := manager.AddQuote(content, speaker, submitter, interaction.GuildID)
	response := singleQuoteResponse(session, quote)
	err := session.InteractionRespond(interaction, &response)
	if err != nil {
		log.Panicf("Unable to send response: %v", err)
	}
}

func makeOptionMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, option := range options {
		optionMap[option.Name] = option
	}
	return optionMap
}

func clampAmount(amount int) int {
	if amount < minAmount {
		return minAmount
	} else if amount > maxAmount {
		return maxAmount
	} else {
		return amount
	}
}

func quoteThisCommandHandler(manager data.Manager, session *discordgo.Session, icEvent *discordgo.InteractionCreate) {
	messageID := icEvent.ApplicationCommandData().TargetID
	message, err := session.ChannelMessage(icEvent.ChannelID, messageID)
	if err != nil {
		log.Panicln("Failed to find message")
	}

	log.Println(message.Content)

	quote := manager.AddQuote(message.Content, message.Author, icEvent.Interaction.Member.User, icEvent.Interaction.GuildID)

	response := singleQuoteResponse(session, quote)

	err = session.InteractionRespond(icEvent.Interaction, &response)
	if err != nil {
		log.Panicf("Unable to send response: %v", err)
	}
}

var commandHandlers = map[string]func(manager data.Manager, session *discordgo.Session, icEvent *discordgo.InteractionCreate){
	quoteSlashCommands.Name:      quoteSlashCommandHandler,
	quoteThisMessageCommand.Name: quoteThisCommandHandler,
}

func interactionCreateHandler(manager data.Manager) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(session *discordgo.Session, icEvent *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[icEvent.ApplicationCommandData().Name]; ok {
			handler(manager, session, icEvent)
		}
	}
}

func ready(session *discordgo.Session, _ *discordgo.Ready) {
	err := session.UpdateGameStatus(0, "/quote")
	if err != nil {
		fmt.Println("Error updated Bot Status")
	}
}

func guildCreateHandler(manager data.Manager, commandMap map[string][]string) func(*discordgo.Session, *discordgo.GuildCreate) {
	return func(session *discordgo.Session, event *discordgo.GuildCreate) {
		if event.Guild.Unavailable {
			return
		}

		if !manager.GuildExists(event.Guild) {
			manager.AddGuild(event.Guild)
		}
		guild := manager.FindGuild(event.Guild.ID)

		fmt.Println("Login: ", guild.Name)

		for _, member := range event.Guild.Members {
			if manager.UserExists(member.User.ID, guild) {
				continue
			}
			manager.AddUser(member.User, guild)
		}

		commandMap[event.Guild.ID] = registerAllCommands(session, event.Guild.ID)

		migration.AttemptMigrateLegacyQuotes(manager, session, event)

		for _, channel := range event.Guild.Channels {
			if channel.ID == event.Guild.ID {
				_, _ = session.ChannelMessageSend(channel.ID, "QuoteBot is ready! Type /quote")
				break
			}
		}
	}
}

func guildUpdateHandler(manager data.Manager) func(*discordgo.Session, *discordgo.GuildUpdate) {
	return func(session *discordgo.Session, update *discordgo.GuildUpdate) {
		manager.UpdateGuild(update.Guild)
	}
}

func memberAddHandler(manager data.Manager) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(session *discordgo.Session, add *discordgo.GuildMemberAdd) {
		guild := manager.FindGuild(add.GuildID)

		if manager.UserExists(add.User.ID, guild) {
			return
		}

		manager.AddUser(add.User, guild)
	}
}

func memberUpdateHandler(manager data.Manager) func(*discordgo.Session, *discordgo.GuildMemberUpdate) {
	return func(session *discordgo.Session, update *discordgo.GuildMemberUpdate) {
		guild := manager.FindGuild(update.GuildID)

		manager.UpdateGuildUser(update.User, guild)
	}
}
