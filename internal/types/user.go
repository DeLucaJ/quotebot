package types

import "go.mongodb.org/mongo-driver/bson/primitive"

// User - An object representing a User
type User struct {
	ID      primitive.ObjectID   // Mongo ID of this object
	Date    primitive.DateTime   // date added to database
	GuildID primitive.ObjectID   // Mongo ID of the guild this user belongs to
	Name    string               // Name of the User
	UserID  string               // Disocrd User ID
	Quotes  []primitive.ObjectID // a list of IDs for all quotes associated with this user
}
