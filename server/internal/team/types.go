package team

// Public is a team as the API returns it.
type Public struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	League         string `json:"league,omitempty"`
	Division       string `json:"division,omitempty"`
	LeagueTableURL string `json:"leagueTableUrl,omitempty"`
	FixturesURL    string `json:"fixturesUrl,omitempty"`
	Coach          string `json:"coach,omitempty"`
	Physio         string `json:"physio,omitempty"`
	ImageURL       string `json:"imageUrl,omitempty"`
	IsActive       bool   `json:"isActive"`
	IsYouth        bool   `json:"isYouth"`
	Ages           int    `json:"ages"`
}

// CreateInput is a new team. Ages below 19 always make a youth team.
type CreateInput struct {
	Name, Description, League, Division, LeagueTable, Fixtures, Coach, Physio string

	IsActive bool
	IsYouth  bool
	Ages     int
}

// UpdateInput changes a team; nil fields are left unchanged and empty
// strings clear optional fields.
type UpdateInput struct {
	Name, Description, League, Division, LeagueTable, Fixtures, Coach, Physio *string

	IsActive    *bool
	IsYouth     *bool
	Ages        *int
	RemoveImage bool
}
