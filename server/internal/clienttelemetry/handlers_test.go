package clienttelemetry_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/clienttelemetry"
	"github.com/COMTOP1/AFC-GO/server/internal/web"
	"github.com/COMTOP1/AFC-GO/server/internal/web/apitest"
)

func newClient() *apitest.Client {
	e := apitest.NewEcho()
	clienttelemetry.NewHandlers().Register(web.NewAPI(e, false))
	return apitest.New(e)
}

func TestIngest_AcceptsAnAPIEvent(t *testing.T) {
	rec := newClient().JSON(t, http.MethodPost, "/api/v1/telemetry", clienttelemetry.Event{
		Type:      "api",
		Name:      "GET /site",
		StartTime: 1000,
		EndTime:   1200,
		Attributes: map[string]string{
			"http.status_code": "200",
		},
	})
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestIngest_AcceptsAnErrorEvent(t *testing.T) {
	rec := newClient().JSON(t, http.MethodPost, "/api/v1/telemetry", clienttelemetry.Event{
		Type:      "error",
		Name:      "unhandled",
		StartTime: 1000,
		Message:   "TypeError: boom",
		Stack:     "at foo (bar.tsx:1)",
	})
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestIngest_RejectsMissingName(t *testing.T) {
	rec := newClient().JSON(t, http.MethodPost, "/api/v1/telemetry", clienttelemetry.Event{
		Type:      "api",
		StartTime: 1000,
	})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestIngest_RejectsMissingStartTime(t *testing.T) {
	rec := newClient().JSON(t, http.MethodPost, "/api/v1/telemetry", clienttelemetry.Event{
		Type: "api",
		Name: "GET /site",
	})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestIngest_RejectsUnknownType(t *testing.T) {
	rec := newClient().JSON(t, http.MethodPost, "/api/v1/telemetry", clienttelemetry.Event{
		Type:      "bogus",
		Name:      "GET /site",
		StartTime: 1000,
	})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestIngest_RejectsUnknownFields(t *testing.T) {
	rec := newClient().JSON(t, http.MethodPost, "/api/v1/telemetry", map[string]any{"unknownField": true})
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
