// Package otel provides shared OpenTelemetry initialization for all Go services.
//
// Uses OTLP gRPC exporter pointing directly to Google Cloud Telemetry API.
// No sidecar collector needed — traces go straight to Cloud Trace.
//
// Environment variables:
//   - OTEL_EXPORTER_OTLP_ENDPOINT: override endpoint (default: Cloud Trace via ADC)
//   - OTEL_SERVICE_NAME: fallback for service name
//   - OTEL_TRACES_SAMPLER_ARG: sampling rate (default: 1.0)
//   - GCP_PROJECT_ID: for log correlation format
package otel

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Init initializes OpenTelemetry for the service.
// Returns a shutdown function that flushes pending spans.
//
// In Cloud Run: exports directly to Google Cloud Telemetry API (no sidecar needed).
// Locally: if OTEL_EXPORTER_OTLP_ENDPOINT is set, uses that (e.g. Jaeger localhost:4317).
// If neither is available: uses no-op exporter (graceful degradation).
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	if envName := os.Getenv("OTEL_SERVICE_NAME"); envName != "" {
		serviceName = envName
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: create resource: %w", err)
	}

	spanExporter, err := buildExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel: create exporter: %w", err)
	}

	sampler := buildSampler()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(spanExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// GRPCServerInterceptors returns server options with OTel interceptors.
func GRPCServerInterceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}
}

// GRPCDialOptions returns dial options with OTel client interceptors.
func GRPCDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

// HTTPTransport returns an http.RoundTripper that propagates trace context.
func HTTPTransport() http.RoundTripper {
	return otelhttp.NewTransport(http.DefaultTransport)
}

// buildExporter creates the appropriate span exporter based on environment.
func buildExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	if endpoint != "" {
		// Explicit endpoint (local dev: Jaeger, or custom collector)
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpointURL(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("otel: create local exporter: %w", err)
		}
		return exp, nil
	}

	// In Cloud Run: export directly to Google Cloud Telemetry API
	// Uses Application Default Credentials (ADC) automatically.
	gcpEndpoint := "telemetry.googleapis.com:443"
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(gcpEndpoint),
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
	)
	if err != nil {
		// Can't reach GCP (local without ADC) — fallback to no-op
		return noopExporter{}, nil //nolint:nilerr // intentional fallback
	}
	return exporter, nil
}

// buildSampler reads OTEL_TRACES_SAMPLER_ARG and builds the appropriate sampler.
func buildSampler() sdktrace.Sampler {
	rateStr := os.Getenv("OTEL_TRACES_SAMPLER_ARG")
	if rateStr == "" {
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}

	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil || rate < 0 || rate > 1 {
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}

	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(rate))
}

// noopExporter discards all spans (used when no exporter is available).
type noopExporter struct{}

func (noopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }
func (noopExporter) Shutdown(context.Context) error                             { return nil }
