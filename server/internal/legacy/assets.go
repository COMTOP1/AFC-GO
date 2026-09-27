// Package legacy holds the server-rendered template site that is being
// replaced by the React client. It is deleted once the client ships.
package legacy

import "embed"

// Public holds the legacy static assets served under /public/.
//
//go:embed public
var Public embed.FS
