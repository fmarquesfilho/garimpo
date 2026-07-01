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
├── protos/                          # Contratos gRPC (.proto + buf config)
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
├── services/                        # Microserviços Go (gRPC)
│   ├── collector/                   # Shopee API (:50051)
│   ├── publisher/                   # Telegram + WhatsApp Meta (:50052)
│   ├── alerter/                     # Preço + Telegram (:50053)
│   └── scheduler/                   # Cron + orquestração (:50054)
├── internal/                        # Código Go existente (monólito legado)
├── deploy/
│   ├── cloud-run-service.yaml       # Template multi-container (com placeholders)
│   └── cloud-run-deploy-now.yaml    # Deploy real (garimpo-500114)
├── cloudflare-worker/               # Routing split v1/v2
├── docs/
│   ├── guias/configurar-whatsapp-meta.md
│   └── ...
├── docker-compose.yml               # Dev local
└── Makefile                         # Targets unificados
```

---

## Impacto na qualidade

### Cobertura de testes

| Stack | Testes | Notas |
|-------|--------|-------|
| Go (internal) | 13 pacotes, 85-90% nos pacotes de domínio | Sem alteração |
| Go (services) | 12 testes novos, 11-33% cobertura | Validações + fluxos |
| C# (xUnit) | 10 testes | Multi-tenant, persistence, entities |
| Frontend | 109 testes Vitest + E2E Playwright | Sem alteração |

### Análise estática

| Ferramenta | Resultado |
|-----------|-----------|
| golangci-lint | ✅ 0 issues |
| arch-go | ✅ 100% compliance, 40% coverage (12 regras) |
| buf lint | ✅ protos válidos |
| dotnet build (warnings as errors) | ✅ 0 warnings |
| check-file-size (400 linhas) | ✅ passa (gen/ excluído) |

---

## ADRs criadas/atualizadas

| ADR | Título | Status |
|-----|--------|--------|
| 0012 | Migração C# + Go microservices | aceite |
| 0013 | WhatsApp Meta Cloud API | aceite |

---

## Problemas encontrados e resolvidos

| Problema | Solução |
|----------|---------|
| buf lint: service names sem sufixo "Service" | Renomear nos .proto |
| Dockerfile C#: `COPY ../protos/` fora do context | Context = repo root, depois context = src/ com stubs pré-gerados |
| Grpc.Tools não roda em Alpine (musl vs glibc) | Stubs C# pré-gerados e commitados |
| Docker buildx cria OCI manifest index (Cloud Run rejeita) | `--provenance=false` no build |
| Imagens arm64 em Mac (Cloud Run requer amd64) | `--platform linux/amd64` |
| Cloud Run exige startup probe em sidecars com deps | TCP probes adicionados |
| Service account sem acesso a Secret Manager | IAM policy binding adicionado |
| go.mod 1.26.4 vs Dockerfile golang:1.24 | Atualizar para golang:1.26-alpine |
| arch-go coverage caiu com novos pacotes | Excluir gen/**, adicionar regras para services/ |
| check-file-size falha em .pb.go gerados | Excluir ./gen/* do script |

---

## Próximos passos

| Task | Título | Status | Nota |
|------|--------|--------|------|
| T-0024 | Testar WhatsApp Meta Cloud API | next | Guia em docs/guias/ |
| T-0005 | Alertas configuráveis por usuário | next | Desbloqueada |
| T-0018 | Migrar publicação para C# | backlog | |
| T-0019 | Migrar lojas/buscas para C# | backlog | |
| T-0020 | PG fonte primária + BQ analytics-only | backlog | |
| T-0022 | Descomissionar monólito Go | backlog | |

### Para testar o sistema

1. **Frontend** → login no `garimpei.app.br` (Firebase Auth)
2. **Curadoria C#** → com JWT, acessar `/api/v2/curadoria/ranking?keyword=serum`
3. **WhatsApp** → seguir `docs/guias/configurar-whatsapp-meta.md`
4. **Rollback** → `wrangler.toml` V2_ENABLED=false + redeploy
