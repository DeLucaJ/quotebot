package types

// User - An object representing a User
type User struct {
	// mongo object ID
	GuildID string // Discord Guild ID
	Name    string // Name of the User
	UserID  string // Disocrd User ID
	// date added (mongo)
}
