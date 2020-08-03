package types

import "go.mongodb.org/mongo-driver/bson/primitive"

// Quote - Object representing a quote
type Quote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`       // the Mongo object ID of the quote
	Date      primitive.DateTime `bson:"date,omitempty"`      // The date the quote was added
	Content   string             `bson:"content,omitempty"`   // The content of the quote
	Speaker   primitive.ObjectID `bson:"speaker,omitempty"`   // The ID of the one who spoke the quote
	Submitter primitive.ObjectID `bson:"submitter,omitempty"` // The ID of the one who submitted the quote
	Guild     primitive.ObjectID `bson:"guild,omitempty"`     // the ID of the Guild the quote was posted in
}
