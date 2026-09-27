package whatson

import "time"

// Event is a what's-on entry as the API returns it.
type Event struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Date        time.Time `json:"date"`
	DateOfEvent time.Time `json:"dateOfEvent"`
	ImageURL    string    `json:"imageUrl,omitempty"`
}

// Period filters events by date of event relative to today.
type Period string

const (
	PeriodAll    Period = "all"
	PeriodFuture Period = "future"
	PeriodPast   Period = "past"
)

// CreateInput is a new event. Content is sanitised HTML.
type CreateInput struct {
	Title       string
	Content     string
	DateOfEvent time.Time
}

// UpdateInput changes an event; nil fields are left unchanged.
type UpdateInput struct {
	Title       *string
	Content     *string
	DateOfEvent *time.Time
	RemoveImage bool
}
