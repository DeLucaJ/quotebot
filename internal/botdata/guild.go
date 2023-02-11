package botdata

import (
	"gorm.io/gorm"
)

// Guild - Represents a Discord Server associated with this bot
type Guild struct {
	gorm.Model
	DiscordID string // The Discord ID of the guild
	Name      string // name of the guild
}
