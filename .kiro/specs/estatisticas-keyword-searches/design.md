# Design Document

## Overview

Segmentar os dados analíticos do Analyzer por fonte (loja vs keyword) e refatorar a
página `/estatisticas` para dar visibilidade às buscas por keyword. O C# API não precisa
de mudanças (proxy transparente com `GetFromJsonAsync<object>`). O trabalho se concentra
no Analyzer Python (queries BigQuery com classificação por prefixo `loja-`) e no frontend
(novos painéis no dashboard).

## Components and Interfaces

### Analyzer Python — GET /estatisticas (atualizado)

**Response shape (backward-compatible + novos campos):**

```json
{
  "dias": 7,
  "total_amostras": 1250,
  "resumo": {
    "total_produtos": 1250,
    "preco_medio": 89.50,
    "comissao_media": 0.0845,
    "vendas_media": 42.3,
    "nota_media": 4.7,
    "preco_mediana": 75.00,
    "comissao_mediana": 0.08
  },
  "por_fonte": {
    "lojas": {
      "total_produtos": 800,
      "preco_medio": 95.20,
      "comissao_media": 0.09,
      "total_coletas": 15
    },
    "keywords": {
      "total_produtos": 450,
      "preco_medio": 79.30,
      "comissao_media": 0.075,
      "total_coletas": 8
    }
  }
}
```

**Lógica de classificação (SQL):**

```sql
CASE WHEN keyword LIKE 'loja-%' THEN 'loja' ELSE 'keyword' END AS fonte
```

Essa classificação é derivada do padrão que o Collector já usa ao gravar snapshots:
- Coleta de loja: `keyword = "loja-{shop_id}"` (ex: `loja-920292999`)
- Coleta por keyword: `keyword = "{termo}"` (ex: `serum vitamina c`)

### Analyzer Python — GET /evolucao (atualizado)

**Response shape (backward-compatible + novos campos):**

```json
{
  "dias": 30,
  "lojas": [
    {"busca_id": "loja-920292999", "produtos": 50, "variacao_media_pct": -0.03, "pontos": [...]}
  ],
  "keywords": [
    {"busca_id": "serum vitamina c", "produtos": 30, "variacao_media_pct": -0.05, "pontos": [...]}
  ],
  "resumo": {
    "total_lojas": 3,
    "total_produtos": 1200,
    "total_quedas": 45,
    "total_altas": 20,
    "preco_medio_global": 89.50,
    "variacao_media_global_pct": -0.012
  },
  "resumo_keywords": {
    "total_quedas": 12,
    "total_altas": 8
  },
  "total_lojas": 3
}
```

**Mudança:** O array `lojas_lista` existente é separado em dois arrays (`lojas` e
`keywords`) baseado no prefixo. O `resumo` continua global. Um novo `resumo_keywords`
contém quedas/altas apenas de keyword searches.

### C# API — sem mudanças

Os endpoints `/api/estatisticas` e `/api/lojas/evolucao` usam `GetFromJsonAsync<object>`
e retornam o JSON do Analyzer sem transformação. Campos novos passam automaticamente.

### Frontend — /estatisticas (refatorado)

**Layout proposto:**

```
┌────────────────────────────────────────────────────────────────────┐
│ 📊 Dashboard                                         [7 dias ▾]   │
├────────────────────────────────────────────────────────────────────┤
│ ┌────────┐ ┌────────────────┐ ┌──────────────┐ ┌───────────────┐ │
│ │ Lojas  │ │ Produtos (tot) │ │ Publicações  │ │ Taxa sucesso  │ │
│ │   3    │ │     1250       │ │     12       │ │     100%      │ │
│ └────────┘ └────────────────┘ └──────────────┘ └───────────────┘ │
│ ┌──────────────────┐ ┌──────────────────┐ ┌─────────────────────┐│
│ │ Buscas keyword   │ │ Produtos (kw)    │ │ ↓ Quedas  ↑ Altas  ││
│ │       8          │ │      450         │ │   45        20      ││
│ │                  │ │                  │ │ 33 lojas · 12 kw    ││
│ └──────────────────┘ └──────────────────┘ └─────────────────────┘│
├────────────────────────────────────────────────────────────────────┤
│ ┌──────────────────────────┐ ┌──────────────────────────────────┐ │
│ │ 🏆 Mais publicados       │ │ 📈 Preço médio (lojas)          │ │
│ │   1. Sérum X — 3×        │ │   ──── loja A ──── (sparkline)  │ │
│ │   2. Perfume Y — 2×      │ │   ──── loja B ──── (sparkline)  │ │
│ └──────────────────────────┘ └──────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────────────────┐
│ │ 📈 Preço médio (keywords)                                       │
│ │   ──── "serum vitamina c" ──── (-5.2%)                          │
│ │   ──── "protetor solar"  ──── (+1.1%)                           │
│ │   ──── "retinol"         ──── (-2.8%)                           │
│ └──────────────────────────────────────────────────────────────────┘
└────────────────────────────────────────────────────────────────────┘
```

## Data Models

### BigQuery — campo `keyword` nos snapshots (existente, sem mudança)

| Fonte | Valor do campo `keyword` | Exemplo |
|-------|--------------------------|---------|
| Coleta de loja | `loja-{shop_id}` | `loja-920292999` |
| Coleta por keyword | `{termo de busca}` | `serum vitamina c` |

A classificação é feita em query-time pelo Analyzer (não requer coluna extra).

### Analyzer response — `por_fonte` (novo)

```python
@dataclass
class FonteStats:
    total_produtos: int
    preco_medio: float
    comissao_media: float
    total_coletas: int  # COUNT(DISTINCT keyword) — quantas keywords/lojas distintas

@dataclass
class EstatisticasResponse:
    dias: int
    total_amostras: int
    resumo: dict           # existente (global)
    por_fonte: dict        # novo: {"lojas": FonteStats, "keywords": FonteStats}
```

### Analyzer response — `resumo_keywords` (novo no /evolucao)

```python
@dataclass
class ResumoKeywords:
    total_quedas: int   # produtos com variação < -1% apenas em keyword searches
    total_altas: int    # produtos com variação > +1% apenas em keyword searches
```

## Error Handling

| Cenário | Comportamento |
|---------|--------------|
| Analyzer offline | C# API retorna fallback vazio (já implementado). Frontend mostra "0" nos MetricCards |
| BigQuery sem dados para keywords | Analyzer retorna `por_fonte.keywords` com zeros. Frontend mostra "0" com estilo muted |
| Apenas 1 data point para uma keyword | Frontend mostra Badge com preço em vez de MiniChart |
| Campo `por_fonte` ausente (versão antiga do Analyzer) | Frontend usa fallback: `por_fonte?.keywords?.total_coletas ?? 0` |

## Testing Strategy

| Camada | O que testar | Como |
|--------|-------------|------|
| Analyzer | Classificação loja vs keyword | `ruff check` + syntax check no CI (sem testes unitários Python hoje) |
| Frontend | Renderização condicional dos novos painéis | Vitest unit test para computed values (`$derived`) |
| Integração | Fluxo completo | E2E local via `mise run test:e2e:estatisticas` (manual, não CI) |

## Correctness Properties

1. **Invariante de soma:** `por_fonte.lojas.total_produtos + por_fonte.keywords.total_produtos == total_amostras`
2. **Classificação determinística:** O prefixo `loja-` é set pelo Collector Go no momento da gravação e nunca muda
3. **Backward compatibility:** O campo `resumo` continua com os totais globais (lojas + keywords). Frontends antigos que não leiam `por_fonte` continuam funcionando
4. **Data ownership preservado:** Analyzer lê BigQuery → C# proxy → Frontend exibe. Nenhuma fronteira é cruzada

## Summary of Changes

| Arquivo | Tipo | Descrição |
|---------|------|-----------|
| `services/analyzer/routes/estatisticas.py` | Ajuste | Adicionar `por_fonte` com segmentação por prefixo `loja-` |
| `services/analyzer/routes/evolucao.py` | Ajuste | Separar `lojas_lista` em `lojas` + `keywords`; adicionar `resumo_keywords` |
| `web/src/routes/estatisticas/+page.svelte` | Refactor | Novos MetricCards (keyword), painel "Preço médio (keywords)", breakdown quedas/altas |
| `docs/08-fluxos-sequencia.md` | Documentação | Atualizar Fluxo 9 com novos campos e classificação por fonte |
