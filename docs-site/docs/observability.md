# Observabilidade — Garimpei

**Atualizado:** 2026-07-10
**Stack:** OpenTelemetry (Go + C# + Python + Svelte) → Google Cloud Trace + Cloud Logging
**ADR:** [0028-adapter-cloudflare-otel](decisoes/0028-adapter-cloudflare-otel.md)

---

## Arquitetura

```
┌─────────────────┐
│  Browser (SPA)  │  OTel Web SDK: propaga traceparent em todo fetch()
│  Svelte/CF Pages│
└────────┬────────┘
         │ traceparent: 00-{trace_id}-{span_id}-01
         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                        Cloud Run (garimpei-v2)                                    │
│                                                                                  │
│  ┌────────────┐   gRPC+traceparent   ┌───────────┐                              │
│  │garimpei-api│──────────────────────→│ collector │                              │
│  │  (C# .NET) │──────────────────────→│ publisher │                              │
│  │   :8080    │──────────────────────→│ scheduler │                              │
│  │            │──HTTP+traceparent────→│  analyzer │                              │
│  └─────┬──────┘                       └─────┬─────┘                              │
│        │ OTLP                               │ OTLP                               │
│        ▼                                    ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐         │
│  │              OTel Collector Sidecar (Google-Built)                    │         │
│  │              OTLP :4317 → Cloud Trace + Cloud Logging                │         │
│  └──────────────────────────────┬──────────────────────────────────────┘         │
└─────────────────────────────────┼────────────────────────────────────────────────┘
                                  ▼
                      ┌───────────────────────┐
                      │   GCP Console          │
                      │  • Cloud Trace         │
                      │  • Log Explorer        │
                      └───────────────────────┘
```

## Fluxos rastreáveis

| Fluxo | Spans gerados |
|-------|---------------|
| Busca de produtos | browser → garimpei-api → collector.Fetch |
| Salvar busca agendada | browser → garimpei-api → scheduler.SetSchedule |
| Job de coleta (cron) | scheduler.job → collector.FetchShop → Cloud Tasks enqueue |
| Alerta de preço | Cloud Tasks → scheduler.processAlert → analyzer.quedas → publisher.Publish |
| Publicação | browser → garimpei-api → collector.GenerateAffiliateLink → publisher.Publish |

## Como inspecionar traces

### Cloud Trace (GCP Console)

1. Abrir: https://console.cloud.google.com/traces/list?project=garimpo-500114
2. Filtrar por service name: `garimpei-api`, `collector`, `scheduler`, `analyzer`, `publisher`
3. Clicar num trace para ver o waterfall de spans

### Cloud Logging (Log Explorer)

1. Filtrar por trace_id: `logging.googleapis.com/trace="projects/garimpo-500114/traces/{TRACE_ID}"`
2. Todos os logs de todos os services que participaram daquela operação aparecem juntos

### Desenvolvimento local (Jaeger)

```bash
# Inicia Jaeger
docker run -d --name jaeger -p 16686:16686 -p 4317:4317 -p 4318:4318 \
  jaegertracing/all-in-one:latest

# Configura OTel para dev
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# Roda a app
cd web && bun run dev

# Ver traces: http://localhost:16686
```

## Configuração

### Variáveis de ambiente (por serviço)

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Endpoint do OTel Collector | (não set = no-op) |
| `OTEL_SERVICE_NAME` | Nome do serviço nos traces | (parâmetro do Init) |
| `OTEL_TRACES_SAMPLER_ARG` | Sampling rate (0.0 a 1.0) | `1.0` (100%) |
| `GCP_PROJECT_ID` | Project ID para formato de log correlation | `garimpo-500114` |

### Graceful degradation

- Sem `OTEL_EXPORTER_OTLP_ENDPOINT` → SDK usa NoOp (zero overhead)
- Collector indisponível → spans descartados silenciosamente (app continua)
- Testes rodam sem collector (CI não precisa de infra OTel)

## Implementação por linguagem

### Go (collector, publisher, scheduler)

```go
import garimpotel "github.com/fmarquesfilho/garimpo/internal/otel"

func main() {
    ctx := context.Background()
    shutdown, _ := garimpotel.Init(ctx, "collector")
    defer shutdown(ctx)

    logger := slog.New(garimpotel.SlogHandler(""))
    slog.SetDefault(logger)

    srv := grpc.NewServer(garimpotel.GRPCServerInterceptors()...)
}
```

### C# (garimpei-api)

```csharp
builder.Services.AddOpenTelemetry()
    .ConfigureResource(r => r.AddService("garimpei-api"))
    .WithTracing(tracing => tracing
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddGrpcClientInstrumentation()
        .AddOtlpExporter());
```

### Python (analyzer)

```python
from otel_setup import init_otel
init_otel("analyzer")  # antes do FastAPI()
```

### Svelte (frontend — browser + Worker)

**Browser** (`src/lib/telemetry.js`): propaga traceparent em todo fetch.
**Worker** (`src/instrumentation.server.js`): registra TracerProvider com W3C propagation.

## Testes de validação

```bash
mise run test:e2e-traces        # Verifica propagação em produção
mise run test:e2e-services      # Verifica health de todos os serviços
mise run test:e2e-scheduler     # Verifica fluxo agendamento → coleta

# Cross-service E2E (com traceparent para debug)
mise run test:e2e-coleta        # Scheduler → Collector → BigQuery snapshots
mise run test:e2e-alertas       # Analyzer → quedas/novidades → Publisher → Telegram
mise run test:e2e-publicacoes   # Publicação agendada → Publisher gRPC
```

### Cross-service E2E com traceparent

Cada teste cross-service gera `traceparent` W3C em cada step. O summary final
lista todos os trace_ids gerados:

```
  🔭 Traces gerados (para debug via Cloud Trace):
     criar-busca-agendada: 4f3a7b2c1d...
     collector-fetch: 8e2f1a9c3b...
     verificar-snapshots-bq: 1c4d5e6f7a...

  Investigar: mise run debug:trace <trace_id>
```

Isso permite **debugar falhas** investigando o trace no Cloud Trace — cada span
mostra latência, erros, e qual serviço participou da operação.

## Atributos de span padronizados

| Atributo | Onde | Exemplo |
|----------|------|---------|
| `garimpei.tenant_id` | API (root span) | `"abc123"` |
| `garimpei.busca_id` | Scheduler (job) | `"e449fe39-..."` |
| `garimpei.marketplace` | Collector (fetch) | `"shopee"` |
| `garimpei.channel` | Publisher (publish) | `"telegram"` |
| `garimpei.alert_type` | Scheduler (alert) | `"price_drop"` |
