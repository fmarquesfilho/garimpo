package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// SlogHandler returns a slog.Handler that injects GCP Cloud Logging trace
// correlation fields when a span is active in the context.
//
// Fields injected:
//   - logging.googleapis.com/trace: projects/{project_id}/traces/{trace_id}
//   - logging.googleapis.com/spanId: hex span_id
//   - severity: mapped from slog level
//
// When no span is active, logs are emitted without trace fields.
func SlogHandler(projectID string) slog.Handler {
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT_ID")
	}
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Map slog level to GCP severity
			if a.Key == slog.LevelKey {
				a.Key = "severity"
				level := a.Value.Any().(slog.Level)
				switch {
				case level < slog.LevelInfo:
					a.Value = slog.StringValue("DEBUG")
				case level < slog.LevelWarn:
					a.Value = slog.StringValue("INFO")
				case level < slog.LevelError:
					a.Value = slog.StringValue("WARNING")
				default:
					a.Value = slog.StringValue("ERROR")
				}
			}
			// Rename "msg" → "message" for Cloud Logging
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	})

	return &traceHandler{base: base, projectID: projectID}
}

type traceHandler struct {
	base      slog.Handler
	projectID string
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		sc := span.SpanContext()
		traceID := sc.TraceID().String()
		spanID := sc.SpanID().String()

		record.AddAttrs(
			slog.String("logging.googleapis.com/trace",
				fmt.Sprintf("projects/%s/traces/%s", h.projectID, traceID)),
			slog.String("logging.googleapis.com/spanId", spanID),
		)
	}
	return h.base.Handle(ctx, record)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{base: h.base.WithAttrs(attrs), projectID: h.projectID}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{base: h.base.WithGroup(name), projectID: h.projectID}
}
