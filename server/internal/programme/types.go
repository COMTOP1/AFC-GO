package programme

import "time"

// Public is a match programme as the API returns it.
type Public struct {
	ID      int           `json:"id"`
	Name    string        `json:"name"`
	Date    time.Time     `json:"date"`
	FileURL string        `json:"fileUrl"`
	Season  *PublicSeason `json:"season,omitempty"`
}

// PublicSeason is a programme season.
type PublicSeason struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CreateInput is a new programme; a file is required. SeasonID 0 means none.
type CreateInput struct {
	Name     string
	Date     time.Time
	SeasonID int
}
