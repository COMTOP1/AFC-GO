package setting

// InfoContent is the body of GET and PUT /info.
type InfoContent struct {
	Content string `json:"content"`
}

// DisplayEmail is the body of PUT /settings/display-email.
type DisplayEmail struct {
	Email string `json:"email"`
}

// TeamEmailModeInput is the body of PUT /settings/team-email-mode.
type TeamEmailModeInput struct {
	// Mode is "manager" (each manager's own email), "team" (the team's
	// contact address) or "site" (the site-wide contact email).
	Mode TeamEmailMode `json:"mode"`
}
