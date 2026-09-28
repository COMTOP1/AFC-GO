package setting

// InfoContent is the body of GET and PUT /info.
type InfoContent struct {
	Content string `json:"content"`
}

// DisplayEmail is the body of PUT /settings/display-email.
type DisplayEmail struct {
	Email string `json:"email"`
}
