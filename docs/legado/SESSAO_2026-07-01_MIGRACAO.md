# Sessão 2026-07-01 — Início da Migração Arquitetural (ADR-0012)

## Resumo

Sessão de implementação da Fase 0 e início da Fase 1 da migração proposta na
ADR-0012: separar o monólito Go em Web App C# (ASP.NET Core 10) + microserviços
Go (gRPC). O sistema foi deployado em produção com routing split ativo.

Todas as decisões pendentes da ADR foram resolvidas, o mono-repo foi estruturado,
a Fase 0 foi concluída integralmente, e a Fase 1 iniciou com curadoria, multi-tenant
e routing split em produção.

---

## Decisões tomadas

| # | Decisão | Justificativa |
|---|---------|---------------|
| 1 | PostgreSQL para dados da app, BigQuery para analytics | PG é transacional; BQ fica para métricas e conversões |
| 2 | Mono-repo (Go + C# + protos) | Simplifica CI, proto sharing, e versionamento |
| 3 | Cloud Run multi-container | Sidecars gRPC na mesma instância; migrar para serviços separados se necessário |
| 4 | Scheduler como serviço Go separado | Mantém controle fino de timing com goroutines/cron nativo |
| 5 | WhatsApp migra de Maytapi para Meta Cloud API | Remove intermediário, reduz custo, acesso a features oficiais (ADR-0013) |
| 6 | Neon como banco PostgreSQL (free tier) | Validação rápida sem custo; migrar para Cloud SQL em produção definitiva |

---

## Tarefas concluídas nesta sessão

| Task | Título | Fase |
|------|--------|------|
| T-0009 | Setup mono-repo (Go + C# + protos) + Docker Compose | 0 |
| T-0010 | Proto definitions + shopee-collector gRPC server | 0 |
| T-0011 | Publisher gRPC server | 0 |
| T-0012 | Scheduler gRPC server | 0 |
| T-0013 | C# Web App — auth Firebase + health + OpenTelemetry | 0 |
| T-0014 | PostgreSQL schema + EF Core migrations | 0 |
| T-0015 | Multi-tenant (EF Core global query filters) | 1 |
| T-0016 | Curadoria controller + scoring port em C# | 1 |
| T-0017 | Routing split (Cloudflare Worker v1→v2) | 1 |
| T-0021 | Cloud Run multi-container deploy | 0 |
| T-0023 | Migrar WhatsApp de Maytapi para Meta Cloud API | — |
| T-0004 | ScopedStore por owner_uid (resolvida por T-0015) | 1 |

**Tarefas criadas:**
| Task | Título | Status |
|------|--------|--------|
| T-0024 | Testar publicação WhatsApp via Meta Cloud API | next |

---

## Deploy em produção

### Infraestrutura provisionada

| Recurso | Detalhes |
|---------|----------|
| Cloud Run (garimpei-v2) | Multi-container, southamerica-east1, scale 0-2 |
| PostgreSQL | Neon (sa-east-1), database `neondb`, migrations aplicadas |
| Artifact Registry | 5 imagens (garimpei-api-v2, collector, publisher, alerter, scheduler) |
| Cloudflare Worker | Routing split v1/v2 ativo |
| Secret Manager | GARIMPEI_PG_CONNECTION_STRING criado |

### URLs

| Serviço | URL |
|---------|-----|
| C# Web App (direto) | https://garimpei-v2-879269475961.southamerica-east1.run.app |
| C# via Cloudflare | https://garimpei.app.br/api/v2/* |
| Go legado | https://garimpei.app.br/api/* |
| Frontend | https://garimpei.app.br/ |

### Validação em produção

| Teste | Resultado |
|-------|-----------|
| Health check C# | ✅ Healthy |
| Health/ready (PG Neon) | ✅ Healthy |
| OpenAPI spec | ✅ Gerado automaticamente |
| Routing split /api/v2 → C# | ✅ Header x-garimpei-backend: csharp |
| Auth rejeita sem JWT | ✅ 401 |
| Go legado inalterado | ✅ /api/health, /api/candidatos funcionam |
| Frontend SPA | ✅ 200 |

---

## Estrutura do mono-repo (resultado final)

```
garimpo/
├── protos/                          # Contratos gRPC (.proto + buf)
│   ├── collector/v1/
│   ├── publisher/v1/
│   ├── alerter/v1/
│   └── scheduler/v1/
├── gen/go/                          # Stubs Go gerados (commitados)
├── src/                             # Web App C# (.NET 10, Minimal API)
│   ├── Garimpei.Api/                # Endpoints, middleware, Dockerfile
│   ├── Garimpei.Application/        # MediatR handlers (futuro)
│   ├── Garimpei.Domain/             # Entities, interfaces, ScoringService
│   ├── Garimpei.Infrastructure/     # EF Core, TenantContext, gRPC clients
│   ├── Garimpei.Protos/             # Stubs C# pré-gerados (buf)
│   └── Garimpei.Tests/              # xUnit (10 testes)
├── services/                        # Microserviços
│   ├── collector/                   # Go gRPC — Shopee API (:50051)
│   ├── publisher/                   # Go gRPC — Telegram + WhatsApp (:50052)
│   ├── alerter/                     # Go gRPC — Alertas de preço (:50053)
│   ├── scheduler/                   # Go gRPC — Cron + orquestração (:50054)
│   └── analyzer/                    # Python FastAPI — Analytics/BigQuery (:8060)
├── internal/                        # Código Go compartilhado (apenas o necessário)
│   ├── source/                      # → collector
│   ├── publish/                     # → publisher
│   ├── alerts/                      # → alerter
│   ├── store/                       # → alerter (SnapshotRepo)
│   ├── domain/                      # → source (Product model)
│   ├── apperr/                      # → todos (sentinel errors)
│   └── crypto/                      # → store (criptografia)
├── web/                             # Frontend SvelteKit (Cloudflare Pages)
├── cloudflare-worker/               # Routing (/* → Pages, /api/* → Cloud Run)
├── deploy/
│   ├── cloud-run-service.yaml       # Template (placeholders)
│   └── cloud-run-deploy-now.yaml    # Deploy produção (garimpo-500114)
├── docs/                            # Documentação + ADRs + guias
├── docker-compose.yml               # Dev local
├── .github/workflows/ci.yml         # CI unificado
└── Makefile                         # Targets unificados
```

---

## Impacto na qualidade

### Métricas antes vs depois

| Métrica | Antes (monólito Go) | Depois (arquitetura final) |
|---------|--------------------|-----------------------------|
| Linhas de código Go | ~12.000 | ~10.300 (77 arquivos) |
| Linhas removidas | — | **-10.268** (sessão inteira) |
| Arquivos C# | 0 | 61 |
| Arquivos Python | 0 | 8 |
| Stacks | 1 (Go) | 3 (C# + Go + Python) |
| Testes Go | 13 pacotes | 10 pacotes + 4 serviços (12 testes novos) |
| Testes C# | 0 | 10 (xUnit) |
| Multi-tenancy | Não | Sim (EF Core global filters) |
| Persistência | BigQuery only | PostgreSQL + BigQuery |
| Deploy | Monólito único | 6 containers + CDN |
| Frontend hosting | Servido pelo backend | CDN global (Cloudflare Pages) |
| WhatsApp | Maytapi (intermediário) | Meta Cloud API (direto) |
| Código morto | Sim (httpapi, scheduler GCP, etc.) | Zero |

### Cobertura de testes

| Stack | Testes | Cobertura |
|-------|--------|-----------|
| Go (internal) | source 87%, publish 62%, store 36% | Core paths cobertos |
| Go (services) | 12 testes | 11-33% (validações + fluxos) |
| C# (xUnit) | 10 testes | Multi-tenant, persistence, entities |
| Frontend | 109 Vitest + E2E Playwright | Sem alteração |

### Análise estática

| Ferramenta | Resultado |
|-----------|-----------|
| golangci-lint | ✅ 0 issues |
| arch-go | ✅ 100% compliance, 50% coverage (9 regras) |
| buf lint | ✅ protos válidos |
| dotnet build (warnings as errors) | ✅ 0 warnings |
| check-file-size (400 linhas) | ✅ 0 violações |
| ruff (Python) | ✅ 0 issues |

---

## ADRs criadas/atualizadas

| ADR | Título | Status |
|-----|--------|--------|
| 0012 | Migração C# + Go microservices | aceite |
| 0013 | WhatsApp Meta Cloud API | aceite |
| 0014 | Analyzer Python (FastAPI + BigQuery) | aceite |

---

## Problemas encontrados e resolvidos

| Problema | Solução |
|----------|---------|
| buf lint: service names sem sufixo "Service" | Renomear nos .proto |
| Dockerfile C#: `COPY ../protos/` fora do context | Stubs pré-gerados, context = src/ |
| Grpc.Tools não roda em Alpine (musl vs glibc) | Stubs C# pré-gerados via buf |
| Docker buildx cria OCI manifest index | `--provenance=false` no build |
| Imagens arm64 em Mac (Cloud Run requer amd64) | `--platform linux/amd64` |
| Cloud Run exige startup probe em sidecars | TCP probes adicionados |
| Service account sem acesso a Secret Manager | IAM policy binding |
| Dockerfile golang:1.24 vs go.mod 1.26.4 | golang:1.26-alpine |
| Proto não expunha commission (scoring zerava) | Adicionar campo commission ao proto |
| Frontend espera campos em português | DTO de compatibilidade no compat endpoint |
| /api/admin/me não existia (menu admin sumiu) | Endpoint compat com AdminEmails config |
| Cloud Build trigger antigo (Dockerfile raiz) | Trigger deletado |
| Cloud Run usa cache de imagem :latest | Deploy com digest SHA para forçar pull |

---

## Próximos passos

| Task | Título | Status | Nota |
|------|--------|--------|------|
| T-0024 | Testar WhatsApp Meta Cloud API | next | Guia em docs/guias/ |
| T-0005 | Alertas configuráveis por usuário | next | Desbloqueada |
| T-0002 | Persistir conversões no BigQuery | next | Depende do scheduler |
| T-0007 | Recomendação IA personalizada | backlog | Analyzer Python pronto |

### Para testar o sistema

1. **Frontend** → `garimpei.app.br` (login Firebase) — Cloudflare Pages
2. **Busca** → keyword na barra → produtos com comissão e score
3. **Admin** → links privilegiados aparecem para AdminEmails
4. **API v2** → `garimpei.app.br/api/v2/curadoria/ranking?keyword=serum` (precisa JWT)
5. **WhatsApp** → seguir `docs/guias/configurar-whatsapp-meta.md`
6. **Rollback** → `wrangler.toml` V2_ENABLED=false + redeploy
