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
| Web App (novo) | C# / ASP.NET Core 10, Minimal API, EF Core, MediatR |
| Backend legado | Go 1.26, Cloud Run |
| Microserviços | Go, gRPC (collector, publisher, alerter, scheduler) |
| Frontend | SvelteKit 2, Svelte 5, Vite 8 |
| DB transacional | PostgreSQL 17 (Cloud SQL) |
| DB analytics | BigQuery |
| Autenticação | Firebase Auth (JWT, validado em ambos backends) |
| Canais | Telegram Bot API, Meta WhatsApp Business Cloud API |
| CI/CD | GitHub Actions (deploy-gcp.yml + ci-csharp.yml) |
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

## Microserviços gRPC

| Serviço | Porta | Responsabilidade | Proto |
|---------|-------|-----------------|-------|
| collector | 50051 | Fetch de produtos Shopee (keyword/shop) | `collector/v1/collector.proto` |
| publisher | 50052 | Publicação em Telegram/WhatsApp | `publisher/v1/publisher.proto` |
| alerter | 50053 | Verificação de preço + notificação | `alerter/v1/alerter.proto` |
| scheduler | 50054 | Cron jobs + orquestração dos outros serviços | `scheduler/v1/scheduler.proto` |

Todos rodam como sidecars no Cloud Run multi-container. Comunicação via localhost.
Health checks gRPC + graceful shutdown em todos.

## Deploy

### Cloud Run multi-container (novo — ADR-0012)

```yaml
# deploy/cloud-run-service.yaml
containers:
  - garimpei-api (C#, ingress :8080)
  - collector (Go, gRPC :50051)
  - publisher (Go, gRPC :50052)
  - alerter (Go, gRPC :50053)
  - scheduler (Go, gRPC :50054)
```

Container dependencies: C# espera sidecars ficarem healthy antes de receber tráfego.

### CI/CD Pipeline

```
push main
  ├─ deploy-gcp.yml (Go legado)
  │    └─ test-go → build → deploy Cloud Run
  └─ ci-csharp.yml (C# + protos)
       └─ build → test → proto-lint → proto-sync-check → docker build
```

### Monólito Go (legado — coexistência)

O monólito Go continua servindo tráfego nas rotas `/api/*` durante a migração.
Rotas são migradas gradualmente para `/api/v2/*` (C#) com feature flags no
Cloudflare Worker (T-0017).

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
