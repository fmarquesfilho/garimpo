package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestInit_NoEndpoint_DoesNotFail(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := Init(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("Init should not fail without endpoint: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown func should not be nil")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown should not fail: %v", err)
	}
}

func TestInit_WithServiceName(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_SERVICE_NAME", "override-name")

	shutdown, err := Init(context.Background(), "original-name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer shutdown(context.Background()) //nolint:errcheck
}

func TestSlogHandler_WithoutSpan_NoTraceFields(t *testing.T) {
	// Test that the handler can be created and used without panic
	logger := slog.New(SlogHandler("test-project"))
	logger.Info("test message without span context")
}

func TestSlogHandler_WithSpan_InjectsTraceFields(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, nil)
	handler := &traceHandler{base: base, projectID: "garimpo-500114"}

	logger := slog.New(handler)
	logger.InfoContext(ctx, "test with trace")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	traceField, ok := entry["logging.googleapis.com/trace"].(string)
	if !ok {
		t.Fatal("expected logging.googleapis.com/trace field")
	}
	expected := "projects/garimpo-500114/traces/4bf92f3577b34da6a3ce929d0e0e4736"
	if traceField != expected {
		t.Errorf("trace field = %q, want %q", traceField, expected)
	}

	spanField, ok := entry["logging.googleapis.com/spanId"].(string)
	if !ok {
		t.Fatal("expected logging.googleapis.com/spanId field")
	}
	if spanField != "00f067aa0ba902b7" {
		t.Errorf("spanId = %q, want %q", spanField, "00f067aa0ba902b7")
	}
}

func TestGRPCServerInterceptors_NotEmpty(t *testing.T) {
	opts := GRPCServerInterceptors()
	if len(opts) == 0 {
		t.Error("expected at least one server option")
	}
}

func TestGRPCDialOptions_NotEmpty(t *testing.T) {
	opts := GRPCDialOptions()
	if len(opts) == 0 {
		t.Error("expected at least one dial option")
	}
}

func TestHTTPTransport_NotNil(t *testing.T) {
	transport := HTTPTransport()
	if transport == nil {
		t.Error("expected non-nil transport")
	}
}
