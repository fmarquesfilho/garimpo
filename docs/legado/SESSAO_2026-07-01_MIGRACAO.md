# Sessão 2026-07-01 — Início da Migração Arquitetural (ADR-0012)

## Resumo

Sessão de implementação da Fase 0 da migração proposta na ADR-0012: separar o
monólito Go em Web App C# (ASP.NET Core 10) + microserviços Go (gRPC).

Todas as decisões pendentes da ADR foram resolvidas, o mono-repo foi estruturado,
e a Fase 0 foi concluída integralmente.

---

## Decisões tomadas

| # | Decisão | Justificativa |
|---|---------|---------------|
| 1 | PostgreSQL para dados da app, BigQuery para analytics | PG é transacional; BQ fica para métricas e conversões |
| 2 | Mono-repo (Go + C# + protos) | Simplifica CI, proto sharing, e versionamento |
| 3 | Cloud Run multi-container | Sidecars gRPC na mesma instância; migrar para serviços separados se necessário |
| 4 | Scheduler como serviço Go separado | Mantém controle fino de timing com goroutines/cron nativo |
| 5 | WhatsApp migra de Maytapi para Meta Cloud API | Remove intermediário, reduz custo, acesso a features oficiais (ADR-0013) |

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
| T-0021 | Cloud Run multi-container deploy | 0 |
| T-0004 | ScopedStore por owner_uid (resolvida por T-0015) | 1 |
| T-0023 | Migrar WhatsApp de Maytapi para Meta Cloud API | — |

---

## Estrutura do mono-repo (resultado)

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
│   ├── Garimpei.Domain/             # Entities, interfaces, IOwnedEntity
│   ├── Garimpei.Infrastructure/     # EF Core, TenantContext, gRPC clients
│   ├── Garimpei.Protos/             # Stubs C# pré-gerados
│   └── Garimpei.Tests/              # xUnit (multi-tenant, persistence)
├── services/                        # Microserviços Go (gRPC)
│   ├── collector/                   # Shopee API (Fetch, FetchShop)
│   ├── publisher/                   # Telegram + WhatsApp (Meta Cloud API)
│   ├── alerter/                     # Verificação de preço + Telegram
│   └── scheduler/                   # Cron nativo + orquestração
├── internal/                        # Código Go existente (monólito legado)
├── deploy/
│   └── cloud-run-service.yaml       # Cloud Run multi-container spec
├── docker-compose.yml               # Dev local (PG, BQ emulator, C#, Go)
└── Makefile                         # Targets unificados
```

---

## Impacto na qualidade

### Antes desta sessão

| Métrica | Valor |
|---------|-------|
| Stacks | Go only |
| Testes Go | 13 pacotes com testes |
| Testes C# | — |
| Cobertura Go (services) | 0% (não existiam) |
| Multi-tenancy | Não implementado |
| Deploy | Monólito único no Cloud Run |
| Persistência | BigQuery only (analytics + transacional misturados) |

### Depois desta sessão

| Métrica | Valor |
|---------|-------|
| Stacks | Go (microserviços) + C# (Web App) |
| Testes Go | 13 pacotes internos + 4 microserviços (12 novos testes) |
| Testes C# | 10 testes xUnit (tenancy + persistence) |
| Cobertura Go (services) | 11-33% (validações, fluxos sem I/O externo) |
| Multi-tenancy | ✅ Global query filters (EF Core) + auto-set OwnerUid |
| Deploy | Multi-container (C# ingress + 4 sidecars Go gRPC) |
| Persistência | PostgreSQL (transacional) + BigQuery (analytics) |

### Ganhos de qualidade

1. **Isolamento de dados** — tenant A não vê dados de tenant B (query filter automático)
2. **Separação de responsabilidades** — coleta/publicação/alertas isolados em microserviços
3. **Observabilidade** — OpenTelemetry + Serilog no C#, slog estruturado no Go
4. **Contratos tipados** — gRPC com proto definitions (quebra de contrato detectada no CI)
5. **CI mais robusto** — proto lint + sync check, arch-go com regras para services
6. **Health checks** — gRPC health em todos os sidecars, PG connectivity no C#

---

## Impacto na manutenção

### Positivo

- **Produtividade C#** — novas features (multi-tenant, CQRS, DDD) são mais naturais
- **Microserviços estáveis** — collector/publisher/alerter raramente mudam; isolados
- **Deploy atômico** — multi-container garante versionamento conjunto
- **Proto como contrato** — alterações na interface são explícitas e versionadas
- **Testes de isolamento** — regressão de multi-tenancy é detectada automaticamente

### Custo adicionado

- **Dois runtimes** — .NET + Go no mesmo repo (mitiga-se com Docker + CI separados)
- **Proto sync** — alteração de `.proto` requer regeneração (CI valida drift)
- **Complexidade de rede** — gRPC entre containers (mitigado: localhost no multi-container)
- **Curva de aprendizado** — EF Core, MediatR, DI para quem vem do Go

### Dívida técnica identificada

| Item | Prioridade | Nota |
|------|-----------|------|
| `internal/httpapi/whatsapp.go` usa `ErrMaytapi` legado | Baixa | Remover quando migrar handler |
| `internal/httpapi/httpapi_test.go` (936 linhas) | Média | Refatorar em testes por handler |
| `internal/store` coverage 36% | Média | Adicionar testes de integração (BigQuery emulator) |
| alerter gRPC coverage 11% | Baixa | Precisa de mock do SnapshotRepo |

---

## Próximos passos (Fase 1)

| Task | Título | Status |
|------|--------|--------|
| T-0016 | Curadoria controller + scoring port em C# | next |
| T-0017 | Routing split (Cloudflare Worker v1→v2) | backlog |
| T-0005 | Alertas configuráveis por usuário | next (desbloqueada) |

A Fase 0 está completa. O sistema pode coexistir (Go monólito serve tráfego atual,
C# Web App pronto para receber rotas novas). A próxima etapa é portar a lógica de
curadoria (T-0016) e fazer o routing split no Cloudflare Worker (T-0017).
