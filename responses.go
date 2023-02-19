package main

import (
	"fmt"
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/bwmarrin/discordgo"
	"time"
)

func multiQuoteResponse(session *discordgo.Session, quotes []data.Quote) discordgo.InteractionResponse {
	if quotes[0].SpeakerID == 0 {
		return emptyResponse()
	}

	var quoteEmbeds = make([]*discordgo.MessageEmbed, len(quotes))

	for index, quote := range quotes {
		quoteEmbeds[index] = quoteToEmbed(session, quote)
	}

	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: quoteEmbeds,
		},
	}
}

func singleQuoteResponse(session *discordgo.Session, quote data.Quote) discordgo.InteractionResponse {
	if quote.SpeakerID == 0 {
		return emptyResponse()
	}

	quoteEmbeds := []*discordgo.MessageEmbed{
		quoteToEmbed(session, quote),
	}

	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: quoteEmbeds,
		},
	}
}

func emptyResponse() discordgo.InteractionResponse {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Sorry, there are no quotes matching your search",
		},
	}
}

func quoteToEmbed(session *discordgo.Session, quote data.Quote) *discordgo.MessageEmbed {
	footer := discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Submitted by %s", quote.Submitter.Name),
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
	return &embed
}
