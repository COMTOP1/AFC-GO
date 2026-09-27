package document

// Public is a downloadable club document as the API returns it.
type Public struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	FileURL string `json:"fileUrl"`
}

// CreateInput is a new document; a file is required.
type CreateInput struct {
	Name string
}
