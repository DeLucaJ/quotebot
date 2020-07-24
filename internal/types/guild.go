package types

// Guild - Represents a Discord Server associated with this bot
type Guild struct {
	// mongo object ID
	DiscordID string
	Users     []User
	Quotes    []Quote
	// date added (mongo)
}
