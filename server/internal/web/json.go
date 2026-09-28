package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
)

// DateLayout is the date-only format API inputs use (dates without times,
// e.g. date of birth or date of event).
const DateLayout = "2006-01-02"

// BindJSON decodes the request body into dst, rejecting unknown fields.
func BindJSON(c echo.Context, dst any) error {
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return svcerr.InvalidField("body", "invalid JSON: "+err.Error())
	}
	return nil
}

// ParamID parses a positive integer path parameter.
func ParamID(c echo.Context, name string) (int, error) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id < 1 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}

// FormFile returns the uploaded file for field, or nil when none was sent.
func FormFile(c echo.Context, field string) (*upload.File, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
			return nil, nil //nolint:nilerr // uploads are optional; absence is not an error
		}
		return nil, svcerr.InvalidField(field, "invalid upload: "+err.Error())
	}
	return upload.FromHeader(fh), nil
}

// FormString returns the form value for field, or nil when the field was not
// sent at all (as opposed to sent empty).
//
// Unlike echo's own FormParams (which merges the URL query string into
// request.Form), this reads only the request body: request.PostForm holds
// urlencoded body values plus, after ParseMultipartForm, the multipart
// text fields. That keeps a query parameter of the same name from
// masquerading as a form field, which matters for PATCH's absent-vs-sent
// semantics (FormBool and FormDate rely on this too, via FormString).
func FormString(c echo.Context, field string) *string {
	req := c.Request()
	if strings.HasPrefix(req.Header.Get(echo.HeaderContentType), echo.MIMEMultipartForm) {
		if err := req.ParseMultipartForm(32 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) { // 32 MB, echo's own default
			return nil
		}
	} else if err := req.ParseForm(); err != nil {
		return nil
	}
	values, ok := req.PostForm[field]
	if !ok || len(values) == 0 {
		return nil
	}
	v := values[0]
	return &v
}

// FormBool parses an optional boolean form field ("true"/"false"/"1"/"0").
func FormBool(c echo.Context, field string) (*bool, error) {
	raw := FormString(c, field)
	if raw == nil || *raw == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(*raw)
	if err != nil {
		return nil, svcerr.InvalidField(field, "must be true or false")
	}
	return &b, nil
}

// FormDate parses an optional YYYY-MM-DD form field.
func FormDate(c echo.Context, field string) (*time.Time, error) {
	raw := FormString(c, field)
	if raw == nil || *raw == "" {
		return nil, nil
	}
	d, err := time.Parse(DateLayout, *raw)
	if err != nil {
		return nil, svcerr.InvalidField(field, "must be a date in YYYY-MM-DD format")
	}
	return &d, nil
}

// FormInt parses an optional integer form field.
func FormInt(c echo.Context, field string) (*int, error) {
	raw := FormString(c, field)
	if raw == nil || *raw == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(*raw)
	if err != nil {
		return nil, svcerr.InvalidField(field, field+" must be a whole number")
	}
	return &n, nil
}
