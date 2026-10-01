package web

import (
	"errors"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	cacheImmutable = "public, max-age=31536000, immutable"
	cacheNone      = "no-cache"
	notBuiltPage   = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>AFC Aldermaston</title></head>` +
		`<body><p>The web client has not been built. Run <code>yarn build:client</code>.</p></body></html>`
)

// SPARoutes are the React client's routes (client/App.tsx; a test keeps them
// in step). Any other page path still gets the app, which shows "Page not
// found", but with a 404 status so crawlers and link checkers see it.
var SPARoutes = []string{
	"/",
	"/teams", "/teams/new", "/team/:id", "/team/:id/edit",
	"/news", "/news/new", "/news/:id", "/news/:id/edit",
	"/whatson", "/whatson/new", "/whatson/:id", "/whatson/:id/edit",
	"/gallery", "/documents", "/programmes", "/sponsors", "/players", "/users",
	"/info", "/info/edit", "/contact", "/account", "/reset/:token", "/design",
}

// IsAppRoute reports whether p matches one of SPARoutes; ":name" matches one
// non-empty path segment.
func IsAppRoute(p string) bool {
	for _, r := range SPARoutes {
		if routeMatches(r, p) {
			return true
		}
	}
	return false
}

func routeMatches(pattern, p string) bool {
	if pattern == "/" || p == "/" {
		return pattern == p
	}
	if !strings.HasPrefix(p, "/") {
		return false
	}
	want := strings.Split(pattern[1:], "/")
	got := strings.Split(p[1:], "/")
	if len(want) != len(got) {
		return false
	}
	for i, seg := range want {
		if got[i] == "" {
			return false
		}
		if !strings.HasPrefix(seg, ":") && seg != got[i] {
			return false
		}
	}
	return true
}

func isAPI(p string) bool {
	return p == "/api" || strings.HasPrefix(p, APIPrefix)
}

// MountSPA serves the React client at the site root, leaving /api alone.
// With proxy set (development) every non-API request is reverse-proxied to
// the Vite dev server, including HMR websockets. Otherwise files (the built
// client; may be nil) are served: hashed assets, root files, and index.html
// for page paths (200 for known app routes, 404 otherwise).
func MountSPA(e *echo.Echo, files fs.FS, proxy *url.URL) {
	var h echo.HandlerFunc
	if proxy != nil {
		slog.Info("serving the web client from the Vite dev server", "url", proxy.String())
		rp := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(proxy)
			r.SetXForwarded()
		}}
		proxied := echo.WrapHandler(rp)
		h = func(c echo.Context) error {
			if isAPI(c.Request().URL.Path) {
				return echo.ErrNotFound
			}
			return proxied(c)
		}
	} else {
		s := &spa{files: files}
		s.built = files != nil && isFile(files, "index.html")
		if !s.built {
			slog.Warn("web client not built: pages will answer 503 until `yarn build` is run")
		}
		h = s.serve
	}
	e.Any("/", h)
	e.Any("/*", h)
}

type spa struct {
	files fs.FS
	built bool
}

func (s *spa) serve(c echo.Context) error {
	p := c.Request().URL.Path
	if isAPI(p) {
		return echo.ErrNotFound
	}
	if m := c.Request().Method; m != http.MethodGet && m != http.MethodHead {
		return c.String(http.StatusMethodNotAllowed, "method not allowed")
	}
	if !s.built {
		c.Response().Header().Set("Cache-Control", cacheNone)
		return c.HTML(http.StatusServiceUnavailable, notBuiltPage)
	}

	rel := strings.TrimPrefix(p, "/")
	isAsset := rel == "assets" || strings.HasPrefix(rel, "assets/")
	switch {
	case rel == "":
		return s.file(c, "index.html", cacheNone, http.StatusOK)
	case (isAsset || path.Ext(rel) != "") && (!fs.ValidPath(rel) || strings.Contains(rel, "\\")):
		// Only these read a file named by the URL; never let it escape the build.
		return c.String(http.StatusNotFound, "not found")
	case isAsset:
		return s.file(c, rel, cacheImmutable, http.StatusOK)
	case path.Ext(rel) != "":
		return s.file(c, rel, cacheNone, http.StatusOK)
	case IsAppRoute(p):
		return s.file(c, "index.html", cacheNone, http.StatusOK)
	default:
		return s.file(c, "index.html", cacheNone, http.StatusNotFound)
	}
}

// file writes name from the build with status, or a plain 404. It never
// falls back to index.html, so a missing hashed asset is visible rather than
// served as HTML.
func (s *spa) file(c echo.Context, name, cacheControl string, status int) error {
	b, err := fs.ReadFile(s.files, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || isDir(s.files, name) {
			return c.String(http.StatusNotFound, "not found")
		}
		return err
	}
	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = http.DetectContentType(b)
	}
	c.Response().Header().Set("Cache-Control", cacheControl)
	return c.Blob(status, ctype, b)
}

func isFile(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func isDir(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && info.IsDir()
}
