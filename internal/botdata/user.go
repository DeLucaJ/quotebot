package botdata

import "gorm.io/gorm"

// User - An object representing a User
type User struct {
	gorm.Model
	Name      string // Name of the User
	DiscordID string // Discord User ID
	GuildID   uint   // database ID of the guild this user belongs to
	Guild     Guild  // the guild this user belongs to
}
