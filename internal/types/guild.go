package types

import "go.mongodb.org/mongo-driver/bson/primitive"

// Guild - Represents a Discord Server associated with this bot
type Guild struct {
	ID        primitive.ObjectID   // the ID of the Guild in the database
	Date      primitive.DateTime   // date the Guild was added to the database
	DiscordID string               // The Discord ID of the guild
	Users     []primitive.ObjectID // the IDs of the Users associated with this guild
	Quotes    []primitive.ObjectID // the IDs of the Quotes associated with this server

}
