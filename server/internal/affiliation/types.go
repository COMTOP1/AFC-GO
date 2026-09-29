package affiliation

// Public is an affiliation as the API returns it.
type Public struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Website  string `json:"website,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CreateInput is a new affiliation; an image is required.
type CreateInput struct {
	Name    string
	Website string
}
