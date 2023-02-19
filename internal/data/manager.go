package data

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
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
	log.Println("Initializing Postgresql Client")

	//// Connect to Postgres
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defers the context CancelFunc

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		defer cancel()
		log.Println("Error connecting to Postgres Client: " + err.Error())
	}

	// Initialize the Database
	err = db.AutoMigrate(&Guild{}, &User{}, &Quote{})
	if err != nil {
		log.Println("Error creating Database: " + err.Error())
	}

	// initializes the singleton Manager
	return Manager{
		Context:    ctx,
		CancelFunc: cancel,
		Database:   db,
	}
}

// Shutdown - Ends the connection to the Database and cleans the context
func (manager Manager) Shutdown() {
	cancel := manager.CancelFunc
	defer cancel()
}

// AddGuild - adds a guild to the Database
func (manager Manager) AddGuild(guild *discordgo.Guild) {
	guildEntry := Guild{
		DiscordID: guild.ID,
		Name:      guild.Name,
	}
	manager.Database.Create(&guildEntry)
}

// AddUser - adds a user to the Database
func (manager Manager) AddUser(user *discordgo.User, guild Guild) {
	userEntry := User{
		Name:      user.Username,
		DiscordID: user.ID,
		GuildID:   guild.ID,
	}

	manager.insertUser(userEntry)
}

// AddQuote - adds a Quote to the Database
func (manager Manager) AddQuote(content string, speaker *discordgo.User, submitter *discordgo.User, guildID string) Quote {
	if len(content) == 0 {
		return Quote{
			Content: "Sorry, but I can't accept empty quotes or quotes with only embedded content",
		}
	}

	guildEntry := manager.FindGuild(guildID)

	// Check if speaker user exists, if not make that user
	if !manager.UserExists(speaker.ID, guildEntry) {
		manager.AddUser(speaker, guildEntry)
	}
	speakerEntry := manager.FindUser(speaker.ID, guildEntry.ID)

	if manager.QuoteExists(Quote{Content: content, SpeakerID: speakerEntry.ID}) {
		return Quote{
			Content: "Sorry, but a quote with that content already exists for this user",
		}
	}

	// Check if submitter user exists, if not make that user
	if !manager.UserExists(submitter.ID, guildEntry) {
		manager.AddUser(submitter, guildEntry)
	}
	submitterEntry := manager.FindUser(submitter.ID, guildEntry.ID)

	quote := Quote{
		Content:     content,
		SpeakerID:   speakerEntry.ID,
		Speaker:     speakerEntry,
		SubmitterID: submitterEntry.ID,
		Submitter:   submitterEntry,
		GuildID:     guildEntry.ID,
	}

	manager.insertQuote(quote)

	// Temporary Timestamp to account for the add event bug
	quote.CreatedAt = time.Now()

	return quote
}

func (manager Manager) AddLegacyQuote(content string, speaker User, submitter User, guild Guild) Quote {
	if len(content) == 0 {
		return Quote{
			Content: "Sorry, but I can't accept empty quotes or quotes with only embedded content",
		}
	}

	if manager.QuoteExists(Quote{Content: content, SpeakerID: speaker.ID}) {
		return Quote{
			Content: "Sorry, but a quote with that content already exists for this user",
		}
	}

	quote := Quote{
		Content:     content,
		SpeakerID:   speaker.ID,
		Speaker:     speaker,
		SubmitterID: submitter.ID,
		Submitter:   submitter,
		GuildID:     guild.ID,
		Guild:       guild,
	}
	manager.insertQuote(quote)
	return quote
}

// GetRandomQuote - Chooses a random quote from a specific guild
func (manager Manager) GetRandomQuote(guildID string) Quote {
	guildEntry := manager.FindGuild(guildID)

	return manager.chooseQuoteRandomly(guildEntry.Quotes)
}

func (manager Manager) GetNRandomQuotes(guildID string, amount int) []Quote {
	guildEntry := manager.FindGuild(guildID)

	return manager.chooseNRandomQuotes(guildEntry.Quotes, amount)
}

func (manager Manager) GetRandomQuoteBySpeaker(speakerID string, guildID string) Quote {
	guildEntry := manager.FindGuild(guildID)
	speakerEntry := manager.FindUser(speakerID, guildEntry.ID)
	quotes := manager.FindManyQuotes(&Quote{SpeakerID: speakerEntry.ID, GuildID: guildEntry.ID})

	return manager.chooseQuoteRandomly(quotes)
}

func (manager Manager) GetNRandomQuotesBySpeaker(speakerID string, guildID string, amount int) []Quote {
	guildEntry := manager.FindGuild(guildID)
	speakerEntry := manager.FindUser(speakerID, guildEntry.ID)
	quotes := manager.FindManyQuotes(&Quote{SpeakerID: speakerEntry.ID, GuildID: guildEntry.ID})

	return manager.chooseNRandomQuotes(quotes, amount)
}

// helper for random quotes
func (manager Manager) chooseQuoteRandomly(quotes []Quote) Quote {
	if len(quotes) > 0 {
		return quotes[rand.Intn(len(quotes))]
	}

	return Quote{}
}

func (manager Manager) chooseNRandomQuotes(quotes []Quote, amount int) []Quote {
	if amount == 1 {
		return []Quote{
			manager.chooseQuoteRandomly(quotes),
		}
	}

	if len(quotes) < amount {
		return quotes
	}

	unselectedQuotes := make([]Quote, len(quotes))
	copy(unselectedQuotes, quotes)
	var selectedQuotes []Quote

	for len(selectedQuotes) < amount && len(unselectedQuotes) > 0 {
		index := rand.Intn(len(unselectedQuotes))
		quote := unselectedQuotes[index]
		selectedQuotes = append(selectedQuotes, quote)
		unselectedQuotes = append(unselectedQuotes[:index], unselectedQuotes[index+1:]...)
	}
	return selectedQuotes
}

func (manager Manager) QuoteExists(query Quote) bool {
	var existing Quote
	result := manager.Database.Where(&query).First(&existing)
	return result.RowsAffected > 0
}

func (manager Manager) GuildExists(guild *discordgo.Guild) bool {
	return manager.GuildExistsByID(guild.ID)
}

// GuildExistsByID - returns true if the user exists, false otherwise
func (manager Manager) GuildExistsByID(guildID string) bool {
	var existing Guild
	result := manager.Database.Where(&Guild{DiscordID: guildID}).First(&existing)
	if result.Error != nil {
		log.Println("Error checking for guild existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

// UserExists - returns true if the user exists, false otherwise
func (manager Manager) UserExists(userID string, guild Guild) bool {
	var existing User
	result := manager.Database.Where(&User{GuildID: guild.ID, DiscordID: userID}).First(&existing)
	if result.Error != nil {
		log.Println("Error checking for user existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

func (manager Manager) UserExistsByName(userName string, guild Guild) bool {
	var existing User
	result := manager.Database.Where(&User{GuildID: guild.ID, Name: userName}).First(&existing)
	if result.Error != nil {
		log.Println("Error checking for user existence: ", result.Error)
	}
	return result.RowsAffected > 0
}

// InsertGuild adds a Guild to the database
// move find and construction logic into separate event code or constructor
func (manager Manager) insertGuild(guild Guild) {
	result := manager.Database.Create(&guild)
	if result.Error != nil {
		log.Println("Error inserting guild: ", result.Error)
		panic(result.Error)
	}

	log.Println("Guild Added: ", guild.Name, result.Name())
}

// InsertUser adds a User to the database
// move find & construction logic into separate event code or constructor
func (manager Manager) insertUser(user User) {
	// insert document into Database
	result := manager.Database.Create(&user)
	if result.Error != nil {
		log.Println("Error inserting user: ", result.Error)
	}
	log.Println("User Added: ", user.Name, result.Name())
}

// InsertQuote adds a Quote to the database
// move find calls and construction into the event code or separate constructor
func (manager Manager) insertQuote(quote Quote) {
	//insert quote into DB
	result := manager.Database.Create(&quote)
	if result.Error != nil {
		log.Println("Error inserting quote: ", result.Error)
	}
	log.Printf("Quote Added: \"%s\" - %s, submitted by %s", quote.Content, quote.Speaker.Name, quote.Submitter.Name)
}

func (manager Manager) FindGuild(guildID string) Guild {
	var guildEntry Guild
	result := manager.Database.
		Where(&Guild{DiscordID: guildID}).
		Preload(clause.Associations).
		Preload("Quotes.Speaker").
		Preload("Quotes.Submitter").
		First(&guildEntry)

	if result.Error != nil {
		log.Println(fmt.Sprintf("Error retrieving guild of ID %s: %s", guildID, result.Error))
	}
	return guildEntry
}

func (manager Manager) FindUser(userID string, guildID uint) User {
	var userEntry User
	result := manager.Database.Where(&User{DiscordID: userID, GuildID: guildID}).First(&userEntry)
	if result.Error != nil {
		log.Println(fmt.Sprintf("Error retrieving user of ID %s: %s", userID, result.Error))
	}
	return userEntry
}

func (manager Manager) FindUserByName(userName string, guildID uint) User {
	var userEntry User
	result := manager.Database.Where(&User{Name: userName, GuildID: guildID}).First(&userEntry)
	if result.Error != nil {
		log.Println(fmt.Sprintf("Error retrieving user of Name %s: %s", userName, result.Error))
	}
	return userEntry
}

func (manager Manager) FindQuote(query *Quote) Quote {
	var quoteEntry Quote
	result := manager.Database.
		Where(&query).
		Preload(clause.Associations).
		First(&quoteEntry)

	if result.Error != nil {
		log.Println("Error retrieving quote", result.Error)
	}
	return quoteEntry
}

func (manager Manager) FindManyQuotes(query *Quote) []Quote {
	var quotes []Quote

	result := manager.Database.Where(&query).
		Preload(clause.Associations).
		Find(&quotes)

	if result.Error != nil {
		log.Println("Error retrieving quotes", result.Error)
	}

	return quotes
}

func (manager Manager) UpdateGuild(discordGuild *discordgo.Guild) {
	guild := manager.FindGuild(discordGuild.ID)
	guild.DiscordID = discordGuild.ID
	guild.Name = discordGuild.Name
	manager.Database.Save(&guild)
}

func (manager Manager) UpdateGuildUser(discordUser *discordgo.User, guild Guild) {
	user := manager.FindUser(discordUser.ID, guild.ID)
	user.DiscordID = discordUser.ID
	user.Name = discordUser.Username
	manager.Database.Save(&user)
}
