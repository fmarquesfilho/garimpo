# Garimpei

Plataforma de curadoria e publicação automatizada para afiliados Shopee.
Busca produtos, monitora lojas, rankeia por potencial de retorno, e publica em
canais (Telegram, WhatsApp) — tudo multi-tenant com rastreamento de conversão.

**URL:** https://garimpei.app.br

## Arquitetura

Arquitetura poliglota orientada a serviços (ADR-0012):

| Stack | Responsabilidade |
|-------|-----------------|
| **C# (ASP.NET Core 10)** | API principal: CRUD, auth, multi-tenant, orquestração |
| **Go (gRPC)** | Microserviços I/O: coleta Shopee, publicação, alertas, scheduling |
| **Python (FastAPI)** | Analytics: queries BigQuery, variações preço, séries temporais |
| **SvelteKit** | Frontend SPA via CDN (Cloudflare Pages) |

Deploy: **Cloud Run multi-container** (6 containers, scale-to-zero).

```
                Cloudflare (Edge)
                      │
        ┌─────────────┼─────────────┐
        │ Pages (CDN)  │ Worker       │
        │ Frontend     │ /api → C#    │
        └──────────────┴──────┬──────┘
                              │
        Cloud Run multi-container
        ┌─────────────────────┴────────────────────┐
        │ garimpei-api (C#)  ← ingress :8080       │
        │ collector (Go)     ← gRPC :50051         │
        │ publisher (Go)     ← gRPC :50052         │
        │ alerter (Go)       ← gRPC :50053         │
        │ scheduler (Go)     ← gRPC :50054         │
        │ analyzer (Python)  ← REST :8060          │
        └───────────┬────────────────┬─────────────┘
                    │                │
            PostgreSQL (Neon)   BigQuery (GCP)
```

## Desenvolvimento local

```bash
# 1. Subir dependências
docker compose up -d postgres

# 2. Aplicar migrations
cd src && dotnet ef database update --project Garimpei.Infrastructure --startup-project Garimpei.Api

# 3. Rodar API (porta 5000 em dev)
cd src && dotnet run --project Garimpei.Api

# 4. Rodar frontend (porta 5173)
cd web && npm install && npm run dev
# Apontar frontend para API C#:
# VITE_API_BASE=http://localhost:5000 npm run dev
```

## Testes

```bash
# C# (23 testes: persistence + arquitetura)
cd src && dotnet test

# Go (microserviços + internal)
go test ./...

# Frontend (unit + E2E)
cd web && npx vitest run        # unit
cd web && npm test              # E2E (Playwright)

# Drift checks (cross-stack consistency)
./scripts/check-api-contract.sh
./scripts/check-config-consistency.sh
./scripts/check-schema-sync.sh
```

## Estrutura do repositório

```
src/                    C# Web App (ASP.NET Core 10, Clean Architecture)
services/
  collector/            Shopee API collector (Go, gRPC)
  publisher/            Telegram/WhatsApp publisher (Go, gRPC)
  alerter/              Alertas de preço (Go, gRPC)
  scheduler/            Orquestrador cron (Go, gRPC)
  analyzer/             Analytics BigQuery (Python, FastAPI)
web/                    Frontend SvelteKit 5
cloudflare-worker/      Routing inteligente (domínio garimpei.app.br)
protos/                 Contratos gRPC (Protocol Buffers)
deploy/                 Cloud Run YAMLs + BigQuery schema
docs/                   Documentação (arquitetura, fluxos, ADRs)
backlog/                Product backlog (tasks-as-code)
scripts/                CI scripts (drift checks, docs, validação)
```

## Documentação

- `docs/01-visao-e-negocio.md` — Visão do produto
- `docs/02-arquitetura.md` — Arquitetura detalhada
- `docs/03-fluxos-e-modelo.md` — Entidades e regras de negócio
- `docs/06-qualidade-e-testes.md` — CI, testes, fitness functions
- `docs/decisoes/` — ADRs (Architecture Decision Records)
- `api/openapi.yaml` — OpenAPI 3.1 spec

## CI

Push para `main` executa:
- Go: build + test + lint + arch-go
- C#: build + test (com PostgreSQL)
- Python: ruff lint + syntax check
- Proto: buf lint + sync check
- Frontend: build + lint + vitest
- API contract: drift frontend↔backend + config consistency + schema sync
- Docker: build de todas as 6 imagens
