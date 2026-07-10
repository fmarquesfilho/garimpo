// Package otel provides shared OpenTelemetry initialization for all Go services.
//
// Usage:
//
//	shutdown, err := otel.Init(ctx, "collector")
//	if err != nil { log.Fatal(err) }
//	defer shutdown(ctx)
//
// The package reads configuration from environment variables:
//   - OTEL_EXPORTER_OTLP_ENDPOINT (default: http://localhost:4317)
//   - OTEL_SERVICE_NAME (fallback: serviceName parameter)
//   - OTEL_TRACES_SAMPLER_ARG (default: 1.0 — 100% sampling)
//   - GCP_PROJECT_ID (for log correlation format)
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
)

// Init initializes OpenTelemetry for the service.
// Returns a shutdown function that flushes pending spans.
// If OTEL_EXPORTER_OTLP_ENDPOINT is not set, uses a no-op exporter (graceful degradation).
func Init(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	// Override service name from env if set
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

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	var spanExporter sdktrace.SpanExporter
	if endpoint == "" {
		// No collector configured — use no-op (spans are discarded)
		spanExporter = noopExporter{}
	} else {
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpointURL(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("otel: create OTLP exporter: %w", err)
		}
		spanExporter = exp
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

// noopExporter discards all spans (used when no collector is configured).
type noopExporter struct{}

func (noopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }
func (noopExporter) Shutdown(context.Context) error                             { return nil }
