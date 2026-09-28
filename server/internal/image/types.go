package image

// Public is a gallery photo as the API returns it.
type Public struct {
	ID       int    `json:"id"`
	Caption  string `json:"caption,omitempty"`
	ImageURL string `json:"imageUrl"`
}

// CreateInput is a new gallery photo; the image is required.
type CreateInput struct {
	Caption string
}
