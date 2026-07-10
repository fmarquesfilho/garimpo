# Technical Design — Observabilidade Unificada

## 1. Visão Geral da Arquitetura

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                        Cloud Run (garimpei-v2)                                    │
│                                                                                  │
│  ┌────────────┐   gRPC+traceparent   ┌───────────┐                              │
│  │garimpei-api│──────────────────────→│ collector │                              │
│  │  (C# .NET) │──────────────────────→│ publisher │                              │
│  │   :8080    │──────────────────────→│ scheduler │                              │
│  │            │──HTTP+traceparent────→│  analyzer │                              │
│  └─────┬──────┘                       └─────┬─────┘                              │
│        │                                    │                                    │
│        │ OTLP :4317                         │ OTLP :4317                         │
│        ▼                                    ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐         │
│  │              OTel Collector Sidecar (Google-Built)                    │         │
│  │              - Receives: OTLP gRPC (:4317)                           │         │
│  │              - Exports: Cloud Trace + Cloud Logging                   │         │
│  └─────────────────────────────────────────────────────────────────────┘         │
│                              │                                                   │
└──────────────────────────────┼───────────────────────────────────────────────────┘
                               ▼
                   ┌───────────────────────┐
                   │   GCP Observability    │
                   │  • Cloud Trace         │
                   │  • Cloud Logging       │
                   │  • Log Explorer        │
                   └───────────────────────┘
```

### Propagação de contexto (W3C traceparent)

```
Frontend → [traceparent] → C# API → [gRPC metadata] → Go Services
                                   → [HTTP header]   → Python Analyzer
                                   → [Cloud Tasks header] → Scheduler callback
```

## 2. Componentes e Pacotes

### 2.1 Go — Pacote compartilhado `internal/otel`

**Arquivo:** `internal/otel/otel.go`

Responsabilidades:
- Inicializar `TracerProvider` com OTLP exporter (batch, gRPC)
- Configurar `ParentBased(TraceIDRatioBased(rate))` sampler
- Registrar resource attributes (`service.name`, `service.version`)
- Fornecer gRPC interceptors (client + server) pré-configurados
- Fornecer HTTP transport com propagação automática
- Configurar slog handler wrapper para injetar trace correlation fields
- Shutdown graceful com flush de spans pendentes

**Dependências (adicionar ao go.mod):**
```
go.opentelemetry.io/otel v1.43.0
go.opentelemetry.io/otel/sdk v1.43.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.68.0
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0
go.opentelemetry.io/otel/propagation
```

**Interface pública:**
```go
package otel

// Init inicializa OpenTelemetry para o serviço.
// Retorna shutdown function que deve ser chamada no defer/signal handler.
func Init(ctx context.Context, serviceName string) (shutdown func(context.Context) error, err error)

// GRPCServerInterceptors retorna interceptors para grpc.NewServer.
func GRPCServerInterceptors() []grpc.ServerOption

// GRPCDialOptions retorna options para grpc.NewClient.
func GRPCDialOptions() []grpc.DialOption

// HTTPTransport retorna um http.RoundTripper que propaga trace context.
func HTTPTransport() http.RoundTripper

// SlogHandler retorna um slog.Handler que injeta trace fields do context.
func SlogHandler(projectID string) slog.Handler
```

### 2.2 Go — Integração slog ↔ Trace

**Arquivo:** `internal/otel/sloghandler.go`

Custom `slog.Handler` que wrapa `slog.JSONHandler` e adiciona:
- `logging.googleapis.com/trace`: `projects/{project_id}/traces/{trace_id}`
- `logging.googleapis.com/spanId`: hex span_id
- `severity` (mapeado de slog level para Cloud Logging severity)

Extrai trace context do `context.Context` via `trace.SpanFromContext(ctx)`.

### 2.3 C# — Configuração OTel no Program.cs

**Pacotes já instalados** (v1.16.0):
- `OpenTelemetry.Extensions.Hosting`
- `OpenTelemetry.Exporter.OpenTelemetryProtocol`
- `OpenTelemetry.Instrumentation.AspNetCore`
- `OpenTelemetry.Instrumentation.Http`

**Pacote adicional:**
- `OpenTelemetry.Instrumentation.GrpcNetClient` (para propagar em chamadas gRPC outgoing)

**Integração:**
```csharp
// Program.cs — adicionar ao builder.Services
builder.Services.AddOpenTelemetry()
    .ConfigureResource(r => r.AddService("garimpei-api"))
    .WithTracing(tracing => tracing
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddGrpcClientInstrumentation()
        .AddOtlpExporter());
```

**Log correlation** — usar `OpenTelemetry.Extensions.Hosting` que automaticamente popula `Activity.Current` no escopo do ILogger. Configurar o JSON formatter para emitir `logging.googleapis.com/trace`.

### 2.4 Python — Instrumentação FastAPI

**Pacotes a adicionar em `requirements.txt`:**
```
opentelemetry-api==1.30.0
opentelemetry-sdk==1.30.0
opentelemetry-exporter-otlp-proto-grpc==1.30.0
opentelemetry-instrumentation-fastapi==0.51b0
opentelemetry-instrumentation-logging==0.51b0
```

**Integração em `main.py`:**
```python
from otel_setup import init_otel

init_otel("analyzer")  # antes de criar app

app = FastAPI(...)
```

**Arquivo:** `services/analyzer/otel_setup.py`
- Inicializa `TracerProvider` com OTLP exporter
- Instrumenta FastAPI automaticamente
- Configura logging para incluir trace fields no formato GCP
- Graceful degradation: se `OTEL_EXPORTER_OTLP_ENDPOINT` não está set, usa NoOpTracer

### 2.5 OTel Collector Sidecar — Configuração

**Imagem:** `us-docker.pkg.dev/cloud-ops-agents-artifacts/cloud-run-gmp-sidecar/cloud-run-otel-collector:0.3.0`

**Arquivo de config:** `deploy/otel-collector-config.yaml`
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

processors:
  batch:
    timeout: 5s
    send_batch_size: 256
  resourcedetection:
    detectors: [gcp]

exporters:
  googlecloud:
    project: garimpo-500114

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch, resourcedetection]
      exporters: [googlecloud]
    logs:
      receivers: [otlp]
      processors: [batch, resourcedetection]
      exporters: [googlecloud]
```

### 2.6 Cloud Run Deployment — Container adicional

**Adição ao `deploy/cloud-run-deploy-now.yaml`:**
```yaml
- name: otel-collector
  image: us-docker.pkg.dev/cloud-ops-agents-artifacts/cloud-run-gmp-sidecar/cloud-run-otel-collector:0.3.0
  env:
    - name: OTEL_COLLECTOR_CONFIG
      value: |
        receivers:
          otlp:
            protocols:
              grpc:
                endpoint: 0.0.0.0:4317
        processors:
          batch: {}
          resourcedetection:
            detectors: [gcp]
        exporters:
          googlecloud:
            project: garimpo-500114
        service:
          pipelines:
            traces:
              receivers: [otlp]
              processors: [batch, resourcedetection]
              exporters: [googlecloud]
  resources:
    limits:
      cpu: "0.25"
      memory: 128Mi
  startupProbe:
    tcpSocket:
      port: 4317
    periodSeconds: 2
    failureThreshold: 5
```

**Container dependencies atualizado:**
```yaml
run.googleapis.com/container-dependencies: '{"garimpei-api":["collector","publisher","scheduler","analyzer","otel-collector"]}'
```

## 3. Fluxo de Propagação de Trace

### 3.1 Request síncrono (busca de produtos)

```
1. Frontend envia GET /api/candidatos
2. C# API cria root span "GET /api/candidatos"
   └─ Injeta traceparent no gRPC metadata
3. Collector recebe gRPC Fetch, cria child span "collector.Fetch"
   └─ Span inclui attribute garimpei.marketplace=shopee
4. Collector responde
5. C# API finaliza span, responde ao frontend
6. Todos os spans exportados via OTLP → OTel Collector → Cloud Trace
```

### 3.2 Job agendado (coleta + alerta)

```
1. Scheduler cron tick dispara job "busca-{id}"
   └─ Cria root span "scheduler.job.busca-{id}"
   └─ Attributes: garimpei.job_name, garimpei.busca_id
2. Scheduler chama Collector.FetchShop via gRPC
   └─ Propaga traceparent → child span no Collector
3. Scheduler enfileira Cloud Tasks (POST /process-alert)
   └─ Inclui traceparent header na task HTTP
4. Cloud Tasks chama /process-alert no C# API
   └─ C# proxy extrai traceparent, propaga para Scheduler HTTP :8054
5. Scheduler HTTP handler cria child span "scheduler.processAlert"
   └─ Chama Analyzer /quedas (HTTP com traceparent)
   └─ Chama Publisher.Publish (gRPC com metadata)
6. Trace completo visível no Cloud Trace:
   scheduler.job → collector.FetchShop
                 → cloudtasks.enqueue
                 → scheduler.processAlert → analyzer.quedas
                                          → publisher.Publish
```

### 3.3 Cloud Tasks — travessia assíncrona

O Cloud Tasks introduce um gap temporal (delay). Para manter o trace:
- **Enqueue**: serializa `traceparent` header nos `task.HttpRequest.Headers`
- **Callback**: deserializa o `traceparent` e cria um **link** (não parent) ao span original, ou opcionalmente um child span se quiser visualização linear

Decisão: usar **child span** (continuação do trace) para simplificar a visualização no Cloud Trace. O gap temporal será visível como duração do span pai (que inclui o delay do Cloud Tasks).

## 4. Configuração de Ambiente

### Variáveis de ambiente (adicionadas a cada container no deploy manifest):

| Container | Variável | Valor |
|-----------|----------|-------|
| Todos | `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` |
| garimpei-api | `OTEL_SERVICE_NAME` | `garimpei-api` |
| collector | `OTEL_SERVICE_NAME` | `collector` |
| publisher | `OTEL_SERVICE_NAME` | `publisher` |
| scheduler | `OTEL_SERVICE_NAME` | `scheduler` |
| analyzer | `OTEL_SERVICE_NAME` | `analyzer` |
| Todos | `OTEL_TRACES_SAMPLER_ARG` | `1.0` (ajustável) |
| Go services | `GCP_PROJECT_ID` | `garimpo-500114` (já existe no scheduler) |

### Graceful degradation (testes/local):
- Se `OTEL_EXPORTER_OTLP_ENDPOINT` não está configurado → SDK usa NoOp exporter
- Se OTel Collector não está rodando → spans são descartados silenciosamente (batch exporter com timeout)
- Nenhum teste precisa de OTel Collector para passar

## 5. Atributos de Span Padronizados

| Atributo | Onde | Formato | Exemplo |
|----------|------|---------|---------|
| `garimpei.tenant_id` | C# API (root span) | string (Firebase UID) | `"abc123"` |
| `garimpei.busca_id` | Scheduler (job span) | string (UUID) | `"e449fe39-..."` |
| `garimpei.job_name` | Scheduler (job span) | string | `"busca-e449fe39"` |
| `garimpei.alert_type` | Scheduler (alert span) | string | `"price_drop"` |
| `garimpei.marketplace` | Collector (fetch span) | string | `"shopee"` |
| `garimpei.channel` | Publisher (publish span) | string | `"telegram"` |
| `garimpei.chat_id` | Publisher (publish span) | string | `"-100382..."` |

## 6. Estrutura de Arquivos (mudanças)

```
internal/otel/
├── otel.go           ← Init, shutdown, interceptors, transport
├── sloghandler.go    ← Custom slog.Handler com trace correlation
└── otel_test.go      ← Testes do pacote (NoOp quando sem collector)

services/collector/main.go    ← Adicionar otel.Init() + interceptors
services/publisher/main.go    ← Adicionar otel.Init() + interceptors
services/scheduler/main.go    ← Adicionar otel.Init() + interceptors + HTTP transport
services/scheduler/jobs.go    ← Adicionar traceparent no Cloud Tasks
services/scheduler/alerts.go  ← Extrair traceparent no /process-alert

services/analyzer/
├── otel_setup.py             ← NOVO: inicialização OTel Python
├── requirements.txt          ← Adicionar pacotes OTel
└── main.py                   ← Chamar init_otel() no startup

src/Garimpei.Api/
├── Program.cs                ← Adicionar .AddOpenTelemetry() config
├── Garimpei.Api.csproj       ← Adicionar GrpcNetClient instrumentation
└── Middleware/               ← (opcional) TenantSpanEnricher middleware

deploy/
├── cloud-run-deploy-now.yaml ← Adicionar container otel-collector
└── otel-collector-config.yaml← Config do collector (receivers/exporters)
```

## 9. Frontend (Svelte — Cloudflare)

### Contexto e Restrições (atualizado para 2026)

**Descoberta**: Desde SvelteKit 2.31 (agosto 2025), o framework tem suporte first-class a OpenTelemetry:
- `src/instrumentation.server.ts` — carregado antes do app code, configura OTel SDK
- Emite spans automáticos para: `handle` hook, `load` functions, form actions, remote functions
- `event.tracing.root.setAttribute()` permite anotar spans com atributos de negócio

**Limitação atual**: O projeto usa `adapter-static` (SPA pura). O blog do Svelte diz explicitamente: "sorry, `adapter-static`" — não suporta `instrumentation.server.ts`.

**Solução**: Migrar de `adapter-static` para `adapter-cloudflare`. O `adapter-cloudflare` (2026):
- Roda Cloudflare Workers (com server-side) + Static Assets
- Suporta `instrumentation.server.ts` 
- Suporta SSR onde necessário (load functions server-side)
- Mantém fallback SPA para rotas que não precisam de SSR
- O deploy já é via Cloudflare (wrangler) — sem mudança de infra

### Decisão de Design: Migração de Adapter + OTel Nativo

| Fase | O que fazer | Impacto |
|------|-------------|---------|
| **Fase 1** (imediata) | Browser: OTel fetch instrumentation (propaga traceparent) | Zero infra, funciona com adapter-static |
| **Fase 2** (recomendada) | Migrar adapter-static → adapter-cloudflare | Habilita `instrumentation.server.ts` |
| **Fase 3** (com adapter-cloudflare) | `src/instrumentation.server.ts` com OTel SDK | Traces server-side no SvelteKit (load, handle, SSR) |

### 9.1 Fase 1: Propagação de Trace (Browser → Backend) — Funciona AGORA

Funciona com adapter-static. Não precisa de SSR.

**Pacotes:** `@opentelemetry/sdk-trace-web`, `@opentelemetry/instrumentation-fetch`

**Arquivo:** `web/src/lib/telemetry.js`
```javascript
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { W3CTraceContextPropagator } from '@opentelemetry/core';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { registerInstrumentations } from '@opentelemetry/instrumentation';

const provider = new WebTracerProvider({
  resource: { attributes: { 'service.name': 'garimpei-web' } }
});

// Apenas propaga — não exporta spans do browser
provider.register({ propagator: new W3CTraceContextPropagator() });

registerInstrumentations({
  instrumentations: [
    new FetchInstrumentation({
      propagateTraceHeaderCorsUrls: [/garimpei\.app\.br/, /localhost/],
      clearTimingResources: true
    })
  ]
});
```

**Importado em:** `web/src/routes/+layout.js` (ou `+layout.svelte` no `<script>` top-level)

### 9.2 Fase 2: Migração adapter-static → adapter-cloudflare

**Mudanças:**
```diff
- import adapter from '@sveltejs/adapter-static';
+ import adapter from '@sveltejs/adapter-cloudflare';

const config = {
  kit: {
-   adapter: adapter({ pages: 'build', assets: 'build', fallback: '200.html' })
+   adapter: adapter()
  }
};
```

**Impacto:**
- Deploy continua via Cloudflare (wrangler) — mesmo CI
- Rotas existentes continuam como SPA (universal load functions rodam no browser)
- Habilita SSR opcional para rotas que quiserem (futuro)
- Habilita `instrumentation.server.ts`

### 9.3 Fase 3: Traces Server-Side Nativos do SvelteKit

Com adapter-cloudflare + `experimental.tracing.server`:

**`svelte.config.js`:**
```javascript
import adapter from '@sveltejs/adapter-cloudflare';

const config = {
  kit: {
    adapter: adapter(),
    experimental: {
      tracing: { server: true },
      instrumentation: { server: true }
    }
  }
};
```

**`src/instrumentation.server.js`:**
```javascript
// Cloudflare Workers não tem Node — usa SDK leve
import { trace } from '@opentelemetry/api';

// Em Cloudflare Workers, o OTel SDK precisa de exporter HTTP (não gRPC)
// Exporta para o backend C# API que faz proxy → OTel Collector
// Ou exporta direto para Cloud Trace via HTTP (se credenciais disponíveis)

// Minimal setup: apenas registra provider para propagação
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { W3CTraceContextPropagator } from '@opentelemetry/core';

const provider = new WebTracerProvider({
  resource: { attributes: { 'service.name': 'garimpei-web-ssr' } }
});
provider.register({ propagator: new W3CTraceContextPropagator() });
```

**O que ganha com tracing.server:**
- Spans automáticos para cada `load` function server-side
- Span do `handle` hook (middleware chain)
- `event.tracing.root.setAttribute('garimpei.tenant_id', user.uid)` — enriquece trace com dados de negócio
- Traces propagam automaticamente para fetch calls server-side (para a API C#)

### 9.4 Desenvolvimento Local (dev server)

Para dev local (`npm run dev`), o Vite dev server roda Node.js — suporta `instrumentation.server.ts` nativamente. Isso significa que em dev:

- Traces OTel funcionam completos (Node SDK → Jaeger local ou console)
- Não precisa de Cloudflare Workers para testar observabilidade
- O dev quickstart da docs do SvelteKit funciona out-of-the-box

```bash
# Local: view traces com Jaeger
docker run -d --name jaeger -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one
npm run dev
# → http://localhost:16686 para ver traces
```

### 9.5 CORS — Backend aceita traceparent

O C# API precisa aceitar o header `traceparent` em CORS:
```csharp
policy.AllowAnyHeader(); // Inclui traceparent, tracestate
```

### 9.6 Impacto nas Decisões de Backend

A inclusão do frontend com SvelteKit OTel nativo **reforça** as decisões:

| Aspecto | Antes | Agora (2026, com adapter-cloudflare) |
|---------|-------|--------------------------------------|
| Propagação | Browser injeta traceparent manualmente | SvelteKit propaga automaticamente no server fetch |
| Server traces | Inexistente | load/handle spans nativos do SvelteKit |
| Dev local | Sem observabilidade | Jaeger local + traces completos |
| Deploy | adapter-static (Cloudflare Pages) | adapter-cloudflare (Workers + Static Assets) |
| OTel Collector | Só no Cloud Run | Cloud Run (backend) + opcionalmente Cloudflare Worker (frontend SSR) |

**Conclusão**: a migração para `adapter-cloudflare` é a mudança arquitetural mais impactante — habilita observabilidade nativa sem qualquer hack. Todas as decisões de backend se sustentam sem mudança.

## 10. Diagrama Completo (com Frontend)

```
┌─────────────────┐
│  Browser (SPA)  │  OTel Web SDK (fetch instrumentation)
│  Svelte/CF Pages│  Gera traceparent → injeta em todos os fetch()
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
│  └──────────────────────────────┬──────────────────────────────────────┘         │
└─────────────────────────────────┼────────────────────────────────────────────────┘
                                  ▼
                      ┌───────────────────────┐
                      │   GCP Observability    │
                      │  Cloud Trace: trace    │
                      │  completo browser→API  │
                      │  →services (1 view)    │
                      └───────────────────────┘
```

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| OTel Collector crash | Perda de telemetria | Graceful degradation — app continua; probe reinicia collector |
| Overhead de latência | Aumento response time | Batch export assíncrono; benchmark antes/depois |
| Volume de traces alto | Custo GCP Cloud Trace | Sampling configurável (reduzir para 0.1 em produção se necessário) |
| Breaking changes no deploy | Downtime | Sidecar é aditivo; rollback = remover container |
| Testes falham sem collector | CI quebra | NoOp exporter automático quando env var não está set |

## 8. Decisões de Design (ADRs implícitos)

1. **OTLP gRPC (não HTTP)**: gRPC é mais eficiente para batch export e é o default do Google-Built Collector.
2. **Sidecar (não agent externo)**: Cloud Run multi-container já suporta; evita rede extra.
3. **Child span (não link) para Cloud Tasks**: simplifica visualização no Cloud Trace. O gap temporal fica no span duration.
4. **slog Handler custom (não replace global)**: permite teste com handler padrão e produção com trace handler.
5. **GCP-native exporters (não OTLP→GCP bridge)**: Google-Built Collector faz a ponte; SDKs exportam OTLP genérico.
