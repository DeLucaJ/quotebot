package data

import "gorm.io/gorm"

// Quote - Object representing a quote
type Quote struct {
	gorm.Model
	Content     string // The content of the quote
	SpeakerID   uint   // The ID of the one who spoke the Quote
	Speaker     User   // The User that spoke the Quote
	SubmitterID uint   // The ID of the one who submitted the Quote
	Submitter   User   // The User that submitted the Quote
	GuildID     uint   // the ID of the Guild the quote was posted in
	Guild       Guild  // The Guild the Quote was posted in
}
