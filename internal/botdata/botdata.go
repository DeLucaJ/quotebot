package botdata

import {
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "go.mongodb.org/mongo-driver/bson"
}

const uri = "mongodb://localhost:27017"

func init() {
	fmt.Println("Initializeing Mongo Client")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defers the context cancel
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Error connecting to mongo client: " + err.Error())
		log.Fatal(err)
	}

	// Defer a disconnect from the database
	defer func() {
		if err != nil {
			fmt.Println("Error disconnecting from mongo client: " + err.Error())
			log.Fatal(err)
		}
	}()

	// Initialize the Database
	database := client.Database("quotedb")
	guildCollection := database.Collection("guilds")
	userCollection := database.Collection("users")
	quoteCollection := database.Collection("quotes")
}
