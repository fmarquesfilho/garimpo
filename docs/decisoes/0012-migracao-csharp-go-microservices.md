# ADR 0012 — Migração para C# (Web App) + Go (Microserviços gRPC)

**Status:** proposta (aguarda decisão)  
**Data:** 2026-06-29  

## Contexto

O Garimpei hoje é um monólito Go que serve tudo: API REST, frontend estático, coleta
periódica, publicação em canais, e persistência no BigQuery. O sistema funciona bem
para o estágio atual (1 tenant, ~40 rotas, ~5k linhas Go), mas:

1. O autor tem maior conforto e produtividade em C#/.NET (OOP, design patterns, tooling)
2. A evolução futura (multi-tenant SaaS, IA, analytics) pede uma stack que favoreça
   padrões de design avançados (DDD, CQRS, Mediator, etc.)
3. Algumas partes (coleta Shopee, publicação Telegram/WhatsApp) são I/O-bound com
   throttling fino — Go brilha ali e não precisa ser reescrito
4. O frontend (SvelteKit SPA) é independente da stack backend (consome REST/JSON)

## Decisão proposta

Separar o sistema em:

- **Web App principal (C# / ASP.NET Core)** — o monólito "inteligente" que evolui constantemente
- **Microserviços satélite (Go, comunicação gRPC)** — componentes de I/O intensivo que raramente mudam

---

## Mapeamento: o que fica onde

### Web App C# (ASP.NET Core)

| Responsabilidade | Pacote Go atual | Justificação |
|---|---|---|
| API REST (todas as rotas /api/*) | `httpapi` | CRUD-heavy, muita lógica de request/response, ideal para Controllers/Minimal API |
| Autenticação e autorização | `auth` | ASP.NET Identity/JWT middleware é superior ao wrapper manual |
| Multi-tenancy e onboarding | `tenant` | Scoped services, EF Core, DI nativo |
| Persistência (Repository pattern) | `store` | Entity Framework / Dapper + PostgreSQL ou SQL Server |
| Domain model (Product, Busca, etc.) | `domain`, `store` (tipos) | C# records, value objects, rich domain model |
| Lógica de negócio (scoring, ranking) | `scoring`, `strategy`, `engine` | Port para C# — ~400 linhas, puro cálculo |
| Templates e formatação | `publish` (parcial) | Template engine .NET (Scriban/Razor) |
| Admin, logs, métricas | `httpapi`, `logs` | ASP.NET Health Checks, Serilog, OpenTelemetry |
| Scheduler / Jobs | `scheduler` | Hangfire ou Quartz.NET (substituem Cloud Scheduler) |
| Documentação (site Starlight) | `docs-site` | Mantém-se separado (build estático) |

### Microserviços Go (gRPC)

| Serviço | Pacote Go actual | Justificação |
|---|---|---|
| **shopee-collector** | `source`, `coleta` | Throttling (200ms/60s delays), HMAC-SHA256 auth, goroutines para paginação paralela. Raramente muda. |
| **publisher** | `publish` (Telegram, WhatsApp) | HTTP clients com retry, multi-destino, rate limiting. APIs externas não mudam frequentemente. |
| **alerter** | `alerts` | Comparação de snapshots + envio Telegram. Simples, autónomo, fire-and-forget. |

### Contratos gRPC (proto)

```protobuf
// shopee-collector
service Collector {
  rpc Fetch(FetchRequest) returns (FetchResponse);
  rpc FetchShop(FetchShopRequest) returns (FetchResponse);
}

// publisher
service Publisher {
  rpc Publish(PublishRequest) returns (PublishResponse);
  rpc ListGroups(ListGroupsRequest) returns (ListGroupsResponse);
}

// alerter
service Alerter {
  rpc CheckAndNotify(AlertRequest) returns (AlertResponse);
}
```

O Web App C# chama os microserviços Go via gRPC quando precisa de coleta, publicação ou alertas.

---

## Arquitectura alvo

```
                    ┌────────────────────────────────────────────┐
                    │          Web App C# (ASP.NET Core)          │
                    │                                            │
                    │  Controllers/Minimal API                   │
                    │  ├── CuradoriaController                   │
                    │  ├── PublicacaoController                  │
                    │  ├── LojaController                        │
                    │  ├── BuscaController                       │
                    │  ├── OnboardingController                  │
                    │  └── AdminController                       │
                    │                                            │
                    │  Application Layer (MediatR/CQRS)          │
                    │  ├── RankearProdutosQuery                  │
                    │  ├── PublicarOfertaCommand                 │
                    │  ├── SalvarBuscaCommand                    │
                    │  └── ...                                   │
                    │                                            │
                    │  Domain (C# records + value objects)       │
                    │  ├── Product, Scored, Busca                │
                    │  ├── ScoringStrategy (interface)           │
                    │  └── EligibilityPipeline                   │
                    │                                            │
                    │  Infrastructure                            │
                    │  ├── PostgreSQL (EF Core)                  │
                    │  ├── Firebase Auth (JWT validation)        │
                    │  ├── gRPC clients (Collector, Publisher)   │
                    │  └── Hangfire (scheduled jobs)             │
                    └──────┬──────────────┬──────────────┬───────┘
                           │gRPC          │gRPC          │gRPC
                    ┌──────▼──────┐ ┌─────▼──────┐ ┌────▼──────┐
                    │  shopee-    │ │  publisher  │ │  alerter  │
                    │  collector  │ │  (Go)       │ │  (Go)     │
                    │  (Go)       │ │             │ │           │
                    │  Shopee API │ │  Telegram   │ │  Telegram │
                    │  throttling │ │  WhatsApp   │ │  snapshots│
                    │  rotation   │ │  templates  │ │           │
                    └─────────────┘ └─────────────┘ └───────────┘
```

---

## Análise de viabilidade

### Vantagens

| Aspecto | Benefício |
|---|---|
| Produtividade | C# com DI, EF Core, MediatR, FluentValidation — patterns OOP naturais |
| Evolução do domínio | Rich domain model, DDD tactical patterns, refactoring com Rider/VS |
| Testes | xUnit + Moq/NSubstitute + TestContainers — ecossistema maduro |
| Observabilidade | OpenTelemetry .NET SDK + Serilog + Health Checks out of the box |
| Multi-tenancy | EF Core global query filters por tenant, scoped DI |
| Persistência | Migrar de BigQuery (analytics) para PostgreSQL (transacional) |
| I/O intensivo preservado em Go | Goroutines, channels, delays nativos — não reescreve o que funciona |
| Frontend inalterado | SPA consome REST/JSON — não importa se é Go ou C# |

### Riscos e mitigações

| Risco | Mitigação |
|---|---|
| BigQuery → PostgreSQL é uma migração de dados | Fase 1 usa dual-write; fase final migra |
| gRPC adiciona complexidade de rede | Starts local (Docker Compose), Cloud Run suporta gRPC nativo |
| Dois runtimes em produção (dotnet + go) | Docker multi-service, ou Cloud Run services separados |
| Latência gRPC entre serviços | Mesma rede VPC / Cloud Run multi-container. Latência < 5ms |
| Firebase Auth no C# | Library official: `FirebaseAdmin` SDK for .NET |
| Custo operacional duplica? | Cloud Run escala a zero para ambos. Custo ~igual |
| Perda de funcionalidade na migração | Transição gradual (ver plano abaixo) |

### O que NÃO muda

- **Frontend** (SvelteKit SPA) — apenas muda `VITE_API_BASE` se URL mudar
- **Cloudflare Worker** — proxy continua apontando para o endpoint (C# ou Go)
- **Cloud Scheduler** — pode continuar a chamar `/api/coletar` (agora servido pelo C#, que delega ao Go via gRPC)
- **Firebase Auth** — ambos validam o mesmo JWT
- **Shopee API auth** — HMAC-SHA256 fica no microserviço Go (intocado)

---

## Plano de migração gradual

### Fase 0 — Preparação (1-2 sprints)

- [ ] Criar repo `garimpei-api` (C# solution) com estrutura Clean Architecture
- [ ] Definir `.proto` files para os 3 serviços gRPC
- [ ] Criar projecto `Garimpei.Collector` (extrair `source` + `coleta` para gRPC server)
- [ ] Manter monólito Go a funcionar em paralelo (feature flag por rota)

### Fase 1 — Coexistência (2-3 sprints)

- [ ] Web App C# serve rotas novas (ex: T-0004 multi-tenant, T-0005 alertas)
- [ ] Go monólito continua a servir rotas existentes
- [ ] Cloudflare Worker faz routing: `/api/v2/*` → C#, `/api/*` → Go
- [ ] Dual-write: C# escreve em PostgreSQL + BigQuery (transição)
- [ ] `shopee-collector` Go serve via gRPC (chamado por ambos)

### Fase 2 — Migração de rotas (3-4 sprints)

- [ ] Migrar handlers por domínio (curadoria → lojas → publicação → admin)
- [ ] Cada grupo migrado: testes E2E validam paridade
- [ ] Frontend aponta gradualmente para v2 (feature flag no SPA)
- [ ] PostgreSQL torna-se fonte primária; BigQuery vira analytics-only (export)

### Fase 3 — Descomissionar monólito Go (1 sprint)

- [ ] Remover rotas migradas do Go
- [ ] Go vira apenas os 3 microserviços gRPC (collector, publisher, alerter)
- [ ] Cloudflare Worker aponta 100% para C#
- [ ] Documentação actualizada

---

## Impacto no backlog

### Novas tarefas (T-0009+)

| ID | Título | Fase | Estimativa |
|---|---|---|---|
| T-0009 | Setup C# solution (Clean Architecture + Docker) | 0 | M |
| T-0010 | Proto definitions + shopee-collector gRPC server | 0 | M |
| T-0011 | Publisher gRPC server (extract publish package) | 0 | M |
| T-0012 | C# Web App — auth middleware + health + CI | 0 | M |
| T-0013 | Multi-tenant em C# (EF Core + PostgreSQL) | 1 | G |
| T-0014 | Curadoria controller + scoring port em C# | 1 | G |
| T-0015 | Routing split (Cloudflare Worker v1→v2) | 1 | P |
| T-0016 | Migrar handlers de publicação para C# | 2 | G |
| T-0017 | Migrar handlers de lojas/buscas para C# | 2 | G |
| T-0018 | PostgreSQL como fonte primária + export BQ | 2 | G |
| T-0019 | Descomissionar monólito Go | 3 | M |

### Tarefas existentes afetadas

- **T-0002 (conversões)** — pode ser implementada já no C# (Fase 1) se timing alinhar
- **T-0004 (multi-tenant)** — candidata natural para ser a primeira feature em C#
- **T-0005 (alertas config)** — pode usar o alerter Go via gRPC + config em C#

---

## Stack tecnológica proposta (C#)

| Camada | Biblioteca |
|---|---|
| Framework | ASP.NET Core 9 (Minimal API ou Controllers) |
| ORM | Entity Framework Core 9 + Npgsql |
| CQRS | MediatR |
| Validação | FluentValidation |
| Auth | Microsoft.AspNetCore.Authentication.JwtBearer + FirebaseAdmin |
| Logging | Serilog + OpenTelemetry |
| Jobs | Hangfire (PostgreSQL storage) |
| gRPC Client | Grpc.Net.Client |
| Testes | xUnit + NSubstitute + TestContainers + Bogus |
| Docs API | Swashbuckle / NSwag (OpenAPI) |
| Container | .NET 9 Alpine image (~80MB) |

---

## Métricas de sucesso

1. **Paridade funcional** — todas as rotas migradas passam nos mesmos testes E2E
2. **Latência** — p95 das rotas migradas ≤ p95 actual (≤ 200ms excl. Shopee)
3. **Disponibilidade** — zero downtime durante migração (coexistência)
4. **Produtividade** — novos features implementados 2× mais rápido após Fase 2
5. **Cobertura** — ≥80% em C#, mantida em Go

---

## Decisão pendente

Antes de iniciar a Fase 0, decidir:

1. **PostgreSQL vs manter BigQuery?** — PG é melhor para CRUD transacional; BQ pode ficar só para analytics/export
2. **Mono-repo ou multi-repo?** — Mono-repo (Go + C# + protos) simplifica CI; multi-repo isola deployments
3. **Cloud Run multi-container ou serviços separados?** — Multi-container é mais simples (sidecar gRPC); separados escalam independentemente
4. **Timeline:** iniciar Fase 0 na sprint S28 (próxima semana)?

## Consequências

### Se aceitar

- O sistema ganha uma base C# moderna com patterns OOP avançados
- Microserviços Go ficam estáveis e raramente precisam de alterações
- Multi-tenancy e features futuras são implementados com maior velocidade
- Custo operacional sobe levemente (2 runtimes), mitigado por scale-to-zero

### Se rejeitar

- Continua tudo em Go — funciona, mas produtividade limitada pelo conforto com a linguagem
- Multi-tenancy e CQRS são possíveis em Go, mas mais verbosos sem DI nativo
- Sem custo adicional de migração

### Se adiar

- Implementar T-0002 e T-0004 em Go agora (já planejados)
- Reavalar após S28 com mais dados sobre produtividade e necessidades
