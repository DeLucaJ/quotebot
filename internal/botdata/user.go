package botdata

import (
	"time"
)

// User - An object representing a User
type User struct {
	ID      uint      // Mongo ID of this object
	Date    time.Time // date added to database
	GuildID uint      // Mongo ID of the guild this user belongs to
	Name    string    // Name of the User
	UserID  string    // Disocrd User ID
}
