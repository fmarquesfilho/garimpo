# Arquitetura

## Visão geral

O Garimpei usa uma arquitetura poliglota orientada a serviços:
- **C# (ASP.NET Core 10)** — API principal (CRUD, auth, orquestração)
- **Go (gRPC)** — microserviços de I/O intensivo (coleta, publicação, alertas, scheduling)
- **Python (FastAPI)** — analytics e IA (queries BigQuery, detecção de padrões)
- **SvelteKit** — frontend SPA (Cloudflare Pages)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Cloudflare (Edge)                                  │
│                                                                         │
│  ┌─────────────────┐        ┌──────────────────────────────────┐       │
│  │  Pages (CDN)    │        │  Worker (routing)                │       │
│  │  Frontend SPA   │        │  /api/* → Cloud Run              │       │
│  │  (SvelteKit)    │        │  /*     → Pages                  │       │
│  └─────────────────┘        └──────────────┬───────────────────┘       │
└─────────────────────────────────────────────┼───────────────────────────┘
                                              │ HTTPS
┌─────────────────────────────────────────────▼───────────────────────────┐
│                  Cloud Run (multi-container, southamerica-east1)          │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  garimpei-api (C# .NET 10) — ingress container :8080            │   │
│  │                                                                 │   │
│  │  ├── Auth (Firebase JWT)                                        │   │
│  │  ├── Multi-tenant (EF Core global query filters)                │   │
│  │  ├── Curadoria/Scoring (4 fontes: busca, quedas, novos, fav)    │   │
│  │  ├── Publicação (orquestra → publisher gRPC)                    │   │
│  │  ├── Buscas/Lojas CRUD (PostgreSQL)                             │   │
│  │  ├── OpenTelemetry + Serilog                                    │   │
│  │  └── Health checks (/health, /health/ready)                     │   │
│  └────────────┬──────────────┬──────────────┬──────────┬──────────┘   │
│               │gRPC          │gRPC          │gRPC      │HTTP          │
│  ┌────────────▼──┐  ┌───────▼───┐  ┌───────▼──┐  ┌───▼──────────┐   │
│  │  collector    │  │ publisher │  │  alerter │  │   analyzer   │   │
│  │  (Go :50051)  │  │ (Go:50052)│  │ (Go:50053)│  │ (Py :8060)  │   │
│  │               │  │           │  │          │  │              │   │
│  │  Shopee API   │  │ Telegram  │  │ Telegram │  │  BigQuery    │   │
│  │  HMAC-SHA256  │  │ WhatsApp  │  │ preço    │  │  novidades   │   │
│  │  throttling   │  │ Meta API  │  │ alertas  │  │  quedas      │   │
│  └───────────────┘  └───────────┘  └──────────┘  │  evolução    │   │
│                                                   │  estatísticas│   │
│  ┌───────────────┐                               └──────┬───────┘   │
│  │  scheduler    │                                      │            │
│  │  (Go :50054)  │──── orquestra collector/alerter      │            │
│  │  cron nativo  │                                      │            │
│  └───────────────┘                                      │            │
└─────────────────────────────────────────────────────────┼────────────┘
                          │                               │
              ┌───────────▼──────────┐      ┌─────────────▼──────────┐
              │  PostgreSQL (Neon)   │      │  BigQuery              │
              │  dados transacionais │      │  analytics / snapshots │
              │  produtos, buscas,   │      │  conversões, métricas  │
              │  tenants, configs    │      │  séries temporais      │
              └──────────────────────┘      └────────────────────────┘
```

## Vantagens da arquitetura atual vs monólito Go

| Aspecto | Monólito Go (antes) | Arquitetura atual |
|---------|--------------------|--------------------|
| **Produtividade** | Go verboso para CRUD/DI/patterns | C# com DI nativo, EF Core, MediatR, records |
| **Multi-tenancy** | Manual (query por query) | Automático (EF Core global query filters) |
| **Persistência** | BigQuery para tudo (analytics + CRUD) | PostgreSQL (transacional) + BigQuery (analytics) |
| **Isolamento** | Tudo no mesmo processo | Microserviços independentes (deploy/scale separado) |
| **Resiliência** | Falha na coleta derruba toda a API | Sidecar pode falhar sem afetar CRUD |
| **Observabilidade** | slog manual | OpenTelemetry (traces + metrics) + Serilog estruturado |
| **Auth** | Wrapper manual sobre Firebase | ASP.NET Core JWT middleware + claims nativo |
| **Canais** | Maytapi (intermediário pago) | Meta Cloud API oficial (direto, sem markup) |
| **Frontend** | Servido pelo backend (acoplado) | CDN global (Cloudflare Pages, <50ms TTFB) |
| **Analytics** | Go consultando BigQuery (sem ecossistema) | Python + pandas (preparado para ML/IA) |
| **Contratos** | JSON informal (pode quebrar silenciosamente) | Protocol Buffers tipados (breaking change no CI) |
| **Deploy** | Monólito único (tudo ou nada) | 6 containers independentes (rolling update) |
| **Scaling** | Um processo para tudo | Escala por responsabilidade (API vs coleta vs analytics) |
| **Testes** | Integração pesada (httptest + BigQuery mock) | Unitários leves (InMemory DB, gRPC direto) |
| **Código** | ~12000 linhas Go (tudo junto) | Separado por domínio e linguagem |
| **Evolução IA** | Limitada (sem ecossistema ML em Go) | Python pronto para scikit-learn, pandas, LLMs |
| **Custo** | Cloud Run único (~sempre ligado) | Multi-container scale-to-zero + CDN grátis |

### Ganhos mensuráveis

1. **-8247 linhas** de código morto removidas
2. **~50ms TTFB** no frontend (CDN global vs Cloud Run cold start)
3. **Zero custo** de intermediário WhatsApp (Maytapi → Meta direto)
4. **Isolamento de falhas** — coleta Shopee com throttling não bloqueia API
5. **Multi-tenant desde o dia 1** — novo afiliado não vê dados de outro
6. **3 linguagens** cada uma no que faz melhor (C# CRUD, Go I/O, Python analytics)
7. **CI unificado** validando Go + C# + Python + Proto + Frontend em cada push

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
