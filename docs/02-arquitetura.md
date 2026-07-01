# Arquitetura

## Visão geral

O Garimpei está em transição de monólito Go para uma arquitetura híbrida:
**Web App C# (ASP.NET Core 10)** como aplicação principal + **microserviços Go (gRPC)**
para tarefas de I/O intensivo. Ambos coexistem durante a migração (ADR-0012).

```
                    ┌─────────────────────────────────────────────────┐
                    │       Cloud Run (multi-container)                │
                    │                                                 │
navegador ─https──► │  garimpei-api (C# .NET 10) ← ingress :8080     │
                    │    ├─ /api/v2/** (novas rotas)                  │
                    │    ├─ /health, /health/ready                    │
                    │    ├─ PostgreSQL (EF Core, multi-tenant)        │
                    │    └─ gRPC clients → sidecars                   │
                    │                                                 │
                    │  garimpei-api-legacy (Go) ← legado :8081        │
                    │    ├─ /api/** (rotas existentes)                │
                    │    └─ BigQuery (analytics)                      │
                    │                                                 │
                    │  collector (Go gRPC :50051)                     │
                    │    └─ Shopee Affiliate API                      │
                    │  publisher (Go gRPC :50052)                     │
                    │    └─ Telegram Bot API + Meta WhatsApp Cloud    │
                    │  alerter (Go gRPC :50053)                       │
                    │    └─ Verificação preço + Telegram              │
                    │  scheduler (Go gRPC :50054)                     │
                    │    └─ Cron nativo + orquestra via gRPC          │
                    └─────────────────────────────────────────────────┘
                              │                    │
                    ┌─────────▼────────┐  ┌───────▼──────────┐
                    │  Cloud SQL (PG)  │  │  BigQuery         │
                    │  dados app       │  │  analytics/export │
                    └──────────────────┘  └──────────────────┘

Cloudflare Worker ──routing──► /api/v2/* → C#, /api/* → Go (legado)
```

## Stack

| Camada | Tecnologia |
|---|---|
| Web App | C# / ASP.NET Core 10, Minimal API, EF Core, MediatR |
| Microserviços I/O | Go, gRPC (collector, publisher, alerter, scheduler) |
| Analytics | Python, FastAPI, pandas, BigQuery |
| Frontend | SvelteKit 2, Svelte 5, Vite 8 |
| DB transacional | PostgreSQL 17 (Neon) |
| DB analytics | BigQuery |
| Autenticação | Firebase Auth (JWT, validado no C#) |
| Canais | Telegram Bot API, Meta WhatsApp Business Cloud API |
| CI | GitHub Actions (ci.yml — Go + C# + Python + Proto + Frontend + Docker) |
| Hosting frontend | Cloudflare Pages |
| Proxy/Routing | Cloudflare Workers |
| Infra | Cloud Run multi-container, Artifact Registry, Secret Manager |
| Observabilidade | OpenTelemetry + Serilog (C#), slog JSON (Go) |
| Contratos | Protocol Buffers (buf) — Go + C# stubs pré-gerados |

## Persistência (dual)

| Store | Uso | Acesso |
|-------|-----|--------|
| PostgreSQL | Produtos, buscas, tenants, configs, publicações | Web App C# (EF Core) |
| BigQuery | Conversões, snapshots, métricas históricas, export | Microserviços Go + Go legado |

A separação segue a ADR-0012: PostgreSQL para dados transacionais (CRUD),
BigQuery para analytics e séries temporais.

## Multi-tenancy

Isolamento por `owner_uid` (Firebase user_id):

1. **JWT** → TenantMiddleware extrai `user_id` do claim
2. **EF Core global query filter** → `WHERE owner_uid = @tenant` automático
3. **SaveChanges** → novas entidades recebem `owner_uid` do contexto
4. **Rejeição** → 401 se claim ausente em rotas autenticadas

Ver ADR-0012, T-0015.

## Microserviços gRPC + REST

| Serviço | Porta | Stack | Responsabilidade | Proto/API |
|---------|-------|-------|-----------------|-----------|
| collector | 50051 | Go gRPC | Fetch de produtos Shopee (keyword/shop) | `collector/v1/collector.proto` |
| publisher | 50052 | Go gRPC | Publicação em Telegram/WhatsApp | `publisher/v1/publisher.proto` |
| alerter | 50053 | Go gRPC | Verificação de preço + notificação | `alerter/v1/alerter.proto` |
| scheduler | 50054 | Go gRPC | Cron jobs + orquestração dos outros serviços | `scheduler/v1/scheduler.proto` |
| analyzer | 8060 | Python REST | Analytics, novidades, quedas, evolução | FastAPI (OpenAPI auto-gerado) |

Todos rodam como sidecars no Cloud Run multi-container. Comunicação via localhost.
Health checks gRPC + graceful shutdown em todos.

## Deploy

### Cloud Run multi-container (produção)

```yaml
# deploy/cloud-run-deploy-now.yaml
containers:
  - garimpei-api (C#, ingress :8080)
  - collector (Go, gRPC :50051)
  - publisher (Go, gRPC :50052)
  - alerter (Go, gRPC :50053)
  - scheduler (Go, gRPC :50054)
  - analyzer (Python, HTTP :8060)
```

Container dependencies: C# espera sidecars ficarem healthy antes de receber tráfego.

Deploy manual:
```bash
# Build e push (--platform linux/amd64 --provenance=false)
docker build ... -f src/Garimpei.Api/Dockerfile src/
docker build ... -f services/collector/Dockerfile .
docker build ... -f services/analyzer/Dockerfile services/analyzer/
# (publisher, alerter, scheduler análogos)

# Deploy
gcloud run services replace deploy/cloud-run-deploy-now.yaml --region=southamerica-east1
```

### Frontend (Cloudflare Pages)

```bash
cd web && npm run build
npx wrangler pages deploy build --project-name garimpei-web
```

### CI Pipeline

```
push main → ci.yml
  ├─ go (build + test + lint + arch-go + docs-check)
  ├─ csharp (build + test)
  ├─ python (ruff + syntax)
  ├─ proto (lint + sync check)
  ├─ frontend (build + lint + vitest)
  └─ docker (build all 6 images)
```

### Routing (Cloudflare Worker)

```
garimpei.app.br/api/* → Cloud Run (C# garimpei-v2)
garimpei.app.br/*     → Cloudflare Pages (frontend SPA)
```

Rollback: `V2_ENABLED=false` no Worker → `/api/*` volta para Go legado (se ainda existir).

## Coleta e scheduler

O scheduler (microserviço Go) substitui o Cloud Scheduler:
- Cron nativo com `robfig/cron` (timezone BRT)
- Chama collector via gRPC para buscar produtos
- Gerenciável via gRPC (SetSchedule, TriggerJob, ListJobs)

### Amostragem rotativa (lojas)

Para lojas monitoradas, a coleta usa paginação rotativa:
- `rotation_cursor` armazena a próxima página por loja
- Cada ciclo avança 2 páginas (100 produtos)
- `full_scan_at` registra quando completou varredura do catálogo inteiro

### Throttling

- 200ms entre páginas da mesma loja
- 60s entre lojas diferentes
- HTTP 429 → espera 30s, retenta até 3×

## Análise estática

- **golangci-lint** — estilo e bugs no Go
- **arch-go** — restrições arquiteturais (12 regras, 100% compliance)
  - `services/*` não importam `internal/httpapi`
  - `internal/domain` não importa infra
- **buf lint** — validação dos .proto (STANDARD rules)
- **proto sync check** — CI verifica que stubs commitados estão atualizados
- **dotnet build** — warnings as errors no C#
- **eslint + stylelint** — frontend
- **check-file-size** — máx 400 linhas (exclui `gen/`)

## Logs e observabilidade

| Stack | Ferramenta |
|-------|-----------|
| C# | Serilog (structured) + OpenTelemetry (traces + metrics via OTLP) |
| Go (services) | `slog` JSON handler |
| Go (legado) | `slog` com campos por request |
| Cloud Run | → Cloud Logging (JSON) |

## ADRs relevantes

- [ADR-0003](/docs/decisoes/0003-deploy-gcp/) — Deploy no GCP
- [ADR-0012](/docs/decisoes/0012-migracao-csharp-go-microservices/) — Migração C# + Go
- [ADR-0013](/docs/decisoes/0013-whatsapp-meta-cloud-api/) — WhatsApp Meta Cloud API
- [ADR-0014](/docs/decisoes/0014-analyzer-python-fastapi/) — Analyzer Python (FastAPI + BigQuery)
