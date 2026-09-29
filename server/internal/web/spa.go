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

// SPAPrefix is where the React client is served until it replaces the legacy
// site (sub-project 5 moves it to /).
const SPAPrefix = "/app"

const (
	cacheImmutable = "public, max-age=31536000, immutable"
	cacheNone      = "no-cache"
	notBuiltPage   = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>AFC Aldermaston</title></head>` +
		`<body><p>The web client has not been built. Run <code>yarn build:client</code>.</p></body></html>`
)

// MountSPA serves the React client under /app. With proxy set (development)
// every /app request is reverse-proxied to the Vite dev server, including
// HMR websockets. Otherwise files (the built client; may be nil) are served
// with an index.html fallback for client-side routes.
func MountSPA(e *echo.Echo, files fs.FS, proxy *url.URL) {
	var h echo.HandlerFunc
	if proxy != nil {
		slog.Info("serving "+SPAPrefix+" from the Vite dev server", "url", proxy.String())
		rp := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(proxy)
			r.SetXForwarded()
			// RemoveTrailingSlash has already turned /app/ into /app, but Vite
			// (base '/app/') only serves the root, and its HMR websocket, at /app/.
			if r.Out.URL.Path == SPAPrefix {
				r.Out.URL.Path = SPAPrefix + "/"
				r.Out.URL.RawPath = ""
			}
		}}
		h = echo.WrapHandler(rp)
	} else {
		s := &spa{files: files}
		s.built = files != nil && isFile(files, "index.html")
		if !s.built {
			slog.Warn("web client not built: " + SPAPrefix + " will answer 503 until `yarn build` is run")
		}
		h = s.serve
	}
	e.Any(SPAPrefix, h)
	e.Any(SPAPrefix+"/*", h)
}

type spa struct {
	files fs.FS
	built bool
}

func (s *spa) serve(c echo.Context) error {
	if m := c.Request().Method; m != http.MethodGet && m != http.MethodHead {
		return c.String(http.StatusMethodNotAllowed, "method not allowed")
	}
	if !s.built {
		c.Response().Header().Set("Cache-Control", cacheNone)
		return c.HTML(http.StatusServiceUnavailable, notBuiltPage)
	}

	rel := strings.TrimPrefix(strings.TrimPrefix(c.Request().URL.Path, SPAPrefix), "/")
	switch {
	case rel == "":
		return s.file(c, "index.html", cacheNone)
	case !fs.ValidPath(rel) || strings.Contains(rel, "\\"):
		return c.String(http.StatusNotFound, "not found")
	case rel == "assets" || strings.HasPrefix(rel, "assets/"):
		return s.file(c, rel, cacheImmutable)
	case path.Ext(rel) != "":
		return s.file(c, rel, cacheNone)
	default:
		return s.file(c, "index.html", cacheNone)
	}
}

// file writes name from the build, or a plain 404. It never falls back to
// index.html, so a missing hashed asset is visible rather than served as HTML.
func (s *spa) file(c echo.Context, name, cacheControl string) error {
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
	return c.Blob(http.StatusOK, ctype, b)
}

func isFile(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func isDir(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && info.IsDir()
}
