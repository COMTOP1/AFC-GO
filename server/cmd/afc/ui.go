package main

import "embed"

// uiFiles is the built React client. `yarn build:server` and the Dockerfile
// copy build/client into ui/ before compiling; only ui/.keep is committed, so
// a Go-only build serves a "client not built" page for every page.
//
//go:embed all:ui
var uiFiles embed.FS
