package botdata

import "gorm.io/gorm"

// User - An object representing a User
type User struct {
	gorm.Model
	GuildID   uint   // database ID of the guild this user belongs to
	Name      string // Name of the User
	DiscordID string // Disocrd User ID
}
