// Package sanitize cleans rich-text HTML submitted from the editors.
package sanitize

import "github.com/microcosm-cc/bluemonday"

// policy is the legacy editor policy (views/news.go, whatson.go, info.go).
// A bluemonday policy is safe for concurrent use once built.
var policy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("a", "ul", "ol", "li", "h2", "b", "i", "u", "strike", "div", "br", "p",
		"blockquote", "pre", "hr")
	p.AllowAttrs("class").OnElements("h2")
	p.AllowAttrs("href", "style").OnElements("a")
	p.AllowURLSchemes("mailto", "http", "https")
	p.RequireNoFollowOnLinks(false)
	// Justification - via inline style
	p.AllowAttrs("style").OnElements("div", "p", "h2", "span")
	return p
}()

// HTML returns s with everything outside the editor policy removed.
func HTML(s string) string {
	return policy.Sanitize(s)
}
