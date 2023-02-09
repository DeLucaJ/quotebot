package botdata

import (
	"time"
)

// Quote - Object representing a quote
type Quote struct {
	ID        uint      // the Mongo object ID of the quote
	Date      time.Time // The date the quote was added
	Content   string    // The content of the quote
	Speaker   uint      // The ID of the one who spoke the quote
	Submitter uint      // The ID of the one who submitted the quote
	Guild     uint      // the ID of the Guild the quote was posted in
}
