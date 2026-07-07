# Garimpei

Plataforma de curadoria e publicação automatizada para afiliados Shopee.
Busca produtos, monitora lojas, rankeia por potencial de retorno, e publica em
canais (Telegram, WhatsApp) — tudo multi-tenant com rastreamento de conversão.

**URL:** https://garimpei.app.br
**Docs:** https://garimpei.app.br/docs

## Arquitetura

Arquitetura poliglota orientada a serviços (ADR-0012):

| Stack | Responsabilidade |
|-------|-----------------|
| **C# (ASP.NET Core 10)** | API principal: CRUD, auth, multi-tenant, orquestração |
| **Go (gRPC)** | Microserviços I/O: coleta Shopee, publicação, scheduling |
| **Python (FastAPI)** | Analytics: queries BigQuery, variações preço, séries temporais |
| **SvelteKit** | Frontend SPA via CDN (Cloudflare Pages) |

Deploy: **Cloud Run multi-container** (5 containers, scale-to-zero).

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
        │ scheduler (Go)     ← gRPC :50054         │
        │ analyzer (Python)  ← REST :8060          │
        └───────────┬────────────────┬─────────────┘
                    │                │
            PostgreSQL (Neon)   BigQuery (GCP)
```

## Desenvolvimento local

Requer: [mise](https://mise.jdx.dev/) (task runner + version manager).

```bash
# Instalar ferramentas (Go, Node, Python, dotnet-ef, golangci-lint)
mise install

# Subir todos os serviços (PostgreSQL, API C#, Analyzer, BQ emulator)
mise run up

# Aplicar migrations no banco local
mise run test:csharp   # aplica automaticamente antes dos testes

# Frontend (porta 5173)
cd web && npm install && npm run dev
```

## Testes

```bash
# Tudo de uma vez
mise run test

# Por stack
mise run test:go         # Go (microserviços + internal)
mise run test:csharp     # C# (68 testes: persistence + arch + integração)
mise run test:web        # Frontend (141 testes Vitest)

# E2E com integração real (requer Collector + Firebase Emulator)
mise run test:e2e:lojas           # ResolveShop + adicionar lojas
mise run test:e2e:buscas-agendadas  # Scheduler + keywords + coleta

# Checks de integridade cross-stack
mise run checks          # contratos, data ownership, schema sync, drift
```

## Estrutura do repositório

```
src/                    C# Web App (ASP.NET Core 10, Clean Architecture)
services/
  collector/            Shopee API collector (Go, gRPC)
  publisher/            Telegram/WhatsApp publisher (Go, gRPC)
  scheduler/            Orquestrador cron + Cloud Tasks (Go, gRPC)
  analyzer/             Analytics BigQuery (Python, FastAPI)
web/                    Frontend SvelteKit 5 + shadcn-svelte
cloudflare-worker/      Routing inteligente (domínio garimpei.app.br)
protos/                 Contratos gRPC (Protocol Buffers)
contracts/              Contratos de serviço (registry.yaml + JSON Schemas)
deploy/                 Cloud Run YAMLs + BigQuery schema
docs/                   Documentação (arquitetura, fluxos, ADRs)
backlog/                Product backlog (tasks-as-code)
.mise/tasks/            Tasks de CI/check (mise run)
```

## Documentação

- `docs/01-visao-e-negocio.md` — Visão do produto
- `docs/02-arquitetura.md` — Arquitetura detalhada + data ownership
- `docs/03-fluxos-e-modelo.md` — Entidades, buscas agendadas, pipelines
- `docs/04-operacao-shopee.md` — Integração Shopee (GraphQL + API v4)
- `docs/06-qualidade-e-testes.md` — CI, testes, fitness functions
- `docs/08-fluxos-sequencia.md` — Diagramas de sequência (11 fluxos)
- `docs/decisoes/` — ADRs (24 Architecture Decision Records)
- `contracts/` — Contratos de serviço (ADR-0020)
- `api/openapi.yaml` — OpenAPI 3.1 spec

## CI

Push para `main` executa automaticamente:

- Go: build + test + lint + arch-go
- C#: build + test (com PostgreSQL service)
- Python: ruff lint + syntax check
- Proto: buf lint + sync check + breaking changes
- Frontend: build + lint + vitest + Playwright
- Contratos: service-contracts + api-contract + config-consistency + schema-sync + data-ownership
- Docker: build de todas as 5 imagens
- Deploy: migrations (Neon) → Cloud Run + Cloudflare Pages

## Mise tasks úteis

```bash
mise tasks              # lista todas as tasks
mise run up             # sobe serviços locais
mise run ci             # simula CI completo
mise run prepush        # verificação pré-push (~1min)
mise run deploy:migrate # aplica migrations em produção
mise run checks         # todos os drift checks
```
