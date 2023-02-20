package main

import (
	"fmt"
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/bwmarrin/discordgo"
	"time"
)

func getQuotesResponse(session *discordgo.Session, quotes []data.Quote) discordgo.InteractionResponse {
	if quotes[0].SpeakerID == 0 {
		return emptyResponse("Sorry, there are no quotes matching your search")
	} else {
		return multiQuoteResponse(session, quotes)
	}
}

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

func addQuoteResponse(session *discordgo.Session, quote data.Quote) discordgo.InteractionResponse {
	if quote.SpeakerID == 0 {
		return emptyResponse(quote.Content)
	} else {
		return singleQuoteResponse(session, quote)
	}
}

func singleQuoteResponse(session *discordgo.Session, quote data.Quote) discordgo.InteractionResponse {
	if quote.SpeakerID == 0 {
		return emptyResponse("Sorry, there are no quotes matching your search")
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

func emptyResponse(content string) discordgo.InteractionResponse {
	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	}
}

func quoteToEmbed(session *discordgo.Session, quote data.Quote) *discordgo.MessageEmbed {
	footer := discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Submitted by %s", quote.Submitter.Name),
	}

	speakerInfo, _ := session.GuildMember(quote.Guild.DiscordID, quote.Speaker.DiscordID)

	thumbnail := discordgo.MessageEmbedThumbnail{
		URL: speakerInfo.User.AvatarURL(""),
	}

	embed := discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Color:       speakerInfo.User.AccentColor,
		Title:       fmt.Sprintf("%s", quote.Speaker.Name),
		Description: fmt.Sprintf("\"%s\"", quote.Content),
		Footer:      &footer,
		Thumbnail:   &thumbnail,
		Timestamp:   quote.CreatedAt.Format(time.RFC3339),
	}
	return &embed
}
