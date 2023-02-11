package botdata

import (
	"gorm.io/gorm"
)

// Guild - Represents a Discord Server associated with this bot
type Guild struct {
	gorm.Model
	DiscordID string  // The Discord ID of the guild
	Name      string  // name of the guild
	Users     []User  // All the User entries in the Guild
	Quotes    []Quote // All the Quote entries in the Guild
}
