# Requirements Document

## Introduction

O projeto Garimpei opera como arquitetura multi-serviço em GCP Cloud Run multi-container, com 5 serviços (garimpei-api em C# .NET 10, collector/publisher/scheduler em Go, analyzer em Python FastAPI). Atualmente os logs são isolados por container no Cloud Logging, sem propagação de trace entre serviços e sem correlação — dificultando investigação de problemas em fluxos distribuídos.

Esta feature implementa **Observabilidade Unificada** usando OpenTelemetry SDKs em cada linguagem, Google-Built OpenTelemetry Collector como sidecar, e Cloud Trace + Cloud Logging como backends nativos. O objetivo é que qualquer requisição ou job agendado produza um trace distribuído completo (com trace_id e span_id em todos os logs), permitindo ao desenvolvedor encontrar TODOS os logs e spans de uma operação em uma única query no Log Explorer ou Cloud Trace.

**Decisões já tomadas:**
- Backend nativo GCP (Cloud Trace + Cloud Logging) — sem Grafana/Loki
- OpenTelemetry SDKs para instrumentação em cada linguagem
- Google-Built OTel Collector como sidecar no Cloud Run
- W3C traceparent como header de propagação de contexto
- 100% sampling inicial (configurável via env var)
- Frontend fora de escopo (spec futura)

**Fluxos-chave a rastrear:**
1. Salvar busca com cron → API → Scheduler.SetSchedule
2. Cron tick → Scheduler → Collector.Fetch → Cloud Tasks enqueue
3. Cloud Tasks callback → /process-alert → Analyzer /quedas → Publisher.Publish
4. Publicar oferta → API → Collector.GenerateAffiliateLink → Publisher.Publish
5. Busca frontend → API → Collector.Fetch → response

## Glossary

- **Garimpei_API**: Serviço C# .NET 10 (porta 8080) — ingress REST API que recebe requisições externas e orquestra chamadas aos demais serviços.
- **Collector_Service**: Serviço Go (gRPC porta 50051) — busca produtos de marketplaces (Shopee/Amazon).
- **Publisher_Service**: Serviço Go (gRPC porta 50052) — envia mensagens para Telegram/WhatsApp.
- **Scheduler_Service**: Serviço Go (gRPC porta 50054, HTTP porta 8054) — cron jobs e processamento de alertas via Cloud Tasks.
- **Analyzer_Service**: Serviço Python FastAPI (porta 8060) — queries analíticas no BigQuery.
- **OTel_Collector_Sidecar**: Google-Built OpenTelemetry Collector rodando como container sidecar no Cloud Run — recebe telemetria via OTLP e exporta para Cloud Trace/Logging.
- **Trace**: Representação completa de uma operação distribuída — composta por múltiplos spans correlacionados por um trace_id único.
- **Span**: Unidade de trabalho dentro de um trace — tem nome, duração, atributos e parent span opcional.
- **Trace_ID**: Identificador único de 128 bits (hex 32 chars) que correlaciona todos os spans e logs de uma operação distribuída.
- **Span_ID**: Identificador único de 64 bits (hex 16 chars) que identifica um span individual dentro de um trace.
- **W3C_Traceparent**: Header HTTP padrão W3C (`traceparent: 00-{trace_id}-{span_id}-{flags}`) usado para propagar contexto de trace entre serviços.
- **OTLP**: OpenTelemetry Protocol — protocolo padrão para exportar traces, metrics e logs do SDK para o collector.
- **Sampling_Rate**: Fração de traces capturados (0.0 a 1.0) — controlada por variável de ambiente.
- **Cloud_Trace**: Serviço GCP que armazena e visualiza traces distribuídos.
- **Cloud_Logging**: Serviço GCP que armazena logs estruturados — permite filtro por trace_id via Log Explorer.
- **Log_Correlation**: Associação automática de log entries com trace_id e span_id, permitindo navegação bidirecional entre logs e traces.
- **Context_Propagation**: Mecanismo de passagem do trace context (trace_id + span_id) entre serviços via headers HTTP ou metadata gRPC.

## Requirements

### Requirement 1: Instrumentação OpenTelemetry no Garimpei_API (C# .NET 10)

**User Story:** As a developer, I want the C# API to automatically generate traces for every incoming HTTP request, so that all operations have a trace_id from the entry point.

#### Acceptance Criteria

1. WHEN an HTTP request arrives at Garimpei_API, THE Garimpei_API SHALL create a root span with the request method and route as span name.
2. WHEN a W3C_Traceparent header is present on an incoming request, THE Garimpei_API SHALL use the received trace_id and parent span_id instead of generating new ones.
3. WHEN Garimpei_API makes an outgoing gRPC call to Collector_Service, Publisher_Service, or Scheduler_Service, THE Garimpei_API SHALL inject the W3C_Traceparent into the gRPC metadata.
4. WHEN Garimpei_API makes an outgoing HTTP call to Analyzer_Service, THE Garimpei_API SHALL inject the W3C_Traceparent into the HTTP request headers.
5. THE Garimpei_API SHALL export all spans via OTLP to the OTel_Collector_Sidecar endpoint (localhost:4317).
6. THE Garimpei_API SHALL configure the OpenTelemetry SDK via the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable defaulting to `http://localhost:4317`.
7. THE Garimpei_API SHALL set the service name resource attribute to "garimpei-api" via `OTEL_SERVICE_NAME` environment variable.

### Requirement 2: Instrumentação OpenTelemetry nos Serviços Go (Collector, Publisher, Scheduler)

**User Story:** As a developer, I want the Go services to participate in distributed traces, so that gRPC calls between services appear as child spans in Cloud Trace.

#### Acceptance Criteria

1. WHEN a gRPC request arrives at Collector_Service, Publisher_Service, or Scheduler_Service, THE service SHALL extract the W3C_Traceparent from gRPC metadata and create a child span.
2. WHEN Scheduler_Service makes outgoing gRPC calls to Collector_Service or Publisher_Service, THE Scheduler_Service SHALL propagate the trace context via gRPC metadata.
3. WHEN Scheduler_Service makes an outgoing HTTP call to Analyzer_Service, THE Scheduler_Service SHALL inject the W3C_Traceparent header into the HTTP request.
4. WHEN Scheduler_Service enqueues a Cloud Tasks request, THE Scheduler_Service SHALL include the W3C_Traceparent header in the Cloud Tasks HTTP target headers.
5. WHEN Scheduler_Service receives a Cloud Tasks HTTP callback on /process-alert, THE Scheduler_Service SHALL extract the W3C_Traceparent from the incoming HTTP headers and continue the trace.
6. THE Go services SHALL export spans via OTLP to the OTel_Collector_Sidecar endpoint configured by `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable defaulting to `http://localhost:4317`.
7. THE Go services SHALL use `otelgrpc` interceptors (UnaryInterceptor and StreamInterceptor) for automatic gRPC span creation on both client and server sides.

### Requirement 3: Instrumentação OpenTelemetry no Analyzer_Service (Python FastAPI)

**User Story:** As a developer, I want the Python analyzer to participate in distributed traces, so that analytics queries appear as spans correlated to the original request.

#### Acceptance Criteria

1. WHEN an HTTP request arrives at Analyzer_Service, THE Analyzer_Service SHALL extract the W3C_Traceparent from HTTP headers and create a child span.
2. WHEN Analyzer_Service executes a BigQuery query, THE Analyzer_Service SHALL create a child span with the query name as span attribute.
3. THE Analyzer_Service SHALL export spans via OTLP to the OTel_Collector_Sidecar endpoint configured by `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable defaulting to `http://localhost:4317`.
4. THE Analyzer_Service SHALL use the `opentelemetry-instrumentation-fastapi` package for automatic HTTP span creation.
5. THE Analyzer_Service SHALL set the service name resource attribute to "analyzer" via `OTEL_SERVICE_NAME` environment variable.

### Requirement 4: Logs Estruturados Correlacionados com Trace

**User Story:** As a developer, I want every log entry to include trace_id and span_id automatically, so that I can filter all logs of a distributed operation in Cloud Logging Log Explorer.

#### Acceptance Criteria

1. WHILE a span is active in the Go services, THE slog logger SHALL automatically include `logging.googleapis.com/trace` and `logging.googleapis.com/spanId` fields in every log entry.
2. WHILE a span is active in Garimpei_API, THE ILogger SHALL automatically include `logging.googleapis.com/trace` and `logging.googleapis.com/spanId` fields in every log entry.
3. WHILE a span is active in Analyzer_Service, THE Python logger SHALL automatically include `logging.googleapis.com/trace` and `logging.googleapis.com/spanId` fields in every log entry.
4. THE log field `logging.googleapis.com/trace` SHALL use the format `projects/{project_id}/traces/{trace_id}` as required by Cloud Logging for trace correlation.
5. THE log field `logging.googleapis.com/spanId` SHALL contain the hex-encoded 64-bit span_id of the current active span.
6. WHEN no span is active (background operations without incoming request), THE services SHALL emit logs without trace correlation fields rather than failing.

### Requirement 5: OTel Collector Sidecar no Cloud Run

**User Story:** As a developer, I want a Google-Built OTel Collector sidecar in the Cloud Run deployment, so that telemetry is exported reliably to Cloud Trace and Cloud Logging without SDK-level exporters.

#### Acceptance Criteria

1. THE Cloud Run deployment manifest SHALL include a container named "otel-collector" using the Google-Built OpenTelemetry Collector image.
2. THE OTel_Collector_Sidecar SHALL listen on port 4317 (gRPC OTLP) for receiving telemetry from application containers.
3. THE OTel_Collector_Sidecar SHALL export traces to Cloud Trace and logs to Cloud Logging using the Google Cloud exporter.
4. THE Cloud Run container dependency configuration SHALL ensure application containers start after the OTel_Collector_Sidecar is ready.
5. THE OTel_Collector_Sidecar SHALL have resource limits of at most 0.25 CPU and 128Mi memory.
6. THE OTel_Collector_Sidecar SHALL include a health check endpoint for startup probe validation.

### Requirement 6: Propagação de Contexto via Cloud Tasks

**User Story:** As a developer, I want traces to survive Cloud Tasks async boundaries, so that the entire alert processing pipeline (enqueue → callback → analyzer → publisher) appears as a single trace.

#### Acceptance Criteria

1. WHEN Scheduler_Service enqueues a Cloud Tasks HTTP request, THE Scheduler_Service SHALL include a `traceparent` header in the task HTTP request headers containing the current trace context.
2. WHEN Scheduler_Service receives the Cloud Tasks HTTP callback on /process-alert, THE Scheduler_Service SHALL parse the `traceparent` header from the request and continue the existing trace with a new child span.
3. WHEN the /process-alert handler calls Analyzer_Service and Publisher_Service, THE Scheduler_Service SHALL propagate the restored trace context to those downstream calls.
4. IF the Cloud Tasks request does not contain a `traceparent` header (legacy tasks), THEN THE Scheduler_Service SHALL create a new root trace for that operation.

### Requirement 7: Sampling Configurável

**User Story:** As a developer, I want to control the trace sampling rate via environment variable, so that I can reduce cost in high-traffic scenarios without redeploying.

#### Acceptance Criteria

1. THE services SHALL read the sampling rate from the `OTEL_TRACES_SAMPLER_ARG` environment variable, interpreting it as a float between 0.0 and 1.0.
2. WHEN `OTEL_TRACES_SAMPLER_ARG` is not set, THE services SHALL default to 1.0 (100% sampling).
3. WHEN `OTEL_TRACES_SAMPLER_ARG` is set to a value between 0.0 and 1.0, THE services SHALL use a parent-based trace-id-ratio sampler with that rate.
4. THE sampling decision SHALL be propagated — child services SHALL respect the parent sampling decision received via W3C_Traceparent rather than making independent sampling decisions.

### Requirement 8: Atributos de Span para Identificação de Operações

**User Story:** As a developer, I want spans to carry business-relevant attributes, so that I can search traces by tenant, busca_id, or alert type in Cloud Trace.

#### Acceptance Criteria

1. WHEN Garimpei_API processes a request from an authenticated user, THE root span SHALL include the attribute `garimpei.tenant_id` with the user's tenant identifier.
2. WHEN Scheduler_Service processes a cron job, THE span SHALL include attributes `garimpei.job_name` and `garimpei.busca_id` identifying the specific scheduled search.
3. WHEN Scheduler_Service processes an alert via /process-alert, THE span SHALL include attributes `garimpei.alert_type` and `garimpei.busca_id`.
4. WHEN Collector_Service fetches products, THE span SHALL include attribute `garimpei.marketplace` with the source marketplace name (shopee, amazon).
5. WHEN Publisher_Service publishes a message, THE span SHALL include attributes `garimpei.channel` (telegram, whatsapp) and `garimpei.chat_id`.
6. IF a span results in an error, THEN THE span SHALL record the error status with the exception message as an event.

### Requirement 9: Overhead de Latência Mínimo

**User Story:** As a developer, I want the observability instrumentation to add negligible latency, so that user-facing response times remain unchanged.

#### Acceptance Criteria

1. THE OpenTelemetry instrumentation SHALL add less than 5 milliseconds of latency per request to the end-to-end response time.
2. THE services SHALL export telemetry asynchronously using batch span processors — span export SHALL NOT block the request processing path.
3. THE OTel_Collector_Sidecar SHALL use batch processing for exports to Cloud Trace and Cloud Logging.
4. IF the OTel_Collector_Sidecar is unavailable, THEN THE services SHALL continue processing requests normally without failing — telemetry loss is acceptable.

### Requirement 10: Compatibilidade com Testes e Deployment Existentes

**User Story:** As a developer, I want the observability changes to not break existing tests or deployment, so that the feature can be integrated incrementally without risk.

#### Acceptance Criteria

1. WHEN the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable is not set and no OTel Collector is reachable, THE services SHALL start and operate normally without errors — observability degrades gracefully.
2. THE Go services unit tests SHALL continue passing without requiring an OTel Collector running — the OTel SDK SHALL be initialized with a no-op exporter in test environments.
3. THE C# integration tests SHALL continue passing without requiring an OTel Collector — the OTel SDK SHALL not be registered when `ASPNETCORE_ENVIRONMENT` is "Testing".
4. THE Python tests SHALL continue passing without requiring an OTel Collector — the OTel SDK SHALL use a no-op tracer when `OTEL_EXPORTER_OTLP_ENDPOINT` is not configured.
5. THE existing CI pipeline SHALL not require any new infrastructure or services to pass.
6. THE deployment manifest changes SHALL be backward-compatible — the sidecar container SHALL be additive without modifying existing container configurations.

### Requirement 11: Inicialização Compartilhada do OTel SDK (Go)

**User Story:** As a developer, I want a shared OTel initialization package for the Go services, so that collector, publisher, and scheduler use consistent configuration without code duplication.

#### Acceptance Criteria

1. THE project SHALL provide a shared Go package (internal/otel or similar) that initializes the OpenTelemetry TracerProvider with OTLP exporter, resource attributes, and sampler configuration.
2. THE shared package SHALL accept service name as parameter and configure the `service.name` resource attribute accordingly.
3. WHEN the shared initialization function is called, THE package SHALL register gRPC interceptors for both client and server, and an HTTP transport wrapper for outgoing HTTP calls.
4. THE shared package SHALL provide a shutdown function that flushes pending spans before process exit.
5. THE shared package SHALL configure the slog handler to automatically inject trace correlation fields when a span is active in the context.
