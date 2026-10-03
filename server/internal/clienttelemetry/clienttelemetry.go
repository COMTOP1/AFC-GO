// Package clienttelemetry ingests best-effort telemetry events reported by
// the React client (API call outcomes, unhandled errors) and re-emits them
// as OpenTelemetry spans so they show up alongside server-side traces.
package clienttelemetry

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/clienttelemetry")

// maxAttributes and maxStringLen bound what an untrusted browser can make
// us attach to a span.
const (
	maxAttributes = 20
	maxStringLen  = 2000
)

// Event is one client-reported occurrence: either a completed API call or
// an unhandled error.
type Event struct {
	// Type is "api" or "error".
	Type string `json:"type"`
	// Name identifies what happened, e.g. "GET /site" or "unhandled".
	Name string `json:"name"`
	// StartTime and EndTime are Unix milliseconds. EndTime is zero for
	// instantaneous events (errors), in which case the span is zero-length.
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime,omitempty"`
	// Traceparent is the W3C traceparent header the client sent with the
	// matching API request, if any, so this event joins the same trace.
	Traceparent string `json:"traceparent,omitempty"`
	// Attributes are short, string-valued tags (status code, path, etc).
	Attributes map[string]string `json:"attributes,omitempty"`
	Message    string            `json:"message,omitempty"`
	Stack      string            `json:"stack,omitempty"`
}

// Record turns ev into a span, honouring Traceparent so it joins whatever
// trace the client's matching API request belongs to.
func Record(ctx context.Context, ev Event) {
	if ev.Traceparent != "" {
		ctx = propagation.TraceContext{}.Extract(ctx, propagation.MapCarrier{"traceparent": ev.Traceparent})
	}

	start := msToTime(ev.StartTime)
	end := start
	if ev.EndTime != 0 {
		end = msToTime(ev.EndTime)
	}

	_, span := tracer.Start(ctx, ev.Name, oteltrace.WithTimestamp(start))
	defer span.End(oteltrace.WithTimestamp(end))

	n := 0
	for k, v := range ev.Attributes {
		if n >= maxAttributes {
			break
		}
		span.SetAttributes(attribute.String("client."+k, truncate(v)))
		n++
	}

	if ev.Type == "error" {
		msg := truncate(ev.Message)
		span.SetStatus(codes.Error, msg)
		span.RecordError(errors.New(msg), oteltrace.WithAttributes(
			attribute.String("exception.stacktrace", truncate(ev.Stack)),
		))
	}
}

func msToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}

func truncate(s string) string {
	if len(s) > maxStringLen {
		return s[:maxStringLen]
	}
	return s
}
