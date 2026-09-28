// Package site serves read-only views that combine several domains: shared
// layout data, the home page, the contact page and the team page.
package site

import (
	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/player"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/whatson"
)

// Info is the data every page's layout needs.
type Info struct {
	Year         int           `json:"year"`
	VisitorCount int           `json:"visitorCount"`
	DisplayEmail string        `json:"displayEmail,omitempty"`
	Version      string        `json:"version"`
	Teams        []team.Public `json:"teams"`
}

// Home is the home page. Panels whose data failed to load are omitted.
type Home struct {
	LatestNews   *news.Article        `json:"latestNews,omitempty"`
	NextEvent    *whatson.Event       `json:"nextEvent,omitempty"`
	Sponsors     []sponsor.Public     `json:"sponsors"`
	Affiliations []affiliation.Public `json:"affiliations"`
}

// Contact is the contact page.
type Contact struct {
	DisplayEmail string          `json:"displayEmail,omitempty"`
	People       []ContactPerson `json:"people"`
}

// ContactPerson is a club official listed on the contact page.
type ContactPerson struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// TeamDetail is the public team page.
type TeamDetail struct {
	Team     team.Public      `json:"team"`
	Managers []Manager        `json:"managers"`
	Sponsors []sponsor.Public `json:"sponsors"`
	Players  []player.Member  `json:"players"`
}

// Manager is a team manager's public contact.
type Manager struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
