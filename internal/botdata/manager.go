package botdata

import (
	"context"
	"fmt"
	"log"
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
	bm.addGuild(*event.Guild)
}

// MessageCreate - Event handler called when a message is created in a joined Guild
func (bm Manager) MessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Check to see if the author is this bot
	if message.Author.ID == session.State.User.ID {
		return
	}
	fmt.Println("Recieved a Message: ", message.Content)
}

// adds a Guild to the database
func (bm Manager) addGuild(guild discordgo.Guild) {
	// needs a check for existing guilds
	var oldGuild Guild
	err := bm.guildcol.FindOne(bm.ctx, bson.M{"discordid": guild.ID}).Decode(&oldGuild)
	if err == mongo.ErrNoDocuments {
		// add the guild to the database
		ng := Guild{
			Date:      primitive.NewDateTimeFromTime(time.Now()),
			DiscordID: guild.ID,
			Name:      guild.Name,
		}

		result, err := bm.guildcol.InsertOne(bm.ctx, ng)
		if err != nil {
			fmt.Println("Error inserting guild: ", err)
			panic(err)
		}

		fmt.Print("Login: ", ng.Name, result.InsertedID)
		fmt.Println(" (NEW)")
		return

	} else if err != nil {
		fmt.Println("Error searching for existing guild: ", err)
		log.Fatal(err)
	} else {
		// guild exists
		fmt.Println("Login: ", oldGuild.Name, oldGuild.ID)
	}

}

// adds a User to the database
func (bm Manager) addUser(user discordgo.User, guild discordgo.Guild) {

}

// adds a Quote to the database
func (bm Manager) addQuote(message discordgo.Message) {

}

// sends a message with a random Quote from the database
func (bm Manager) getRandomQuote() {

}

// sends a message with a quote spoken by a specific user
func (bm Manager) getQuoteByUser() {

}

// sends a message with every quote spoken by a specific user
func (bm Manager) getAllQuotesByUser() {

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
