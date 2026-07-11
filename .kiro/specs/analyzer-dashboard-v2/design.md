# Technical Design — Analyzer Dashboard v2

## Overview

Quatro novos endpoints no Analyzer Python (FastAPI) e reestruturação do frontend `/estatisticas` em 3 seções orientadas por perguntas do usuário. Os endpoints consultam BigQuery (snapshots, publicacoes, conversoes) com queries parametrizadas e timeout de 5s. O frontend usa componentes existentes (MetricCard, DashPanel, RankList, Badge) sem novas dependências.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Frontend (Svelte)                           │
│                                                                  │
│  ┌───────────────┐ ┌──────────────────┐ ┌───────────────────┐  │
│  │  🔄 Saúde     │ │ 💰 Oportunidades │ │  📊 Performance   │  │
│  │  /coletas/    │ │ /oportunidades/  │ │  /conversoes/     │  │
│  │   saude       │ │  agora           │ │   resumo          │  │
│  └───────┬───────┘ └────────┬─────────┘ └─────────┬─────────┘  │
│          │                   │                     │             │
│  ┌───────┴───────────────────┴─────────────────────┴───────────┐│
│  │              Painéis colapsáveis (evolução, coletas)          ││
│  │              /evolucao  /estatisticas  /coletas (existentes)  ││
│  └──────────────────────────────────────────────────────────────┘│
└──────────────────────────────────┬───────────────────────────────┘
                                   │ HTTP (via C# proxy)
                                   ▼
┌──────────────────────────────────────────────────────────────────┐
│                    Analyzer (Python FastAPI)                       │
│                                                                   │
│  Existentes (inalterados):                                        │
│    /estatisticas  /coletas  /evolucao  /quedas  /novidades       │
│    /conversoes  /cupons                                           │
│                                                                   │
│  Novos:                                                           │
│    /coletas/saude        → saúde das coletas                     │
│    /oportunidades/agora  → drops + novos não-publicados          │
│    /conversoes/resumo    → receita + canais                      │
│    /alertas/eficacia     → drops → alertas → conversões          │
│                                                                   │
└──────────────────────────────┬────────────────────────────────────┘
                               │ BigQuery queries
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                       BigQuery                                    │
│  snapshots | publicacoes | conversoes | coupon_snapshots          │
└──────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. New Analyzer Routes

Quatro novos arquivos em `services/analyzer/routes/`:

| Arquivo | Endpoint | Query params |
|---------|----------|--------------|
| `saude.py` | `GET /coletas/saude` | — |
| `oportunidades.py` | `GET /oportunidades/agora` | `dias` (1-30, default 7) |
| `resumo_conversoes.py` | `GET /conversoes/resumo` | `dias` (1-180, default 30) |
| `eficacia.py` | `GET /alertas/eficacia` | `dias` (1-180, default 30) |

Todos registrados no `main.py` via `app.include_router(...)`, sem modificar routers existentes.

### 2. C# API Proxies

Novos endpoints em `LojasEndpoints.cs` ou novo `AnalyzerProxyEndpoints.cs`:

```csharp
app.MapGet("/api/coletas/saude", ...) → analyzer/coletas/saude
app.MapGet("/api/oportunidades/agora", ...) → analyzer/oportunidades/agora
app.MapGet("/api/conversoes/resumo", ...) → analyzer/conversoes/resumo
app.MapGet("/api/alertas/eficacia", ...) → analyzer/alertas/eficacia
```

Padrão idêntico aos proxies existentes (`/api/coletas` → analyzer, `/api/estatisticas` → analyzer).

### 3. Frontend Dashboard Restructure

O arquivo `web/src/routes/estatisticas/+page.svelte` é reestruturado:

```
┌─ Saúde ─────────────────────────────────────────────────┐
│ Badge[status] │ "Última coleta: 2h atrás" │ "6/9 24h"  │
│ Keywords atrasadas: [serum, mascara] (se houver)         │
└─────────────────────────────────────────────────────────┘

┌─ Oportunidades ─────────────────────────────────────────┐
│ RankList: Top 5 quedas (nome, -23%, R$49.90)            │
│ Badge: "3 novos hoje" │ Badge: "2 alto-valor"           │
└─────────────────────────────────────────────────────────┘

┌─ Performance ───────────────────────────────────────────┐
│ MetricCard: R$127,40 │ MetricCard: 12 conversões        │
│ MetricCard: Telegram (melhor) │ MetricCard: 67% detecção│
└─────────────────────────────────────────────────────────┘

▸ Evolução de preços (colapsável — charts existentes)
▸ Histórico de coletas (colapsável — tabela existente)
```

### 4. API Client Functions (frontend)

Novas funções em `web/src/lib/api.js`:

```javascript
export function buscarSaudeColetas() { return pegar('/api/coletas/saude'); }
export function buscarOportunidadesAgora({ dias = 7 } = {}) { return pegar(`/api/oportunidades/agora?dias=${dias}`); }
export function buscarResumoConversoes({ dias = 30 } = {}) { return pegar(`/api/conversoes/resumo?dias=${dias}`); }
export function buscarEficaciaAlertas({ dias = 30 } = {}) { return pegar(`/api/alertas/eficacia?dias=${dias}`); }
```

## Data Models

### `/coletas/saude` Response

```json
{
  "ultima_coleta": "2026-07-10T18:00:00Z",
  "minutos_desde_ultima": 120,
  "status": "ok | atrasado | sem_dados",
  "coletas_24h": 6,
  "coletas_esperadas_24h": 9,
  "keywords_atrasadas": ["perfume", "kit coreano"]
}
```

### `/oportunidades/agora` Response

```json
{
  "dias": 7,
  "quedas": [
    {
      "produto_id": "shopee-920292999-100001",
      "nome": "Sérum Vitamina C 30ml",
      "preco_anterior": 79.90,
      "preco_atual": 49.90,
      "variacao": -0.375,
      "loja": "Glory of Seoul",
      "imagem": "https://...",
      "link": "https://..."
    }
  ],
  "novos": [
    {
      "produto_id": "shopee-592884015-300004",
      "nome": "COSRX Niacinamide Serum",
      "preco": 42.90,
      "comissao": 0.11,
      "loja": "COSRX Official",
      "detectado_em": "2026-07-10T08:00:00Z"
    }
  ],
  "alto_valor": [
    {
      "produto_id": "shopee-592884015-300001",
      "nome": "COSRX Snail Mucin",
      "preco": 89.90,
      "comissao": 0.11,
      "vendas": 5600,
      "loja": "COSRX Official"
    }
  ],
  "total_quedas": 12,
  "total_novos": 3,
  "total_alto_valor": 2,
  "filtro_publicacoes": true
}
```

### `/conversoes/resumo` Response

```json
{
  "dias": 30,
  "comissao_total": 127.40,
  "conversoes": 12,
  "produtos_distintos": 8,
  "por_canal": [
    { "canal": "telegram", "comissao": 95.20, "conversoes": 9 },
    { "canal": "whatsapp", "comissao": 32.20, "conversoes": 3 }
  ],
  "melhor_canal": "telegram",
  "status": "ok | sem_dados"
}
```

### `/alertas/eficacia` Response

```json
{
  "dias": 30,
  "quedas_detectadas": 45,
  "alertas_enviados": 30,
  "conversoes_atribuidas": 8,
  "taxa_deteccao": 66.7,
  "taxa_conversao": 26.7,
  "melhor_keyword": "serum",
  "conversoes_disponiveis": true
}
```

## Error Handling

### BigQuery table not found

Cada endpoint wraps a query em try/except. Se a tabela não existir (ex: `conversoes` em conta nova):
- Retorna response estruturado com status `"sem_dados"` ou campo booleano indicando ausência
- Nunca retorna HTTP 500

### Query timeout (>5s)

Usa `job_config.timeout = 5.0` no BigQuery client. Se expirar:
- Retorna dados parciais (se houver) com campo `timeout: true`
- Loga warning para investigação

### Empty datasets (novo usuário)

Todos os endpoints retornam respostas válidas com listas vazias e contadores zerados. O frontend exibe empty-state contextual por seção.

## Correctness Properties

### Property 1: Endpoints existentes inalterados

Nenhum router existente é modificado. Novos routers são adicionados via `app.include_router()` separado. Os 7 endpoints atuais mantêm schema e comportamento idênticos.

**Validates: Requirements 7.1, 7.2**

### Property 2: Queries parametrizadas

Todos os valores vindos de query params (`dias`) usam `ScalarQueryParameter` do BigQuery client. Nenhuma interpolação direta de input do usuário em SQL strings.

**Validates: Requirements 6.3**

### Property 3: Independência entre seções do dashboard

Cada seção do frontend faz sua própria chamada de API em paralelo (`Promise.all`). Se uma falha, as outras exibem normalmente. Erro é contido na seção afetada.

**Validates: Requirements 5.7**

### Property 4: Sem novas dependências Python

Os 4 novos endpoints usam apenas `fastapi`, `google-cloud-bigquery`, e módulos stdlib. Nenhum pip install adicional.

**Validates: Requirements 6.4**

## File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `services/analyzer/routes/saude.py` | Create | `/coletas/saude` endpoint |
| `services/analyzer/routes/oportunidades.py` | Create | `/oportunidades/agora` endpoint |
| `services/analyzer/routes/resumo_conversoes.py` | Create | `/conversoes/resumo` endpoint |
| `services/analyzer/routes/eficacia.py` | Create | `/alertas/eficacia` endpoint |
| `services/analyzer/main.py` | Modify | Register 4 new routers |
| `src/Garimpei.Api/Endpoints/AnalyzerProxyEndpoints.cs` | Create | 4 proxy routes to analyzer |
| `src/Garimpei.Api/Program.cs` | Modify | Register `MapAnalyzerProxyEndpoints()` |
| `web/src/lib/api.js` | Modify | Add 4 new API client functions |
| `web/src/routes/estatisticas/+page.svelte` | Rewrite | 3-section layout + collapsible panels |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Complex BigQuery JOINs in `/oportunidades/agora` (snapshots LEFT JOIN publicacoes) | Use subquery with LIMIT before JOIN. BigQuery handles this efficiently for append-only tables. |
| Dashboard loads 4 endpoints in parallel — N+1 latency | `Promise.all` ensures parallel fetch. Total load time = max(individual latencies) ≈ 2-3s. |
| `conversoes` table may not exist for most users | Try/except wrapper returns `status: "sem_dados"`. Frontend shows "Configure suas credenciais Shopee para ver conversões". |
| Frontend rewrite risk (breaking existing charts) | Charts move to collapsible panel — same component, same data, just repositioned. Covered by E2E tests. |

## Testing Strategy

- **Unit (Python)**: Mock `bq_client.query()` → test each route returns correct schema for empty/populated data
- **Unit (Python)**: Test graceful degradation (table not found → structured response)
- **Unit (Python)**: Test parameter validation (dias out of range → FastAPI validation error)
- **Unit (Svelte)**: Test dashboard sections render independently (one error doesn't break others)
- **Build**: `ruff check services/analyzer/` passes
- **E2E**: `mise run test:e2e-services` validates new endpoints respond in prod after deploy
- **Manual**: Visual inspection of dashboard with 0 data, 7 days data, 30 days data
