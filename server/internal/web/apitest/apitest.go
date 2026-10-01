// Package apitest drives the API in handler tests the way a same-origin
// browser would.
package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// NewEcho returns an Echo set up like production for API routes. Non-API
// errors write "legacy error" so tests can tell the two paths apart.
func NewEcho() *echo.Echo {
	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.HTTPErrorHandler = web.ErrorHandler()
	return e
}

// Client sends requests to an Echo instance, optionally with a session cookie.
type Client struct {
	e      *echo.Echo
	cookie *http.Cookie
}

func New(e *echo.Echo) *Client { return &Client{e: e} }

// As returns a client that sends cookie with every request.
func (c *Client) As(cookie *http.Cookie) *Client { return &Client{e: c.e, cookie: cookie} }

// Do serves req as a same-origin browser request.
func (c *Client) Do(t *testing.T, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	return rec
}

func (c *Client) Get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	return c.Do(t, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))
}

// JSON sends body encoded as JSON; a nil body sends no body.
func (c *Client) JSON(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, r)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.Do(t, req)
}

// FilePart is one file in a multipart request.
type FilePart struct {
	Field, Name, ContentType, Body string
}

// Multipart sends fields and files as multipart/form-data.
func (c *Client) Multipart(t *testing.T, method, path string, fields map[string]string, files ...FilePart) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}
	for _, f := range files {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, f.Field, f.Name))
		h.Set("Content-Type", f.ContentType)
		part, err := w.CreatePart(h)
		require.NoError(t, err)
		_, err = part.Write([]byte(f.Body))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequestWithContext(context.Background(), method, path, &buf)
	req.Header.Set(echo.HeaderContentType, w.FormDataContentType())
	return c.Do(t, req)
}

// Decode unmarshals the response body into T.
func Decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &v), "body: %s", rec.Body.String())
	return v
}

// ErrorOf decodes the JSON error envelope.
func ErrorOf(t *testing.T, rec *httptest.ResponseRecorder) web.APIError {
	t.Helper()
	return Decode[web.ErrorResponse](t, rec).Error
}
