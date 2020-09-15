package botdata

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Manager - struct that holds the data necessary to manage the database
type Manager struct {
	client   *mongo.Client
	ctx      context.Context
	cancel   context.CancelFunc
	database *mongo.Database
	guildcol *mongo.Collection
	usercol  *mongo.Collection
	quotecol *mongo.Collection
}

// Shutdown - Ends the connection to the database and cleans the context
func (bm Manager) Shutdown() {
	cancel := bm.cancel
	client := bm.client
	ctx := bm.ctx

	defer cancel()
	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			fmt.Println("Error disconnecting from mongo client: " + err.Error())
			log.Fatal(err)
		}
	}()
}

// GuildCreate - Event handler called when the logs on or joins a guild
func (bm Manager) GuildCreate(session *discordgo.Session, event *discordgo.GuildCreate) {
	if !bm.guildExists(event.Guild.ID) {
		bm.insertGuild(*event.Guild)
	}
	fmt.Println("Login: ", event.Guild.Name)
}

// MessageCreate - Event handler called when a message is created in a joined Guild
func (bm Manager) MessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Check to see if the author is this bot
	if message.Author.ID == session.State.User.ID {
		return
	}
	// fmt.Println("Recieved a Message: ", message.Content)
}

func (bm Manager) guildExists(guildid string) bool {
	var existing Guild
	err := bm.usercol.FindOne(bm.ctx, bson.M{"discordid": guildid}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		return false
	}
	if err != nil {
		fmt.Println("Error checking for guild existence: ", err)
		log.Fatal(err)
	}
	return true
}

func (bm Manager) userExists(userid string, guildid string) bool {
	var existing User
	err := bm.usercol.FindOne(bm.ctx, bson.M{"userid": userid, "guild": guildid}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		return false
	}
	if err != nil {
		fmt.Println("Error checking for user existence: ", err)
		log.Fatal(err)
	}
	return true
}

// adds a Guild to the database
func (bm Manager) insertGuild(guild discordgo.Guild) Guild {
	guildDoc := Guild{
		Date:      primitive.NewDateTimeFromTime(time.Now()),
		DiscordID: guild.ID,
		Name:      guild.Name,
	}

	result, err := bm.guildcol.InsertOne(bm.ctx, guildDoc)
	if err != nil {
		fmt.Println("Error inserting guild: ", err)
		panic(err)
	}

	fmt.Println("Guild Add: ", guild.Name, result.InsertedID)

	return guildDoc
}

// adds a User to the database
func (bm Manager) insertUser(user discordgo.User, guild discordgo.Guild) {
	if bm.userExists(user.ID, guild.ID) {
		// user in database
		fmt.Println("User already exists: ", user.Username)
	} else {
		// user not in database
		// find the guild object
		guildDoc := bm.findGuild(guild.ID)

		// add the new user
		userDoc := User{
			Date:    primitive.NewDateTimeFromTime(time.Now()),
			GuildID: guildDoc.ID,
			Name:    user.Username,
			UserID:  user.ID,
		}

		// insert document into database
		result, err := bm.usercol.InsertOne(bm.ctx, userDoc)
		if err != nil {
			fmt.Println("Error inserting user: ", err)
			panic(err)
		}
		fmt.Println("User Added: ", userDoc.Name, result.InsertedID)
	}
}

// adds a Quote to the database
func (bm Manager) insertQuote(message discordgo.Message) {

}

func (bm Manager) findGuild(guildid string) Guild {
	var guild Guild
	err := bm.guildcol.FindOne(bm.ctx, bson.M{"discordid": guildid}).Decode(&guild)
	if err != nil {
		fmt.Println("Error retrieving guild: ", err)
		log.Fatal(err)
	}
	return guild
}

func (bm Manager) findUser(userid string, guildid string) User {
	// retrieve guild data
	guild := bm.findGuild(guildid)

	var user User
	err := bm.usercol.FindOne(bm.ctx, bson.M{"userid": userid, "guildid": guild.ID}).Decode(&user)
	if err != nil {
		fmt.Println("Error retrieving user: ", err)
		log.Fatal(err)
	}
	return user
}

func (bm Manager) findQuote(quoteref primitive.ObjectID) Quote {
	var quote Quote

	err := bm.quotecol.FindOne(bm.ctx, bson.M{"_id": quoteref}).Decode(&quote)
	if err != nil {
		fmt.Println("Error retrieving quote: ", err)
		log.Fatal(err)
	}
	return quote
}

// sends a message with a random Quote from the database
func (bm Manager) chooseRandomQuote(guildid string) Quote {
	guild := bm.findGuild(guildid)

	if len(guild.Quotes) > 0 {
		return bm.findQuote(guild.Quotes[rand.Intn(len(guild.Quotes))])
	} else {
		return Quote{}
	}
}

// sends a message with a quote spoken by a specific user
func (bm Manager) choosetQuoteByUser() {

}

// sends a message with every quote spoken by a specific user
func (bm Manager) chooseAllQuotesByUser() {

}

// flags a quote for inspection by administrator
func (bm Manager) flagQutoe() {

}

// sets a different alias for a user in the database
func (bm Manager) renameUser() {

}

// converts all quotes by an unassociated user to this user
func (bm Manager) claimUser() {

}

// Start - Starts the boss and initializes the database connection
// 	uri string: the uri for the database from config.json
func Start(uri string) Manager {
	fmt.Println("Initializeing Mongo Client")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel() // Defers the context cancel

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		// defers closing functions after error
		defer cancel()
		defer func() {
			err := client.Disconnect(ctx)
			if err != nil {
				fmt.Println("Error disconnecting from mongo client: " + err.Error())
				log.Fatal(err)
			}
		}()

		fmt.Println("Error connecting to mongo client: " + err.Error())
		log.Fatal(err)
	}

	// Initialize the Database
	db := client.Database("quotebotx")
	gc := db.Collection("guilds")
	uc := db.Collection("users")
	qc := db.Collection("quotes")

	// initializes the singleton botmanager
	return Manager{
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		database: db,
		guildcol: gc,
		usercol:  uc,
		quotecol: qc,
	}
}
