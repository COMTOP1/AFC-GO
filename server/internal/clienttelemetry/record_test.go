package clienttelemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/COMTOP1/AFC-GO/server/internal/clienttelemetry"
)

func TestRecordRedactsResetTokens(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	clienttelemetry.Record(context.Background(), clienttelemetry.Event{
		Type:       "api",
		Name:       "POST /auth/reset/secret-token",
		StartTime:  1,
		Attributes: map[string]string{"http.path": "/auth/reset/secret-token"},
	})

	spans := rec.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "POST /auth/reset/:token", spans[0].Name())
	for _, a := range spans[0].Attributes() {
		assert.NotContains(t, a.Value.String(), "secret-token", "attribute %s leaks the token", a.Key)
	}
}
