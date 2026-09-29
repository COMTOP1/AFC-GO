package player

import "time"

// Public is a player as the logged-in players list returns it. ImageURL is
// empty whenever PhotoVisible says so.
type Public struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Position    string     `json:"position,omitempty"`
	IsCaptain   bool       `json:"isCaptain"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	Age         *int       `json:"age,omitempty"`
	Team        *TeamRef   `json:"team,omitempty"`
	ImageURL    string     `json:"imageUrl,omitempty"`
}

// TeamRef is the team a player belongs to.
type TeamRef struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsYouth bool   `json:"isYouth"`
}

// Member is a player on a public team page: no date of birth.
type Member struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Position  string `json:"position,omitempty"`
	IsCaptain bool   `json:"isCaptain"`
	ImageURL  string `json:"imageUrl,omitempty"`
}

// CreateInput is a new player.
type CreateInput struct {
	Name        string
	Position    string
	TeamID      int
	DateOfBirth time.Time
	IsCaptain   bool
}

// UpdateInput changes a player; nil fields are left unchanged.
type UpdateInput struct {
	Name        *string
	Position    *string
	TeamID      *int
	DateOfBirth *time.Time
	IsCaptain   *bool
	RemoveImage bool
}
