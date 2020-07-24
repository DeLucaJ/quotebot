package types

// Quote - Object representing a quote
type Quote struct {
	// mongo object ID
	Content   string
	Speaker   User
	Submitter User
	Guild     Guild
	// date added (mongo)
}
