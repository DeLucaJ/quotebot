package botdata

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
func (bm Manager) AddUser(name string, user discordgo.User, guild discordgo.Guild) {
	var guildEntry Guild
	bm.Database.Where(&Guild{DiscordID: guild.ID}).Find(&guildEntry)

	userEntry := User{
		Name:      name,
		DiscordID: user.ID,
		GuildID:   guildEntry.ID,
	}

	bm.insertUser(userEntry)
}

// AddQuote - adds a Quote to the Database
func (bm Manager) AddQuote(content string, speaker discordgo.User, submitter discordgo.User, guild discordgo.Guild) {
	var guildEntry Guild
	var speakerEntry User
	var submitterEntry User

	if !bm.GuildExists(guild) {
		fmt.Println("error: attempt to add quote to non-existent guild")
	}
	bm.Database.Where(&Guild{DiscordID: guild.ID}).Find(&guildEntry)

	// Check if speaker user exists, if not make that user
	if !bm.UserExists(speaker, guildEntry) {
		bm.AddUser(speaker.Username, speaker, guild)
	}

	// Check if submitter user exists, if not make that user
	if !bm.UserExists(submitter, guildEntry) {
		bm.AddUser(submitter.Username, submitter, guild)
	}

	quote := Quote{
		Content:   content,
		Speaker:   speakerEntry.ID,
		Submitter: submitterEntry.ID,
		Guild:     guildEntry.ID,
	}

	bm.insertQuote(quote)
}

// ChooseRandomQuote - Chooses a random quote from a specific guild
func (bm Manager) ChooseRandomQuote(guild discordgo.Guild) Quote {
	guildEntry := bm.findGuild(guild)
	quotes := bm.findManyQuotes(&Quote{Guild: guildEntry.ID})

	return bm.randomQuote(quotes)
}

// ChooseRandomQuoteBySpeaker - Chooses a random quote from guild and speaker
func (bm Manager) ChooseRandomQuoteBySpeaker(speaker discordgo.User, guild discordgo.Guild) Quote {
	guildEntry := bm.findGuild(guild)
	speakerEntry := bm.findUser(speaker)
	quotes := bm.findManyQuotes(&Quote{Speaker: speakerEntry.ID, Guild: guildEntry.ID})

	return bm.randomQuote(quotes)
}

// helper for random qutoes
func (bm Manager) randomQuote(quotes []Quote) Quote {
	if len(quotes) > 0 {
		return quotes[rand.Intn(len(quotes))]
	}

	//returns empty quote to be checked
	return Quote{}
}

// GuildExists - returns true if the user exists, false otherwise
func (bm Manager) GuildExists(guild discordgo.Guild) bool {
	var existing Guild
	result := bm.Database.Where(&Guild{DiscordID: guild.ID}).First(&existing)
	if result.Error != nil {
		fmt.Println("Error checking for guild existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

// UserExists - returns true if the user exists, false otherwise
func (bm Manager) UserExists(user discordgo.User, guild Guild) bool {
	var existing User
	result := bm.Database.Where(&User{GuildID: guild.ID, DiscordID: user.ID}).First(&existing)
	if result.Error != nil {
		fmt.Println("Error checking for user existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

// adds a Guild to the database
// move find and construction logic into separate event code or constructor
func (bm Manager) insertGuild(guild Guild) {
	result := bm.Database.Create(&guild)
	if result.Error != nil {
		fmt.Println("Error inserting guild: ", result.Error)
		panic(result.Error)
	}

	fmt.Println("Guild Add: ", guild.Name, result.Name())
}

// adds a User to the database
// move find & construction logic into separate event code or constructor
func (bm Manager) insertUser(user User) {
	// insert document into Database
	result := bm.Database.Create(&user)
	if result.Error != nil {
		fmt.Println("Error inserting user: ", result.Error)
	}
	fmt.Println("User Added: ", user.Name, result.Name())
}

// adds a Quote to the database
// move find calls and construction into the event code or separate constructor
func (bm Manager) insertQuote(quote Quote) {
	//insert quote into DB
	result := bm.Database.Create(&quote)
	if result.Error != nil {
		fmt.Println("Error inserting quote: ", result.Error)
	}
	fmt.Println("Quote Added: ", result.Name())
}

func (bm Manager) findGuild(guild discordgo.Guild) Guild {
	var guildEntry Guild
	err := bm.Database.Where(&Guild{DiscordID: guild.ID}).First(&guildEntry)
	if err != nil {
		fmt.Println("Error retrieving guild: ", err)
	}
	return guildEntry
}

func (bm Manager) findUser(user discordgo.User) User {
	var userEntry User
	bm.Database.Where(&User{DiscordID: user.ID}).First(&userEntry)
	return userEntry
}

func (bm Manager) findQuote(query *Quote) Quote {
	var quoteEntry Quote
	bm.Database.Where(&query).First(&quoteEntry)
	return quoteEntry
}

func (bm Manager) findManyQuotes(query *Quote) []Quote {
	var quotes []Quote

	bm.Database.Where(&query).Find(&quotes)

	return quotes
}

/* // flags a quote for inspection by administrator
func (bm Manager) flagQutoe() {

}

// sets a different alias for a user in the Database
func (bm Manager) renameUser() {

}

// converts all quotes by an unassociated user to this user
func (bm Manager) claimUser() {

} */
