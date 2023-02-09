package botdata

import (
	"time"
)

// Guild - Represents a Discord Server associated with this bot
type Guild struct {
	ID        uint      // the ID of the Guild in the database
	Date      time.Time // date the Guild was added to the database
	DiscordID string    // The Discord ID of the guild
	Name      string    // name of the guild
}
