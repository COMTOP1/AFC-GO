package views

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/files"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

// DownloadFunc keeps /download?s=<code>&id=<n> working for old links and
// templates; it delegates to the same service as GET /api/v1/files.
func (v *Views) DownloadFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DownloadFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))

	id, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil || id < 1 {
		return v.error(http.StatusBadRequest, "id must be a positive integer",
			fmt.Errorf("download: invalid id %q", c.QueryParam("id")))
	}
	kind, ok := files.LegacyKind(c.QueryParam("s"))
	if !ok {
		return v.error(http.StatusBadRequest, "unknown download source",
			fmt.Errorf("download: unknown source %q", c.QueryParam("s")))
	}
	u, err := v.fileSvc.URL(c.Request().Context(), kind, id)
	if err != nil {
		if se, isSvc := svcerr.As(err); isSvc && se.Kind == svcerr.KindNotFound {
			return c.String(http.StatusNotFound, se.Message)
		}
		return errors.Join(errors.New("download failed"), err)
	}
	c.Response().Header().Set("Cache-Control", files.CacheControl(kind))
	return c.Redirect(http.StatusFound, u)
}
