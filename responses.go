package main

import (
	"fmt"
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/bwmarrin/discordgo"
	"time"
)

func multiQuoteResponse(session *discordgo.Session, quotes []data.Quote) discordgo.InteractionResponse {
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

func quoteThisMessageResponse(_ data.Quote) discordgo.InteractionResponse {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Content: "This command is not implemented yet, please be patient",
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
