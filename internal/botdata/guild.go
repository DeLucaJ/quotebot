package botdata

import "go.mongodb.org/mongo-driver/bson/primitive"

// Guild - Represents a Discord Server associated with this bot
type Guild struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`       // the ID of the Guild in the database
	Date      primitive.DateTime   `bson:"date,omitempty"`      // date the Guild was added to the database
	DiscordID string               `bson:"discordid,omitempty"` // The Discord ID of the guild
	Name      string               `bson:"name,omitempty"`      // name of the guild
	Users     []primitive.ObjectID `bson:"users,omitempty"`     // the IDs of the Users associated with this guild
	Quotes    []primitive.ObjectID `bson:"quotes,omitempty"`    // the IDs of the Quotes associated with this server
}
