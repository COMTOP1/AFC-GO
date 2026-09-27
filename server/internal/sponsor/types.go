package sponsor

// Public is a sponsor as the API returns it. Team is "A", "O" or "Y"
// (legacy audience codes) or a team ID as a string; empty means unassigned.
type Public struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Website  string `json:"website,omitempty"`
	Purpose  string `json:"purpose,omitempty"`
	Team     string `json:"team"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CreateInput is a new sponsor; an image is required.
type CreateInput struct {
	Name    string
	Website string
	Purpose string
	Team    string
}
