// Package telemetry wires up OpenTelemetry tracing and logging for the app.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Config controls where telemetry is exported to.
type Config struct {
	ServiceName    string
	ServiceVersion string

	// Endpoint is the OTLP/HTTP collector endpoint, e.g. "http://localhost:4318"
	// or "https://collector.example.com". If empty, OTLP export is disabled:
	// tracer.Start calls throughout the app fall back to OpenTelemetry's
	// built-in no-op tracer, and logging stays stdout-only.
	Endpoint string

	// Headers are sent with every OTLP export request, e.g. for collector
	// authentication.
	Headers map[string]string
}

// Setup wires up OpenTelemetry tracing and logging and installs an
// slog.Logger as the process default. It always logs to stdout; when
// cfg.Endpoint is set it additionally exports traces and logs over OTLP/HTTP.
//
// The returned shutdown func flushes and closes any exporters that were
// created; call it (typically deferred) before the process exits.
func Setup(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var errs error
		for _, fn := range shutdownFuncs {
			errs = errors.Join(errs, fn(ctx))
		}
		return errs
	}

	handlers := []slog.Handler{slog.NewTextHandler(os.Stdout, nil)}

	if cfg.Endpoint != "" {
		endpointURL, urlErr := url.Parse(cfg.Endpoint)
		if urlErr != nil {
			return shutdown, fmt.Errorf("failed to parse otlp endpoint %q: %w", cfg.Endpoint, urlErr)
		}
		// WithEndpoint (host[:port], no scheme/path) lets each exporter append
		// its own default signal path (/v1/traces, /v1/logs). WithEndpointURL
		// would take the path as-is instead, defaulting to "/" and 404ing.
		host := endpointURL.Host
		insecure := endpointURL.Scheme != "https"

		res, resErr := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceName(cfg.ServiceName),
				semconv.ServiceVersion(cfg.ServiceVersion),
			),
		)
		if resErr != nil {
			return shutdown, fmt.Errorf("failed to build otel resource: %w", resErr)
		}

		traceOpts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(host),
			otlptracehttp.WithHeaders(cfg.Headers),
		}
		if insecure {
			traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
		}
		traceExporter, tErr := otlptracehttp.New(ctx, traceOpts...)
		if tErr != nil {
			return shutdown, fmt.Errorf("failed to create otlp trace exporter: %w", tErr)
		}
		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)

		logOpts := []otlploghttp.Option{
			otlploghttp.WithEndpoint(host),
			otlploghttp.WithHeaders(cfg.Headers),
		}
		if insecure {
			logOpts = append(logOpts, otlploghttp.WithInsecure())
		}
		logExporter, lErr := otlploghttp.New(ctx, logOpts...)
		if lErr != nil {
			return shutdown, fmt.Errorf("failed to create otlp log exporter: %w", lErr)
		}
		loggerProvider := sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
			sdklog.WithResource(res),
		)
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)

		handlers = append(handlers, otelslog.NewHandler(cfg.ServiceName, otelslog.WithLoggerProvider(loggerProvider)))
	}

	slog.SetDefault(slog.New(slog.NewMultiHandler(handlers...)))

	return shutdown, nil
}

// ParseHeaders parses the OTLP headers env-var format ("key1=value1,key2=value2")
// into a header map suitable for Config.Headers.
func ParseHeaders(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	headers := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return headers
}
