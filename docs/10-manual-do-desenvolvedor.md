# Manual do Desenvolvedor

## Pré-requisitos

- Go 1.24+
- .NET 10 SDK
- Python 3.13+ com `uv`
- Node 24+ com `bun`
- Docker (para PostgreSQL local)
- [mise](https://mise.jdx.dev/) (task runner)
- `buf` CLI (para protos)
- `golangci-lint`, `arch-go` (via mise)

## Setup inicial

```bash
# Clone
git clone https://github.com/fmarquesfilho/garimpo.git
cd garimpo

# Instalar ferramentas via mise
mise install

# Frontend
cd web && bun install && cd ..

# C# dependencies
dotnet restore src/Garimpei.sln

# Go dependencies
go mod download

# PostgreSQL local (Docker)
mise run up
```

## Comandos do mise

### Build

| Comando | Descrição |
|---------|-----------|
| `mise run build` | Build completo (Go + C#) |
| `mise run build:go` | Build dos serviços Go |
| `mise run build:csharp` | Build da solution C# |

### Testes

| Comando | Descrição |
|---------|-----------|
| `mise run test` | Testes unitários (Go + C# + Web) |
| `mise run test:all` | Todos os testes (unitários + E2E locais) |
| `mise run test:go` | Testes Go |
| `mise run test:csharp` | Testes C# (sobe PostgreSQL via Docker) |
| `mise run test:web` | Vitest (unit + contrast + theme) |
| `mise run test:unit` | Apenas testes unitários rápidos |
| `mise run test:integration:cache` | Cache sidecar + collector divergence + full suite |

### Testes E2E (produção — NÃO rodar no CI)

| Comando | Descrição |
|---------|-----------|
| `mise run test:e2e-full-pipeline` | Fluxo completo cross-service com trace único |
| `mise run test:e2e-coleta` | Scheduler → Collector → BigQuery |
| `mise run test:e2e-alertas` | Analyzer → Cloud Tasks → Publisher |
| `mise run test:e2e-publicacoes` | API → Publisher gRPC → Telegram |
| `mise run test:e2e-dashboard` | Change detection + polling + multi-tenant |
| `mise run test:e2e-services` | Health + endpoints de todos os services |
| `mise run test:e2e-scheduler` | Agendamento + trigger manual |
| `mise run test:e2e-traces` | Propagação OTel (traceparent) |
| `mise run test:e2e-analyzer` | BigQuery queries via Analyzer |
| `mise run test:e2e-smoke` | Smoke test rápido (health + auth) |
| `mise run test:e2e-novos` | Pipeline Novos: Collect → BigQuery → Analyzer → Dashboard |
| `mise run test:e2e-local` | 24 testes Playwright com mock |
| `mise run test:e2e-prod` | 8 testes Playwright com APIs reais |

### Testes E2E de integração real (APIs externas)

| Comando | Descrição |
|---------|-----------|
| `mise run test:e2e:buscas-agendadas` | API + Collector + Scheduler (Shopee real) |
| `mise run test:e2e:lojas` | ResolveShop (Shopee real) |
| `mise run test:e2e:publicar` | Publicações (Telegram real) |
| `mise run test:e2e:alertas` | Alertas + Novidades (Analyzer + BigQuery real) |

### Lint

| Comando | Descrição |
|---------|-----------|
| `mise run lint` | Todos os linters |
| `mise run lint:go` | golangci-lint + arch-go (regras de dependência) |
| `mise run lint:python` | Ruff no analyzer |
| `mise run lint:web` | ESLint + Stylelint + Semgrep |

### Drift Checks (integridade)

| Comando | Descrição |
|---------|-----------|
| `mise run checks` | Roda todos os checks abaixo |
| `mise run check:api-contract` | Frontend ↔ Backend (rotas existem?) |
| `mise run check:schema-sync` | BigQuery ↔ Go ↔ C# ↔ Analyzer |
| `mise run check:service-contracts` | Registry + schemas + flows |
| `mise run check:data-ownership` | Fronteiras de dados (PG vs BQ) |
| `mise run check:stale-refs` | Dead code, serviços removidos |
| `mise run check:config-consistency` | Dataset, portas, URLs |
| `mise run check:docs-drift` | Documentação gerada em dia? |
| `mise run check:file-size` | Limite 400 linhas (testes: 900) |
| `mise run check:fixtures-contract` | Golden files válidos |
| `mise run check:fixtures-frontend` | payloadToConfig vs golden |
| `mise run check:fixtures-crosslang` | DeriveCollectionKeys igual em Go/Python/JS |
| `mise run check:api-spec-sync` | OpenAPI spec sincronizada com endpoints |
| `mise run check:format` | Prettier (sem alterar) |
| `mise run check:rules-schema` | busca-rules.json vs schema |
| `mise run check:ui-coverage` | Cobertura da biblioteca de componentes |
| `mise run check:comment-quality` | Anti-patterns em comentários |
| `mise run check:comment-quality` | Anti-patterns em comentários |

### Debug (produção)

| Comando | Descrição |
|---------|-----------|
| `mise run debug:health` | Health check de todos os serviços |
| `mise run debug:logs` | Logs de produção (Cloud Logging) |
| `mise run debug:trace <id>` | Investigar trace específico no Cloud Trace |

Opções de `debug:logs`:
```bash
mise run debug:logs -- --severity ERROR --minutes 60
mise run debug:logs -- --service scheduler --keyword falhou
mise run debug:logs -- --trace-id abc123def456
```

### Docs

| Comando | Descrição |
|---------|-----------|
| `mise run docs` | Gera toda a documentação |
| `mise run docs:sync` | Sincroniza para docs-site |
| `mise run docs:env` | Extrai variáveis de ambiente |
| `mise run docs:er` | Gera Mermaid ER do BigQuery |
| `mise run docs:board` | Kanban do backlog |
| `mise run docs:check` | CI: verifica docs geradas |

### Proto

| Comando | Descrição |
|---------|-----------|
| `mise run proto:generate` | Gera stubs Go e C# |
| `mise run proto:lint` | Lint nos .proto |
| `mise run proto:breaking` | Verifica breaking changes |

### Deploy e DB

| Comando | Descrição |
|---------|-----------|
| `mise run deploy:migrate` | EF Core migrations em produção |
| `mise run db:reset` | Reset banco local |
| `mise run db:reset -- --prod` | Reset banco de produção (Neon) |
| `mise run db:reset -- --prod --bq` | Reset BigQuery de produção |

### Backlog e Geração

| Comando | Descrição |
|---------|-----------|
| `mise run backlog:create` | Cria nova tarefa no backlog |
| `mise run gen:api` | Gera OpenAPI spec |
| `mise run gen:api-reference` | Gera referência de API |

### Outros

| Comando | Descrição |
|---------|-----------|
| `mise run up` | Sobe serviços (dev local) |
| `mise run down` | Para serviços |
| `mise run logs` | Logs dos containers |
| `mise run ci` | Simula CI completo localmente |
| `mise run prepush` | Verificação pré-push (~1min) |

---

## Observabilidade: Fluxo de Snapshots

### Como verificar que snapshots estão fluindo para BigQuery

O pipeline de coleta funciona assim:

```
Scheduler (cron 8h) → Collector.Collect (gRPC)
                        ├→ Busca produtos (Shopee API, ~500ms)
                        ├→ Enfileira snapshot no channel (async, ~0ms)
                        └→ Retorna produtos ao Scheduler
                           ↓ (background goroutine)
                        BigQuery streaming insert (~100ms)
                           ↓
                        Analyzer queries → Dashboard mostra dados
```

### 4 formas de verificar:

**1. Dashboard (UI) — `/estatisticas`**

A seção "Saúde" mostra em tempo real:
- `Última coleta`: timestamp da última coleta
- `Coletas 24h`: quantas coletas aconteceram
- Status: `ok` (< 6h) / `atrasado` (> 6h) / `sem_dados`

Se `coletas_24h` cresce, snapshots estão fluindo.

**2. CLI rápido**

```bash
mise run debug:health
# Mostra: collector ok, analyzer ok (X coletas)
```

**3. Trace no Cloud Trace**

```bash
# Triggera coleta com trace rastreável
mise run test:e2e-full-pipeline
# Output: "Trace raiz: abc123..."

# Investiga o trace
mise run debug:trace abc123
# Mostra: timeline de spans com cada service
```

No Console GCP: `https://console.cloud.google.com/traces/list?project=garimpo-500114`

**4. BigQuery direto**

```sql
SELECT keyword, COUNT(*) as snapshots, MAX(coletado_em) as ultimo
FROM `garimpo-500114.garimpei_prod.snapshots`
WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 24 HOUR)
GROUP BY keyword
ORDER BY ultimo DESC
```

---

## Detecção de Produtos Novos

### Como funciona

O Analyzer detecta "novos" comparando snapshots na janela de tempo:

1. **Coleta** (Scheduler → Collector) persiste snapshot no BigQuery com keyword + timestamp
2. **Analyzer** (`GET /novidades?busca_id=X&dias=7`) compara snapshots:
   - Produto com `aparicoes == 1` na janela = **novo** (primeira vez visto)
   - Produto com variação de preço > 1% = **variação**
3. **Frontend** chama `/api/lojas/novidades` para cada busca que tem lojas monitoradas
4. **Grid** mostra com `_fonte: 'novo'` e badge 🆕

### Fluxo no frontend (página principal `/`)

```
BuscaEngine.executarBusca()
  → fontes.novos == true?
    → carregarOportunidades(buscasComLojas, nomesLojas)
      → para cada busca: buscarNovidades({ buscaId: shop_id ou keyword })
        → GET /api/lojas/novidades?busca_id={X}&dias=7
          → Analyzer: compara snapshots no BigQuery
            → { produtos_novos: [...], variacoes: [...] }
      → extrairNovos(resultado) → marca _fonte: 'novo'
      → extrairQuedas(resultado) → marca _fonte: 'queda'
  → contagens.novos = resultados.filter(p => p._fonte === 'novo').length
  → badge 🆕 atualiza no ToggleGroup
```

### Pré-requisitos para "Novos" funcionar

1. **Ter lojas monitoradas** — o filtro 🆕 só funciona com `buscasComLojas.length > 0`
2. **Coletas recentes** — precisa de pelo menos 2 coletas na janela (7 dias) para comparar
3. **Fonte ativa** — toggle 🆕 Novos precisa estar ativo (é ativo por padrão via `busca-rules.json`)

### Como testar que está funcionando

**Forma rápida (~15s, sem esperar cron):**

```bash
mise run test:e2e-novos
```

Este teste valida instantaneamente o pipeline inteiro:
1. Triggera coleta via Collector
2. Verifica snapshots no BigQuery
3. Chama Analyzer `/novidades` (detecção direta)
4. Chama `/oportunidades/agora` (endpoint do Dashboard)
5. Simula o que o Frontend faz (por loja)
6. Verifica quedas via proxy
7. Valida trace OTel

Se quiser investigar os spans:
```bash
# O output mostra os trace_ids de cada step
mise run debug:trace <trace_id>
```

**Forma manual (UI):**
- Página principal com loja monitorada → toggle 🆕 ativo → badge > 0
- `/estatisticas` → seção Oportunidades → "X novos"

### Quando NÃO funciona (e porquê)

| Sintoma | Causa provável | Como verificar |
|---------|---------------|----------------|
| Badge 🆕 mostra 0 | Nenhuma loja monitorada, ou coletas < 2 na janela | `mise run test:e2e-novos` (step 2) |
| Badge mostra número mas grid vazio | Filtro de comissão/vendas excluindo novos | Desativar filtros e testar |
| Novos na estatísticas mas não na principal | Endpoints diferentes: `/oportunidades/agora` agrega tudo, `/lojas/novidades` filtra por loja | `mise run test:e2e-novos` (steps 3 vs 4) |
| Novos sumindo após 7 dias | Comportamento correto — janela padrão é 7 dias | — |
| Dashboard mostra 10 novos, página principal 0 | Novos vêm de keywords (serum) não de shop_ids | Verificar se busca tem loja monitorada com coletas |

---

## Arquitetura de Serviços

```
┌────────────────────────────────────────────────────────────────────────┐
│                  Cloud Run (multi-container)                             │
│                                                                         │
│  ┌─────────────┐  ┌───────────┐  ┌──────────┐  ┌──────────┐          │
│  │ C# API      │  │ Scheduler │  │Collector │  │ Publisher│          │
│  │ (ingress)   │  │ (gRPC+HTTP│  │ (gRPC)   │  │ (gRPC)   │          │
│  │ port 8080   │  │ 50054+8054│  │ 50051    │  │ 50052    │          │
│  └──────┬──────┘  └─────┬─────┘  └────┬─────┘  └────┬─────┘          │
│         │                │              │              │                │
│         │ gRPC           │ gRPC         │ Shopee API   │ TG/WA          │
│         ▼                ▼              ▼              ▼                │
│  ┌─────────────┐  ┌───────────┐  ┌──────────┐  ┌───────────┐         │
│  │ Cache       │  │Cloud Tasks│  │ BigQuery │  │ Telegram  │         │
│  │ Sidecar     │  │           │  │          │  │ WhatsApp  │         │
│  │ port 50055  │  │           │  │          │  │           │         │
│  │ (L2 LRU)   │  │           │  │          │  │           │         │
│  └──────┬──────┘  └───────────┘  └────┬─────┘  └───────────┘         │
│         │                              │                               │
│  ┌──────┴──────┐            ┌──────────┴──────────┐                   │
│  │ PostgreSQL  │            │   Analyzer (Python)  │                   │
│  │ (Neon)      │            │   FastAPI port 8060  │                   │
│  └─────────────┘            └─────────────────────┘                   │
└────────────────────────────────────────────────────────────────────────┘

         Cloudflare Edge (L1 Cache)
┌──────────────────────────────────┐
│  garimpei-proxy Worker           │
│  Workers Cache (TTL 5min)        │
│  Cache-Tag purge via API         │
└──────────────────────────────────┘
```

### Fronteiras de dados

| Store | Dono | Quem lê | Quem escreve |
|-------|------|---------|--------------|
| PostgreSQL | C# API | C# API | C# API |
| BigQuery | Collector | Analyzer | Collector |
| Cloud Tasks | Scheduler | Scheduler | Scheduler |
| Cache L2 (in-memory) | Cache Sidecar | C# API (Get) | Cache Sidecar (via Collector) |
| Cache L1 (edge) | Cloudflare Worker | Frontend (HIT) | Worker (PUT on MISS) |

Regras enforçadas por `mise run check:data-ownership`:
- Go services não acessam PostgreSQL
- C# API não acessa BigQuery
- Python analyzer não acessa PostgreSQL

---

## Credenciais para testes E2E

Os testes E2E usam credenciais armazenadas em `web/.env.e2e.local` (gitignored):

```bash
E2E_EMAIL=e2e@garimpei.app.br
E2E_PASSWORD=<senha do Firebase Auth>
E2E_BASE_URL=https://garimpei.app.br
```

Copiar de `web/.env.e2e` e preencher a senha.

---

## Git Workflow

- Branch única: `main`
- Push direto com pre-push hook que roda toda a suite (~25s)
- Se pre-push falha: corrigir o código, não bypassar (`--no-verify` proibido)
- Deploy automático via CI quando paths de backend/frontend mudam
- ADRs em `docs/decisoes/NNNN-nome.md`

---

## Links úteis

| Recurso | URL |
|---------|-----|
| Produção | https://garimpei.app.br |
| Docs site | https://garimpei-docs.pages.dev |
| Cloud Console | https://console.cloud.google.com/run?project=garimpo-500114 |
| Cloud Trace | https://console.cloud.google.com/traces/list?project=garimpo-500114 |
| Cloud Logging | https://console.cloud.google.com/logs?project=garimpo-500114 |
| GitHub | https://github.com/fmarquesfilho/garimpo |
