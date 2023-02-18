package main

import (
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

const maxAmount = 20
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

	quote := manager.AddQuote(content, *speaker, *submitter, interaction.GuildID)
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

func quoteThisCommandHandler(_ data.Manager, session *discordgo.Session, icEvent *discordgo.InteractionCreate) {
	// construct new quote from event data
	quote := data.Quote{}

	// filler
	response := quoteThisMessageResponse(quote)

	err := session.InteractionRespond(icEvent.Interaction, &response)
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
