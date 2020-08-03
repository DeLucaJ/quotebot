package types

import "go.mongodb.org/mongo-driver/bson/primitive"

// Quote - Object representing a quote
type Quote struct {
	// mongo object ID
	ID        primitive.ObjectID // the Mongo object ID of the quote
	Date      primitive.DateTime // The date the quote was added
	Content   string             // The content of the quote
	Speaker   primitive.ObjectID // The ID of the one who spoke the quote
	Submitter primitive.ObjectID // The ID of the one who submitted the quote
	Guild     primitive.ObjectID // the ID of the Guild the quote was posted in
}
