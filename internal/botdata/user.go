package botdata

import "go.mongodb.org/mongo-driver/bson/primitive"

// User - An object representing a User
type User struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`       // Mongo ID of this object
	Date    primitive.DateTime `bson:"date,omitempty"`      // date added to database
	GuildID primitive.ObjectID `bson:"guild,omitempty"`     // Mongo ID of the guild this user belongs to
	Name    string             `bson:"name,omitempty"`      // Name of the User
	UserID  string             `bson:"discordid,omitempty"` // Disocrd User ID
}
