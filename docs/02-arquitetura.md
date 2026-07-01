# Arquitetura

## Visão geral

O Garimpei usa uma **arquitetura poliglota orientada a serviços**, onde cada linguagem
é usada no domínio em que é mais produtiva:

- **C# (ASP.NET Core 10)** — API principal: CRUD, autenticação, multi-tenant, orquestração
- **Go (gRPC)** — microserviços de I/O intensivo: coleta Shopee, publicação, alertas, scheduling
- **Python (FastAPI)** — analytics e IA: queries BigQuery, detecção de padrões, séries temporais
- **SvelteKit** — frontend SPA servido via CDN global (Cloudflare Pages)

Esta arquitetura substituiu um monólito Go (~12.000 linhas) em julho de 2026,
conforme documentado na ADR-0012.

---

## Diagrama de arquitetura

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Cloudflare (Edge global)                           │
│                                                                         │
│  ┌─────────────────┐        ┌──────────────────────────────────┐       │
│  │  Pages (CDN)    │        │  Worker (routing inteligente)    │       │
│  │  Frontend SPA   │        │                                  │       │
│  │  SvelteKit      │        │  /api/*  → Cloud Run (C#)        │       │
│  │  ~50ms TTFB     │        │  /docs/* → Pages (Starlight)     │       │
│  │                 │        │  /*      → Pages (Frontend)      │       │
│  └─────────────────┘        └──────────────┬───────────────────┘       │
└─────────────────────────────────────────────┼───────────────────────────┘
                                              │ HTTPS
┌─────────────────────────────────────────────▼───────────────────────────┐
│          Cloud Run multi-container (southamerica-east1, gen2)             │
│          Scale 0→3 | Container deps | Startup probes                    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  garimpei-api (C# .NET 10) — ingress container :8080            │   │
│  │                                                                 │   │
│  │  ┌──────────────────────────────────────────────────────────┐   │   │
│  │  │  Middleware Pipeline                                      │   │   │
│  │  │  Serilog → Auth (Firebase JWT) → TenantMiddleware        │   │   │
│  │  └──────────────────────────────────────────────────────────┘   │   │
│  │                                                                 │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐  │   │
│  │  │ Curadoria  │ │ Publicação │ │   Buscas   │ │   Admin    │  │   │
│  │  │ Scoring    │ │ → gRPC pub │ │  CRUD (PG) │ │  /admin/me │  │   │
│  │  │ 4 fontes   │ │            │ │            │ │            │  │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘  │   │
│  │                                                                 │   │
│  │  ┌──────────────────────────────────────────────────────────┐   │   │
│  │  │  Infrastructure Layer                                     │   │   │
│  │  │  EF Core (PG) | gRPC Clients | HttpClient (analyzer)    │   │   │
│  │  │  Multi-tenant (global query filters por owner_uid)        │   │   │
│  │  │  OpenTelemetry (traces + metrics → OTLP)                 │   │   │
│  │  └──────────────────────────────────────────────────────────┘   │   │
│  └────────────┬──────────────┬──────────────┬──────────┬──────────┘   │
│               │gRPC          │gRPC          │gRPC      │HTTP          │
│               │localhost     │localhost     │localhost  │localhost     │
│  ┌────────────▼──┐  ┌───────▼───┐  ┌───────▼──┐  ┌───▼──────────┐   │
│  │  collector    │  │ publisher │  │  alerter │  │   analyzer   │   │
│  │  (Go :50051)  │  │ (Go:50052)│  │ (Go:50053)│  │ (Py :8060)  │   │
│  │               │  │           │  │          │  │              │   │
│  │  Shopee API   │  │ Telegram  │  │ Telegram │  │  BigQuery    │   │
│  │  GraphQL      │  │ Bot API   │  │ preço    │  │  pandas      │   │
│  │  HMAC-SHA256  │  │ Meta WA   │  │ alertas  │  │  novidades   │   │
│  │  throttling   │  │ Cloud API │  │ snapshot │  │  quedas      │   │
│  │  paginação    │  │ retry     │  │ compare  │  │  evolução    │   │
│  └───────────────┘  └───────────┘  └──────────┘  │  estatísticas│   │
│                                                   └──────┬───────┘   │
│  ┌───────────────┐                                      │            │
│  │  scheduler    │                                      │            │
│  │  (Go :50054)  │──── orquestra collector/alerter      │            │
│  │  robfig/cron  │     via gRPC (timezone BRT)          │            │
│  └───────────────┘                                      │            │
└─────────────────────────────────────────────────────────┼────────────┘
                          │                               │
              ┌───────────▼──────────┐      ┌─────────────▼──────────┐
              │  PostgreSQL (Neon)   │      │  BigQuery (GCP)        │
              │                      │      │                        │
              │  • Products          │      │  • snapshots           │
              │  • Buscas            │      │  • conversões          │
              │  • Tenants           │      │  • métricas históricas │
              │  • Configs           │      │  • séries temporais    │
              │                      │      │                        │
              │  Multi-tenant:       │      │  Read-only pelo        │
              │  owner_uid filter    │      │  analyzer Python       │
              └──────────────────────┘      └────────────────────────┘
```

---

## Princípios arquiteturais

### 1. Cada linguagem no que faz de melhor

| Linguagem | Força utilizada | Exemplo no Garimpei |
|-----------|----------------|---------------------|
| **C#** | OOP, DI nativo, EF Core, patterns (CQRS, Mediator) | Multi-tenant com global query filters automáticos |
| **Go** | Goroutines, channels, I/O concorrente, binários pequenos | Throttling de 200ms/60s na coleta Shopee sem bloquear |
| **Python** | pandas, BigQuery SDK, ecossistema ML | Detecção de quedas de preço em séries temporais |
| **Svelte** | Reatividade, bundle pequeno, compilação AOT | SPA com ~50ms TTFB via CDN |

### 2. Separação de responsabilidades por bounded context

- **Transacional** (C#): "O que o usuário vê e faz" — CRUD, auth, validação
- **I/O intensivo** (Go): "Comunicação com o mundo externo" — APIs, throttling, retry
- **Analítico** (Python): "O que aprendemos dos dados" — padrões, tendências, IA
- **Apresentação** (Svelte): "Como o usuário interage" — reativo, offline-capable

### 3. Isolamento de falhas

Se a Shopee está fora do ar, o sidecar `collector` falha mas o C# continua servindo dados do PostgreSQL. O frontend continua acessível via CDN. O scheduler tenta novamente no próximo cron.

### 4. Scale-to-zero

Todos os containers escalam a zero quando não há tráfego. O primeiro request após inatividade tem cold start (~3-5s C#, ~500ms Go, ~2s Python), mas requests subsequentes são <100ms.

### 5. Contratos tipados (Protocol Buffers)

Alterações na interface entre serviços são explícitas (`.proto` files). O CI detecta breaking changes via `buf breaking`. Stubs Go e C# são pré-gerados e commitados — zero dependência de tooling externo no build.

---

## Vantagens detalhadas

### Produtividade de desenvolvimento

| Antes (Go monólito) | Agora |
|---------------------|-------|
| Cada novo endpoint: handler + validação manual + error handling manual | Minimal API: 5 linhas para um endpoint com validação |
| DI manual (construtor explícito em cada package) | DI nativo do ASP.NET Core (scoped, singleton, transient) |
| Multi-tenancy: `WHERE owner_uid = ?` em cada query | EF Core global filter: automático em todas as queries |
| Auth: parse manual do JWT + lookup de claims | `[Authorize]` ou `RequireAuthorization()` + claims nativo |
| Testes: mock de BigQuery pesado | InMemory DB + TestContainers para integração |

### Segurança (Multi-tenancy)

O monólito Go dependia de cada handler lembrar de filtrar por `owner_uid`. Um erro = vazamento de dados entre tenants.

Agora:
```csharp
// Configurado UMA VEZ no DbContext:
entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);

// TODA query automaticamente filtra. Impossível esquecer.
var produtos = await db.Products.ToListAsync(); // já filtrado!
```

### Performance e custo

| Camada | Antes | Agora | Ganho |
|--------|-------|-------|-------|
| Frontend TTFB | ~800ms (Cloud Run cold start) | ~50ms (CDN Cloudflare) | **16x mais rápido** |
| API cold start | ~3s (monólito pesado) | ~2s (C# slim + sidecars paralelos) | Sidecars ficam prontos |
| WhatsApp | Maytapi ($19/mês + $0.01/msg) | Meta Cloud API (grátis) | **-$19/mês** |
| BigQuery | Usado para CRUD (ineficiente) | Usado só para analytics (eficiente) | Queries mais baratas |
| Hosting frontend | Cloud Run (paga por request) | Cloudflare Pages (grátis) | **-$0/mês** |

### Resiliência

```
Cenário: Shopee API retorna erro 429 (rate limit)

Antes (monólito):
  → Todo o servidor trava por 30s (goroutine bloqueada no retry)
  → Frontend não responde

Agora:
  → Sidecar collector retenta silenciosamente
  → C# continua servindo CRUD do PostgreSQL
  → Frontend carrega normalmente da CDN
  → Scheduler tenta novamente no próximo cron cycle
```

### Evolução para IA (preparado)

O analyzer Python está pronto para evoluir sem tocar C# ou Go:

1. **Scoring ML**: substituir rule-based (45%/35%/20%) por modelo treinado em conversões reais
2. **Recomendação**: "produtos similares ao que você publicou" via embeddings
3. **Detecção de anomalias**: identificar produtos-fantasma automaticamente
4. **Previsão de demanda**: prever quais produtos vão vender baseado em tendências

Tudo isso em pandas + scikit-learn + BigQuery, sem afetar a API principal.

---

## Stack tecnológica completa

| Camada | Tecnologia | Versão |
|--------|-----------|--------|
| Web App | ASP.NET Core (Minimal API) | 10.0 |
| ORM | Entity Framework Core + Npgsql | 10.0 |
| CQRS (futuro) | MediatR | 12.4 |
| Validação (futuro) | FluentValidation | 11.11 |
| Auth | Firebase Auth (JWT Bearer) | — |
| Observabilidade | OpenTelemetry + Serilog | 1.12 / 9.0 |
| Microserviços I/O | Go + gRPC | 1.26 |
| gRPC framework | google.golang.org/grpc | 1.82 |
| Scheduling | robfig/cron/v3 | 3.0 |
| Analytics | Python + FastAPI + pandas | 3.13 / 0.115 / 2.2 |
| BigQuery | google-cloud-bigquery | 3.31 |
| Frontend | SvelteKit 2 + Svelte 5 + Vite 8 | — |
| DB transacional | PostgreSQL (Neon serverless) | 17 |
| DB analytics | BigQuery | — |
| Hosting frontend | Cloudflare Pages | — |
| Routing | Cloudflare Workers | — |
| Container runtime | Cloud Run (gen2, multi-container) | — |
| Registry | Artifact Registry (GCP) | — |
| Secrets | Secret Manager (GCP) | — |
| CI/CD | GitHub Actions | — |
| Contratos | Protocol Buffers (buf) | v2 |
| Lint Go | golangci-lint | latest |
| Arch Go | arch-go | latest |
| Lint Python | ruff | latest |
| Lint Frontend | eslint + stylelint | — |
| Testes Go | go test | — |
| Testes C# | xUnit + InMemory DB | 2.9 |
| Testes Frontend | Vitest + Playwright | — |
| Dead code | Knip (JS) + golangci-lint (Go) | — |

---

## Persistência (estratégia dual)

### PostgreSQL (dados transacionais)

- **Quando usar**: CRUD, dados que o usuário cria/edita, multi-tenant
- **Schema**: EF Core code-first migrations
- **Acesso**: C# (EF Core) — nunca acessado diretamente pelos sidecars Go
- **Tabelas**: Products, Buscas, Tenants (+ futuras: Publicacoes, Favoritos, AlertConfigs)
- **Multi-tenant**: global query filter automático por `owner_uid`
- **Hosting**: Neon (serverless, free tier, sa-east-1)

### BigQuery (dados analíticos)

- **Quando usar**: séries temporais, histórico, queries analíticas pesadas
- **Acesso**: Python (analyzer) para leitura, Go (scheduler/collector) para escrita
- **Tabelas**: snapshots, conversoes, eventos
- **Retenção**: ilimitada (BigQuery free tier: 10GB storage)
- **Queries**: analyzer Python faz as queries complexas (joins, window functions)

### Regra de ouro

> **Se o usuário cria/edita → PostgreSQL**
> **Se o sistema coleta/calcula → BigQuery**

---

## Multi-tenancy em detalhe

```
Request HTTP com JWT Firebase
       │
       ▼
┌─── TenantMiddleware ───┐
│ Extrai "user_id" claim │
│ Seta TenantContext      │
│ (scoped per-request)    │
└────────────┬────────────┘
             │
             ▼
┌─── EF Core DbContext ──┐
│ Global Query Filter:    │
│ WHERE owner_uid = @uid  │
│ (automático em TODA     │
│  query, impossível      │
│  esquecer)              │
└────────────┬────────────┘
             │
             ▼
┌─── SaveChangesAsync ───┐
│ Entidades novas         │
│ recebem owner_uid       │
│ automaticamente         │
└─────────────────────────┘
```

**Garantias:**
- Tenant A nunca vê dados de Tenant B (filtro no banco, não no código)
- Novos endpoints herdam isolamento automaticamente (zero config)
- Admin pode bypassar (futuro: `IgnoreQueryFilters()`)

---

## Microserviços (detalhe por serviço)

### collector (Go, gRPC :50051)

**Responsabilidade:** buscar produtos na API de afiliados da Shopee

- Autenticação HMAC-SHA256 (AppID + Secret + timestamp)
- Paginação com throttling (200ms entre páginas, 60s entre lojas)
- Rotação de catálogo (cursor por loja, full-scan tracking)
- Suporta busca por keyword ou por shop_id

**RPCs:**
- `Fetch(keyword, limit)` → produtos rankeados por comissão
- `FetchShop(shop_id, limit)` → produtos de uma loja específica

### publisher (Go, gRPC :50052)

**Responsabilidade:** enviar ofertas para canais de comunicação

- Telegram: Bot API (sendMessage, sendPhoto com inline keyboard)
- WhatsApp: Meta Cloud API (texto + imagem com caption)
- Multi-destino: dispatcher roteia para o canal correto
- Rate limiting e retry com backoff

**RPCs:**
- `Publish(channel, group_id, content)` → envia mensagem
- `ListGroups(channel)` → lista destinos configurados

### alerter (Go, gRPC :50053)

**Responsabilidade:** detectar variações de preço e notificar

- Compara snapshots da janela de dias configurada
- Threshold configurável (default: 15%)
- Filtro "apenas quedas" (oportunidades)
- Notificação via Telegram (formatação HTML)

**RPCs:**
- `CheckAndNotify(owner_uid, rules[])` → verifica e notifica

### scheduler (Go, gRPC :50054)

**Responsabilidade:** orquestrar jobs periódicos

- Cron nativo (robfig/cron, timezone America/Sao_Paulo)
- Chama collector, publisher, alerter via gRPC
- Gerenciável em runtime (criar/pausar/deletar jobs)

**RPCs:**
- `SetSchedule(job_id, cron, params)` → criar/atualizar job
- `ListJobs(status_filter)` → listar jobs registrados
- `TriggerJob(job_id)` → executar job manualmente

### analyzer (Python, REST :8060)

**Responsabilidade:** queries analíticas no BigQuery

- Novidades: produtos novos detectados entre snapshots
- Quedas: variação negativa de preço acima do threshold
- Evolução: série temporal de preço por loja
- Estatísticas: resumo por categoria (médias, medianas)

**Endpoints:**
- `GET /novidades?busca_id=X&dias=7`
- `GET /quedas?dias=7&threshold=0.15&limit=50`
- `GET /evolucao?dias=30`
- `GET /estatisticas?dias=30`
- `GET /health`

---

## Deploy e operação

### Cloud Run multi-container

6 containers na mesma instância, comunicação via localhost:

| Container | CPU | RAM | Probe |
|-----------|-----|-----|-------|
| garimpei-api (C#) | 1.0 | 512Mi | HTTP /health |
| collector (Go) | 0.5 | 256Mi | TCP :50051 |
| publisher (Go) | 0.25 | 128Mi | TCP :50052 |
| alerter (Go) | 0.25 | 128Mi | TCP :50053 |
| scheduler (Go) | 0.25 | 128Mi | TCP :50054 |
| analyzer (Python) | 0.5 | 256Mi | HTTP /health :8060 |

**Total:** 2.75 vCPU, 1408Mi RAM (quando ativo). **Zero quando idle** (scale-to-zero).

### CI Pipeline

```
push main → GitHub Actions (ci.yml)
  │
  ├─ go: build + test + lint + arch-go + docs-check + file-size
  ├─ csharp: restore + build + test (com PostgreSQL service)
  ├─ python: ruff lint + syntax check
  ├─ proto: buf lint + sync check (Go + C# stubs atualizados?)
  ├─ frontend: npm ci + build + lint:css + lint:js + vitest
  ├─ docker: build all 6 images (validação)
  └─ docs-deploy: sync + build + deploy Cloudflare Pages
```

### Routing (Cloudflare Worker)

```javascript
/api/*   → Cloud Run (C# garimpei-v2)    // Backend
/docs/*  → Cloudflare Pages (Starlight)  // Documentação
/*       → Cloudflare Pages (SvelteKit)  // Frontend
```

Feature flags:
- `V2_ENABLED`: ativa/desativa routing para C# (rollback instantâneo)
- `PAGES_URL`: URL do frontend Pages
- `DOCS_URL`: URL do docs Pages

---

## Qualidade e validação

### Testes

| Stack | Framework | Testes | Cobertura |
|-------|-----------|--------|-----------|
| Go (internal) | go test | source 87%, publish 62%, store 36% | Paths críticos |
| Go (services) | go test | 12 testes (validações + fluxos) | 11-33% |
| C# | xUnit | 10 testes (multi-tenant, persistence) | Isolamento garantido |
| Frontend | Vitest + Playwright | 109 unitários + E2E | Componentes + fluxos |

### Análise estática

| Ferramenta | O que valida |
|-----------|-------------|
| golangci-lint | 50+ linters Go (estilo, bugs, performance) |
| arch-go (9 regras) | Dependências entre pacotes (100% compliance) |
| buf lint | Protos seguem STANDARD rules |
| buf breaking | Detecta breaking changes nos contratos |
| proto sync check | Stubs commitados == protos atuais |
| dotnet build (warnings=errors) | Zero warnings no C# |
| ruff | Lint Python (fast, compatible com flake8) |
| eslint + stylelint | Lint JS/CSS frontend |
| Knip | Dead code/exports no frontend |
| check-file-size | Máx 400 linhas por arquivo (exceto gen/) |
| check-api-spec-sync | OpenAPI spec == endpoints implementados |

### Validação de drift

O CI garante que:
1. Docs gerados estão atualizados (`make docs-check`)
2. Proto stubs estão sincronizados com `.proto` files
3. OpenAPI spec reflete endpoints reais (zero 404)
4. Zero código morto (Knip, golangci-lint unused)

---

## ADRs (Architecture Decision Records)

| ADR | Decisão | Data |
|-----|---------|------|
| [0003](/docs/decisoes/0003-deploy-gcp/) | Deploy no GCP (Cloud Run) | 2026-06 |
| [0012](/docs/decisoes/0012-migracao-csharp-go-microservices/) | Migração para C# + Go microservices | 2026-06 |
| [0013](/docs/decisoes/0013-whatsapp-meta-cloud-api/) | WhatsApp: Maytapi → Meta Cloud API | 2026-07 |
| [0014](/docs/decisoes/0014-analyzer-python-fastapi/) | Analyzer Python (FastAPI + BigQuery) | 2026-07 |
