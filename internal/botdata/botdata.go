package botdata

import (
	"context"
	"fmt"
	"log"
	"time"

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
