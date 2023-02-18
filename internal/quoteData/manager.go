package quoteData

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Manager - struct that holds the data necessary to manage the Database
type Manager struct {
	Context    context.Context
	CancelFunc context.CancelFunc
	Database   *gorm.DB
}

// Start - Starts the boss and initializes the Database connection
//
//	uri string: the uri for the Database from config.json
func Start(dsn string) Manager {
	fmt.Println("Initializing Postgresql Client")

	//// Connect to Postgres
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defers the context CancelFunc

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		defer cancel()
		fmt.Println("Error connecting to Postgres Client: " + err.Error())
	}

	// Initialize the Database
	err = db.AutoMigrate(&Guild{}, &User{}, &Quote{})
	if err != nil {
		fmt.Println("Error creating Database: " + err.Error())
	}

	// initializes the singleton Manager
	return Manager{
		Context:    ctx,
		CancelFunc: cancel,
		Database:   db,
	}
}

// Shutdown - Ends the connection to the Database and cleans the context
func (bm Manager) Shutdown() {
	cancel := bm.CancelFunc
	defer cancel()
}

// AddGuild - adds a guild to the Database
func (bm Manager) AddGuild(guild discordgo.Guild) {
	guildEntry := Guild{
		DiscordID: guild.ID,
		Name:      guild.Name,
	}
	bm.Database.Create(&guildEntry)
}

// AddUser - adds a user to the Database
func (bm Manager) AddUser(user discordgo.User, guild Guild) {
	userEntry := User{
		Name:      user.Username,
		DiscordID: user.ID,
		GuildID:   guild.ID,
	}

	bm.InsertUser(userEntry)
}

// AddQuote - adds a Quote to the Database
func (bm Manager) AddQuote(content string, speaker discordgo.User, submitter discordgo.User, guildID string) Quote {
	guildEntry := bm.FindGuild(guildID)

	// Check if speaker user exists, if not make that user
	if !bm.UserExists(speaker.ID, guildEntry) {
		bm.AddUser(speaker, guildEntry)
	}
	speakerEntry := bm.FindUser(speaker.ID, guildEntry.ID)

	// Check if submitter user exists, if not make that user
	if !bm.UserExists(submitter.ID, guildEntry) {
		bm.AddUser(submitter, guildEntry)
	}
	submitterEntry := bm.FindUser(submitter.ID, guildEntry.ID)

	quote := Quote{
		Content:     content,
		SpeakerID:   speakerEntry.ID,
		Speaker:     speakerEntry,
		SubmitterID: submitterEntry.ID,
		Submitter:   submitterEntry,
		GuildID:     guildEntry.ID,
	}

	bm.InsertQuote(quote)

	// Temporary Timestamp to account for the add event bug
	quote.CreatedAt = time.Now()

	return quote
}

// ChooseRandomQuote - Chooses a random quote from a specific guild
func (bm Manager) ChooseRandomQuote(guildID string) Quote {
	guildEntry := bm.FindGuild(guildID)

	return bm.randomQuote(guildEntry.Quotes)
}

// ChooseRandomQuoteBySpeaker - Chooses a random quote from guild and speaker
func (bm Manager) ChooseRandomQuoteBySpeaker(speakerID string, guildID string) Quote {
	guildEntry := bm.FindGuild(guildID)
	speakerEntry := bm.FindUser(speakerID, guildEntry.ID)
	quotes := bm.FindManyQuotes(&Quote{SpeakerID: speakerEntry.ID, GuildID: guildEntry.ID})

	return bm.randomQuote(quotes)
}

// helper for random quotes
func (bm Manager) randomQuote(quotes []Quote) Quote {
	if len(quotes) > 0 {
		return quotes[rand.Intn(len(quotes))]
	}

	//returns empty quote to be checked
	return Quote{}
}

func (bm Manager) GuildExists(guild discordgo.Guild) bool {
	return bm.GuildExistsByID(guild.ID)
}

// GuildExistsByID - returns true if the user exists, false otherwise
func (bm Manager) GuildExistsByID(guildID string) bool {
	var existing Guild
	result := bm.Database.Where(&Guild{DiscordID: guildID}).First(&existing)
	if result.Error != nil {
		fmt.Println("Error checking for guild existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

// UserExists - returns true if the user exists, false otherwise
func (bm Manager) UserExists(userID string, guild Guild) bool {
	var existing User
	result := bm.Database.Where(&User{GuildID: guild.ID, DiscordID: userID}).First(&existing)
	if result.Error != nil {
		fmt.Println("Error checking for user existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

// InsertGuild adds a Guild to the database
// move find and construction logic into separate event code or constructor
func (bm Manager) InsertGuild(guild Guild) {
	result := bm.Database.Create(&guild)
	if result.Error != nil {
		fmt.Println("Error inserting guild: ", result.Error)
		panic(result.Error)
	}

	fmt.Println("Guild Add: ", guild.Name, result.Name())
}

// InsertUser adds a User to the database
// move find & construction logic into separate event code or constructor
func (bm Manager) InsertUser(user User) {
	// insert document into Database
	result := bm.Database.Create(&user)
	if result.Error != nil {
		fmt.Println("Error inserting user: ", result.Error)
	}
	fmt.Println("User Added: ", user.Name, result.Name())
}

// InsertQuote adds a Quote to the database
// move find calls and construction into the event code or separate constructor
func (bm Manager) InsertQuote(quote Quote) {
	//insert quote into DB
	result := bm.Database.Create(&quote)
	if result.Error != nil {
		fmt.Println("Error inserting quote: ", result.Error)
	}
	fmt.Println("Quote Added: ", result.Name())
}

func (bm Manager) FindGuild(guildID string) Guild {
	var guildEntry Guild
	result := bm.Database.
		Where(&Guild{DiscordID: guildID}).
		Preload(clause.Associations).
		Preload("Quotes.Speaker").
		Preload("Quotes.Submitter").
		First(&guildEntry)

	if result.Error != nil {
		fmt.Println(fmt.Sprintf("Error retrieving guild of ID %s: %s", guildID, result.Error))
	}
	return guildEntry
}

func (bm Manager) FindUser(userID string, guildID uint) User {
	var userEntry User
	result := bm.Database.Where(&User{DiscordID: userID, GuildID: guildID}).First(&userEntry)
	if result.Error != nil {
		fmt.Println(fmt.Sprintf("Error retrieving user of ID %s: %s", userID, result.Error))
	}
	return userEntry
}

func (bm Manager) FindQuote(query *Quote) Quote {
	var quoteEntry Quote
	result := bm.Database.
		Where(&query).
		Preload(clause.Associations).
		First(&quoteEntry)

	if result.Error != nil {
		fmt.Println("Error retrieving quote", result.Error)
	}
	return quoteEntry
}

func (bm Manager) FindManyQuotes(query *Quote) []Quote {
	var quotes []Quote

	result := bm.Database.Where(&query).
		Preload(clause.Associations).
		Find(&quotes)

	if result.Error != nil {
		fmt.Println("Error retrieving quotes", result.Error)
	}

	return quotes
}

/* // flags a quote for inspection by administrator
func (bm Manager) flagQuote() {

}

// sets a different alias for a user in the Database
func (bm Manager) renameUser() {

}

// converts all quotes by an unassociated user to this user
func (bm Manager) claimUser() {

} */
