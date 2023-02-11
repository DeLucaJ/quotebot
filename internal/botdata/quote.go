package botdata

import "gorm.io/gorm"

// Quote - Object representing a quote
type Quote struct {
	gorm.Model
	Content   string // The content of the quote
	Speaker   uint   // The ID of the one who spoke the quote
	Submitter uint   // The ID of the one who submitted the quote
	Guild     uint   // the ID of the Guild the quote was posted in
}
