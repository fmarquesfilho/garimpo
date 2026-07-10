# Tasks

## Task 1: Criar pacote compartilhado `internal/otel` (Go)
- **Requirements**: R2, R4, R7, R9, R10, R11
- **Files**:
  - `internal/otel/otel.go` (new)
  - `internal/otel/sloghandler.go` (new)
  - `internal/otel/otel_test.go` (new)
  - `go.mod` (add OTel SDK dependencies)
- [ ] Implementar `Init(ctx, serviceName)` que retorna shutdown func
- [ ] Configurar TracerProvider com OTLP gRPC exporter (batch, async)
- [ ] Configurar sampler ParentBased(TraceIDRatioBased) lendo `OTEL_TRACES_SAMPLER_ARG`
- [ ] Implementar `GRPCServerInterceptors()` com otelgrpc.UnaryServerInterceptor + StreamServerInterceptor
- [ ] Implementar `GRPCDialOptions()` com otelgrpc client interceptors
- [ ] Implementar `HTTPTransport()` com otelhttp.NewTransport (propaga traceparent)
- [ ] Implementar `SlogHandler(projectID)` que wrapa JSONHandler e injeta `logging.googleapis.com/trace` + `spanId`
- [ ] Graceful degradation: se OTEL_EXPORTER_OTLP_ENDPOINT não set → NoOp exporter
- [ ] Testes unitários: Init sem collector não falha, SlogHandler injeta fields corretos
- [ ] `go mod tidy` passa

## Task 2: Instrumentar Collector Service (Go)
- **Requirements**: R2, R4, R8, R11
- **Files**:
  - `services/collector/main.go` (modify)
  - `services/collector/server.go` (modify)
- [ ] Importar e chamar `otel.Init(ctx, "collector")` no main
- [ ] Usar `otel.GRPCServerInterceptors()` no `grpc.NewServer()`
- [ ] Substituir `slog.SetDefault` pelo logger com `otel.SlogHandler`
- [ ] Adicionar `garimpei.marketplace` attribute nos spans de Fetch/FetchShop
- [ ] Defer `shutdown(ctx)` no signal handler
- [ ] Go tests continuam passando sem collector

## Task 3: Instrumentar Publisher Service (Go)
- **Requirements**: R2, R4, R8, R11
- **Files**:
  - `services/publisher/main.go` (modify)
  - `services/publisher/server.go` (modify)
- [ ] Importar e chamar `otel.Init(ctx, "publisher")` no main
- [ ] Usar `otel.GRPCServerInterceptors()` no `grpc.NewServer()`
- [ ] Substituir `slog.SetDefault` pelo logger com `otel.SlogHandler`
- [ ] Adicionar `garimpei.channel` e `garimpei.chat_id` attributes nos spans de Publish
- [ ] Defer `shutdown(ctx)` no signal handler
- [ ] Go tests continuam passando sem collector

## Task 4: Instrumentar Scheduler Service (Go)
- **Requirements**: R2, R4, R6, R8, R11
- **Files**:
  - `services/scheduler/main.go` (modify)
  - `services/scheduler/server.go` (modify)
  - `services/scheduler/jobs.go` (modify)
  - `services/scheduler/alerts.go` (modify)
  - `internal/taskqueue/taskqueue.go` (modify)
- [ ] Importar e chamar `otel.Init(ctx, "scheduler")` no main
- [ ] Usar `otel.GRPCServerInterceptors()` no server + `otel.GRPCDialOptions()` nos clients
- [ ] Usar `otel.HTTPTransport()` para chamadas HTTP ao Analyzer
- [ ] Substituir slog pelo handler com trace correlation
- [ ] Criar span explícito em `dispatchJob` com attributes `garimpei.job_name`, `garimpei.busca_id`
- [ ] Propagar traceparent no Cloud Tasks: adicionar header no `taskspb.HttpRequest.Headers`
- [ ] Extrair traceparent no `handleProcessAlert`: usar `propagation.Extract` dos headers HTTP
- [ ] Propagar contexto restaurado nas chamadas downstream (Analyzer, Publisher)
- [ ] Adicionar `garimpei.alert_type` attribute no span de processamento de alerta
- [ ] Go tests continuam passando sem collector

## Task 5: Instrumentar Garimpei API (C# .NET)
- **Requirements**: R1, R4, R8, R9, R10
- **Files**:
  - `src/Garimpei.Api/Program.cs` (modify)
  - `src/Garimpei.Api/Garimpei.Api.csproj` (add GrpcNetClient instrumentation)
- [ ] Adicionar pacote `OpenTelemetry.Instrumentation.GrpcNetClient`
- [ ] Configurar `AddOpenTelemetry()` com tracing (AspNetCore + Http + GrpcClient + OTLP exporter)
- [ ] Configurar resource com service name "garimpei-api"
- [ ] Configurar CORS para aceitar header `traceparent` e `tracestate`
- [ ] Adicionar middleware para enriquecer span com `garimpei.tenant_id` (do user autenticado)
- [ ] Configurar ILogger JSON format para incluir `logging.googleapis.com/trace` quando Activity.Current existe
- [ ] Não registrar OTel quando `ASPNETCORE_ENVIRONMENT == "Testing"` (compatibilidade testes)
- [ ] Testes C# continuam passando (72/72)
- [ ] Build + deploy não quebra

## Task 6: Instrumentar Analyzer Service (Python)
- **Requirements**: R3, R4, R8, R10
- **Files**:
  - `services/analyzer/otel_setup.py` (new)
  - `services/analyzer/requirements.txt` (modify)
  - `services/analyzer/main.py` (modify)
- [ ] Adicionar dependências OTel no requirements.txt
- [ ] Criar `otel_setup.py` com `init_otel(service_name)`
- [ ] Configurar TracerProvider + OTLP gRPC exporter + batch processor
- [ ] Instrumentar FastAPI automaticamente via `opentelemetry-instrumentation-fastapi`
- [ ] Configurar logging para incluir trace fields formato GCP
- [ ] Criar spans manuais para queries BigQuery (com nome da query como attribute)
- [ ] Graceful degradation: sem OTEL_EXPORTER_OTLP_ENDPOINT → NoOp
- [ ] Chamar `init_otel("analyzer")` antes de criar o app FastAPI
- [ ] `python -c "import main"` continua funcionando (CI syntax check)

## Task 7: OTel Collector Sidecar no Cloud Run
- **Requirements**: R5, R9
- **Files**:
  - `deploy/cloud-run-deploy-now.yaml` (modify)
- [ ] Adicionar container `otel-collector` com imagem Google-Built
- [ ] Configurar via env var `OTEL_COLLECTOR_CONFIG` inline
- [ ] Receivers: otlp (gRPC :4317)
- [ ] Processors: batch + resourcedetection (gcp)
- [ ] Exporters: googlecloud (project: garimpo-500114)
- [ ] Resources: 0.25 CPU, 128Mi memory
- [ ] StartupProbe: tcpSocket port 4317
- [ ] Atualizar container-dependencies para incluir otel-collector
- [ ] Adicionar `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317` a todos os containers
- [ ] Adicionar `OTEL_SERVICE_NAME` a cada container
- [ ] Deploy funciona sem rollback necessário (sidecar é aditivo)

## Task 8: Frontend — Propagação de Trace (Browser)
- **Requirements**: R1 (AC2), R9
- **Files**:
  - `web/src/lib/telemetry.js` (new)
  - `web/src/routes/+layout.js` (modify — import telemetry)
  - `web/package.json` (add @opentelemetry/* dependencies)
- [ ] Instalar `@opentelemetry/sdk-trace-web`, `@opentelemetry/instrumentation-fetch`, `@opentelemetry/core`
- [ ] Criar `telemetry.js` com WebTracerProvider + FetchInstrumentation (propagação only, sem exporter)
- [ ] Configurar `propagateTraceHeaderCorsUrls` para garimpei.app.br + localhost
- [ ] Importar `telemetry.js` no `+layout.js` (import side-effect, browser only)
- [ ] Verificar que todos os fetch em `api.js` agora incluem header `traceparent`
- [ ] Vitest unit tests continuam passando (298)
- [ ] Build SPA continua funcionando com adapter-static

## Task 9: Frontend — Migrar adapter-static → adapter-cloudflare
- **Requirements**: R1 (futuro), R10
- **Files**:
  - `web/svelte.config.js` (modify)
  - `web/package.json` (swap adapter dependency)
  - `.github/workflows/ci.yml` (adjust deploy-web if needed)
- [ ] `npm remove @sveltejs/adapter-static && npm install @sveltejs/adapter-cloudflare`
- [ ] Atualizar svelte.config.js: `import adapter from '@sveltejs/adapter-cloudflare'`
- [ ] Remover config de pages/assets/fallback (adapter-cloudflare auto-configura)
- [ ] Verificar que build produz output compatível com `wrangler pages deploy`
- [ ] Testar localmente com `npm run dev` (deve funcionar igual)
- [ ] E2E local continuam passando (24/24)
- [ ] CI deploy-web continua funcionando

## Task 10: Frontend — instrumentation.server.js (com adapter-cloudflare)
- **Requirements**: R1, R8
- **Files**:
  - `web/svelte.config.js` (add experimental flags)
  - `web/src/instrumentation.server.js` (new)
  - `web/package.json` (add @opentelemetry/sdk-node if needed)
- [ ] Adicionar `experimental: { tracing: { server: true }, instrumentation: { server: true } }`
- [ ] Criar `src/instrumentation.server.js` com OTel SDK setup
- [ ] Configurar trace exporter (OTLP HTTP para dev local, propagation-only em prod CF Workers)
- [ ] Verificar traces aparecem no Jaeger local (`docker run jaegertracing/all-in-one`)
- [ ] Enriquecer root span com `garimpei.tenant_id` via `event.tracing.root.setAttribute`
- [ ] Testes continuam passando

## Task 11: Validação E2E — Trace Distribuído Completo
- **Requirements**: R1-R6, R8
- **Files**:
  - `.mise/tasks/test/e2e-traces` (new)
- [ ] Criar mise task que valida trace propagation em produção
- [ ] Fazer request com traceparent → verificar que Cloud Trace mostra spans de múltiplos serviços
- [ ] Verificar que Cloud Logging filtra logs por trace_id
- [ ] Verificar que busca agendada gera trace (scheduler → collector → cloud tasks → publisher)
- [ ] Documentar como inspecionar traces no GCP Console

## Task 12: Documentação
- **Requirements**: Todas
- **Files**:
  - `docs/observability.md` (new)
  - `docs/06-qualidade-e-testes.md` (update)
- [ ] Documentar arquitetura de observabilidade (diagrama, fluxos)
- [ ] Documentar como ver traces no Cloud Trace Console
- [ ] Documentar como filtrar logs por trace_id no Log Explorer
- [ ] Documentar como rodar observabilidade local (Jaeger)
- [ ] Atualizar doc de qualidade com as novas tasks mise
- [ ] ADR-0028: decisão adapter-static → adapter-cloudflare
