package migration

import (
	"encoding/json"
	"github.com/DeLucaJ/quotebot/internal/data"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
)

const migrateMapFile string = "./legacyData/migrate-data.json"
const legacyQuotesFile string = "./legacyData/quotes.json"

type MigrateData struct {
	Done        bool      `json:"done"`
	BotUserName string    `json:"bot-user-name"`
	GuildName   string    `json:"guild-name"`
	UserMap     []NameMap `json:"user-map"`
}

type NameMap struct {
	UserName string   `json:"user-name"`
	OldNames []string `json:"old-names"`
}

type LegacyQuote struct {
	Speaker string `json:"Speaker"`
	Text    string `json:"Text"`
}

func itsNotTime(guildName string) bool {
	return guildName != migrateData.GuildName
}

var migrateData MigrateData

func legacyToModern(manager data.Manager, migrateMap map[string]string, legacyQuote LegacyQuote, guild data.Guild) {
	if !manager.UserExistsByName(migrateMap[legacyQuote.Speaker], guild) {
		return
	}
	if len(legacyQuote.Text) == 0 {
		return
	}

	speaker := manager.FindUserByName(migrateMap[legacyQuote.Speaker], guild.ID)
	submitter := manager.FindUserByName(migrateData.BotUserName, guild.ID)

	manager.AddLegacyQuote(legacyQuote.Text, speaker, submitter, guild)
}

func init() {
	migrateDataRaw, err := os.ReadFile(migrateMapFile)
	if err != nil {
		log.Panicf("Failled to read migrate data file")
	}

	err = json.Unmarshal(migrateDataRaw, &migrateData)
	if err != nil {
		log.Panicf("Failed to unmarshal migrate map")
	}
}

func AttemptMigrateLegacyQuotes(manager data.Manager, session *discordgo.Session, event *discordgo.GuildCreate) {
	if itsNotTime(event.Guild.Name) {
		return
	}

	if migrateData.Done {
		return
	}

	legacyQuotesRaw, err := os.ReadFile(legacyQuotesFile)
	if err != nil {
		log.Panicf("Failed to read legacy quotes file")
	}

	var migrateMap map[string]string = make(map[string]string)
	for _, entry := range migrateData.UserMap {
		for _, oldName := range entry.OldNames {
			migrateMap[oldName] = entry.UserName
		}
	}

	// get old quotes
	var legacyQuotes []LegacyQuote
	err = json.Unmarshal(legacyQuotesRaw, &legacyQuotes)
	if err != nil {
		log.Panicf("Failed to unmarshal legacy quotes")
	}

	dataGuild := manager.FindGuild(event.Guild.ID)

	// loop through all old quotes and process them
	for _, legacyQuote := range legacyQuotes {
		if _, ok := migrateMap[legacyQuote.Speaker]; ok {
			legacyToModern(manager, migrateMap, legacyQuote, dataGuild)
		}
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = session.ChannelMessageSend(channel.ID, "Legacy quotes have been migrated to QuoteBotX!")
			break
		}
	}

	migrateData.Done = true

	migrateDataJson, err := json.MarshalIndent(migrateData, "", "\t")
	if err != nil {
		log.Printf("Error marhsalling migrate data JSON")
	}

	err = os.WriteFile(migrateMapFile, migrateDataJson, 0644)
	if err != nil {
		log.Printf("Error writing migrate data JSON to file")
	}
}
