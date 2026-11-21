package utils

import (
	"net/url"
	"strings"
)

// IsWebURL reports whether s is an absolute http or https URL with a host,
// the only kind of link editors may save (sponsor and affiliation websites,
// league tables, fixtures). url.ParseRequestURI alone also accepts
// "javascript:..." and "data:..." URLs, which would run script or load
// content of the attacker's choosing when a visitor clicks the link.
func IsWebURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil || u.Host == "" {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return true
	}
	return false
}
