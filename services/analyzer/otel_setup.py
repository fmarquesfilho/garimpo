"""OpenTelemetry initialization for the Analyzer service.

Configures tracing with OTLP gRPC export and FastAPI auto-instrumentation.
Graceful degradation: if OTEL_EXPORTER_OTLP_ENDPOINT is not set, uses NoOp tracer.
"""

import logging
import os

logger = logging.getLogger(__name__)


def init_otel(service_name: str = "analyzer") -> None:
    """Initialize OpenTelemetry SDK. Call BEFORE creating FastAPI app."""
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "")
    if not endpoint:
        logger.info("OTEL_EXPORTER_OTLP_ENDPOINT not set — OTel disabled (no-op)")
        return

    try:
        from opentelemetry import trace
        from opentelemetry.sdk.trace import TracerProvider
        from opentelemetry.sdk.trace.export import BatchSpanProcessor
        from opentelemetry.sdk.resources import Resource
        from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
        from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
        from opentelemetry.instrumentation.logging import LoggingInstrumentor

        resource = Resource.create({"service.name": os.environ.get("OTEL_SERVICE_NAME", service_name)})

        # Sampler: read rate from env (default 1.0 = 100%)
        sampler_arg = os.environ.get("OTEL_TRACES_SAMPLER_ARG", "1.0")
        try:
            rate = float(sampler_arg)
        except ValueError:
            rate = 1.0

        from opentelemetry.sdk.trace.sampling import TraceIdRatioBased, ParentBased
        sampler = ParentBased(root=TraceIdRatioBased(rate))

        provider = TracerProvider(resource=resource, sampler=sampler)
        exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
        provider.add_span_processor(BatchSpanProcessor(exporter))
        trace.set_tracer_provider(provider)

        # Auto-instrument FastAPI (will instrument any app created after this call)
        FastAPIInstrumentor.instrument()

        # Inject trace_id and span_id into Python log records
        LoggingInstrumentor().instrument(set_logging_format=True)

        # Configure logging format for GCP Cloud Logging correlation
        _configure_gcp_logging()

        logger.info(f"OpenTelemetry initialized: service={service_name}, endpoint={endpoint}")

    except ImportError as e:
        logger.warning(f"OTel packages not installed, tracing disabled: {e}")
    except Exception as e:
        logger.warning(f"OTel initialization failed, tracing disabled: {e}")


def _configure_gcp_logging():
    """Configure Python logging to include GCP trace correlation fields."""
    project_id = os.environ.get("GCP_PROJECT_ID", os.environ.get("BQ_PROJECT", ""))
    if not project_id:
        return

    import logging as _logging

    class GCPTraceFilter(_logging.Filter):
        def filter(self, record):
            from opentelemetry import trace
            span = trace.get_current_span()
            ctx = span.get_span_context()
            if ctx and ctx.is_valid:
                record.trace_id = f"projects/{project_id}/traces/{ctx.trace_id:032x}"
                record.span_id = f"{ctx.span_id:016x}"
            else:
                record.trace_id = ""
                record.span_id = ""
            return True

    root_logger = _logging.getLogger()
    root_logger.addFilter(GCPTraceFilter())
