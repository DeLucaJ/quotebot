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
	if !bm.GuildExists(*event.Guild) {
		bm.AddGuild(*event.Guild)
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

// AddGuild - adds a guild to the database
func (bm Manager) AddGuild(guild discordgo.Guild) {
	guilddoc := Guild{
		Date:      primitive.NewDateTimeFromTime(time.Now()),
		DiscordID: guild.ID,
		Name:      guild.Name,
	}
	bm.insertGuild(guilddoc)
}

// AddUser - adds a user to the database
func (bm Manager) AddUser(name string, user discordgo.User, guild discordgo.Guild) {
	guilddoc := bm.findGuild(bson.M{"discordid": guild.ID})

	userdoc := User{
		Date:    primitive.NewDateTimeFromTime(time.Now()),
		Name:    name,
		UserID:  user.ID,
		GuildID: guilddoc.ID,
	}

	bm.insertUser(userdoc)
}

// AddQuote - adds a qutoe to the database
func (bm Manager) AddQuote(content string, speaker discordgo.User, submitter discordgo.User, guild discordgo.Guild) {
	guilddoc := bm.findGuild(bson.M{"discordid": guild.ID})
	speakerdoc := bm.findUser(bson.M{"discordid": speaker.ID, "guild": guild.ID})
	submitterdoc := bm.findUser(bson.M{"discordid": speaker.ID, "guild": guild.ID})

	quote := Quote{
		Date:      primitive.NewDateTimeFromTime(time.Now()),
		Content:   content,
		Speaker:   speakerdoc.ID,
		Submitter: submitterdoc.ID,
		Guild:     guilddoc.ID,
	}

	bm.insertQuote(quote)
}

// ChooseRandomQuote - Chooses a random quote from a specific guild
func (bm Manager) ChooseRandomQuote(guild discordgo.Guild) Quote {
	guilddoc := bm.findGuild(bson.M{"discordid": guild.ID})
	quotes := bm.findManyQuotes(bson.M{"guild": guilddoc.ID})

	return bm.randomQuote(quotes)
}

// ChooseRandomQuoteBySpeaker - Chooses a random quote from guild and speaker
func (bm Manager) ChooseRandomQuoteBySpeaker(speakerid string, guild discordgo.Guild) Quote {
	guilddoc := bm.findGuild(bson.M{"discordid": guild.ID})
	quotes := bm.findManyQuotes(bson.M{"speaker": speakerid, "guild": guilddoc.ID})

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
	err := bm.usercol.FindOne(bm.ctx, bson.M{"discordid": guild.ID}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		return false
	}
	if err != nil {
		fmt.Println("Error checking for guild existence: ", err)
		log.Fatal(err)
	}
	return true
}

// UserExists - returns true if the user exists, false otherwise
func (bm Manager) UserExists(user discordgo.User, guild discordgo.Guild) bool {
	var existing User
	err := bm.usercol.FindOne(bm.ctx, bson.M{"userid": user.ID, "guild": guild.ID}).Decode(&existing)
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
// move find and construction logic into separate event code or constructor
func (bm Manager) insertGuild(guild Guild) {
	result, err := bm.guildcol.InsertOne(bm.ctx, guild)
	if err != nil {
		fmt.Println("Error inserting guild: ", err)
		panic(err)
	}

	fmt.Println("Guild Add: ", guild.Name, result.InsertedID)
}

// adds a User to the database
//move find & construction logic into separate event code or constructor
func (bm Manager) insertUser(user User) {
	// insert document into database
	result, err := bm.usercol.InsertOne(bm.ctx, user)
	if err != nil {
		fmt.Println("Error inserting user: ", err)
		log.Fatal(err)
	}
	fmt.Println("User Added: ", user.Name, result.InsertedID)
}

// adds a Quote to the database
// move find calls and construction into the event code or separate constructor
func (bm Manager) insertQuote(quote Quote) {
	//insert quote into DB
	result, err := bm.quotecol.InsertOne(bm.ctx, quote)
	if err != nil {
		fmt.Println("Error inserting quote: ", err)
		log.Fatal(err)
	}
	fmt.Println("Quote Added: ", result.InsertedID)
}

func (bm Manager) findGuild(query primitive.M) Guild {
	var guild Guild
	err := bm.guildcol.FindOne(bm.ctx, query).Decode(&guild)
	if err != nil {
		fmt.Println("Error retrieving guild: ", err)
		log.Fatal(err)
	}
	return guild
}

func (bm Manager) findManyGuilds(query primitive.M) []Guild {
	var guilds []Guild

	cursor, err := bm.quotecol.Find(bm.ctx, query)
	if err != nil {
		fmt.Println("Error finding qutoes: ", err)
		log.Fatal(err)
	}
	if err = cursor.All(bm.ctx, &guilds); err != nil {
		fmt.Println("Err converting cursor to quote slice: ", err)
		log.Fatal(err)
	}

	return guilds
}

func (bm Manager) findUser(query primitive.M) User {
	// bson.M{"userid": userid, "guildid": guild}
	var user User
	err := bm.usercol.FindOne(bm.ctx, query).Decode(&user)
	if err != nil {
		fmt.Println("Error retrieving user: ", err)
		log.Fatal(err)
	}
	return user
}

func (bm Manager) findManyUsers(query primitive.M) []User {
	var users []User

	cursor, err := bm.usercol.Find(bm.ctx, query)
	if err != nil {
		fmt.Println("Error finding qutoes: ", err)
		log.Fatal(err)
	}
	if err = cursor.All(bm.ctx, &users); err != nil {
		fmt.Println("Err converting cursor to quote slice: ", err)
		log.Fatal(err)
	}

	return users
}

func (bm Manager) findQuote(query primitive.ObjectID) Quote {
	// bson.M{"_id": quoteref}
	var quote Quote

	err := bm.quotecol.FindOne(bm.ctx, query).Decode(&quote)
	if err != nil {
		fmt.Println("Error retrieving quote: ", err)
		log.Fatal(err)
	}
	return quote
}

func (bm Manager) findManyQuotes(query primitive.M) []Quote {
	var quotes []Quote

	cursor, err := bm.quotecol.Find(bm.ctx, query)
	if err != nil {
		fmt.Println("Error finding qutoes: ", err)
		log.Fatal(err)
	}
	if err = cursor.All(bm.ctx, &quotes); err != nil {
		fmt.Println("Err converting cursor to quote slice: ", err)
		log.Fatal(err)
	}

	return quotes
}

/* // flags a quote for inspection by administrator
func (bm Manager) flagQutoe() {

}

// sets a different alias for a user in the database
func (bm Manager) renameUser() {

}

// converts all quotes by an unassociated user to this user
func (bm Manager) claimUser() {

} */
