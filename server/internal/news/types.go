package news

import "time"

// Article is a news article as the API returns it.
type Article struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Date     time.Time `json:"date"`
	ImageURL string    `json:"imageUrl,omitempty"`
}

// CreateInput is a new article. Content is HTML from the editor and is sanitised.
type CreateInput struct {
	Title   string
	Content string
}

// UpdateInput changes an article. Nil fields are left unchanged; an empty
// Content clears it. RemoveImage is ignored when a new image is uploaded.
type UpdateInput struct {
	Title       *string
	Content     *string
	RemoveImage bool
}
