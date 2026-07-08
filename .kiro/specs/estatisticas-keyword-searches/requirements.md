# Requirements Document

## Introduction

Refactor da página `/estatisticas` e do Analyzer Python para diferenciar e exibir dados provenientes de buscas agendadas por keyword (sem loja) separadamente dos dados de monitoramento de lojas. Atualmente o dashboard trata todos os snapshots do BigQuery como um bloco único, sem distinção de fonte. O resultado é que buscas por keyword não têm visibilidade no dashboard — o usuário não sabe quantas coletas por keyword foram feitas, qual a evolução de preço dessas buscas, ou qual a performance comparativa entre lojas monitoradas e buscas avulsas.

O refactor inclui:
- Novo endpoint no Analyzer que retorna estatísticas segmentadas por fonte (loja vs keyword)
- Atualização da página `/estatisticas` para exibir seção dedicada a buscas por keyword
- Documentação do fluxo completo no diagrama de sequência (docs/08-fluxos-sequencia.md)

## Glossary

- **Analyzer**: Serviço Python (FastAPI) que lê BigQuery e retorna dados analíticos via REST. Dono exclusivo das queries analíticas sobre snapshots.
- **Snapshot**: Registro no BigQuery contendo dados de um produto coletado (produto_id, nome, preco, comissao, keyword, coletado_em). O campo `keyword` identifica a fonte da coleta.
- **Busca_Loja**: Busca agendada com `shop_ids` preenchido. O campo `keyword` nos snapshots segue o padrão `loja-{shop_id}` (ex: `loja-920292999`).
- **Busca_Keyword**: Busca agendada sem `shop_ids`, usando apenas termos de busca. O campo `keyword` nos snapshots contém a keyword em si (ex: `serum vitamina c`).
- **C#_API**: Serviço C# que funciona como proxy transparente para o Analyzer e dono do PostgreSQL (publicações, buscas, destinos).
- **Frontend**: Aplicação SvelteKit que faz chamadas paralelas ao C# API para montar o dashboard.
- **Fonte**: Classificação da origem de um snapshot — `loja` (quando keyword começa com `loja-`) ou `keyword` (caso contrário).
- **DashPanel**: Componente application do dashboard que exibe um painel com título e conteúdo (gráficos, listas). Usa Card do shadcn-svelte internamente.
- **MetricCard**: Componente application que exibe uma métrica numérica com label. Usa Card + variantes Tailwind.
- **MiniChart**: Componente application que renderiza uma série temporal de pontos como sparkline SVG.
- **UI_Library**: Stack de componentes: shadcn-svelte pattern (primitivos), Bits UI v2 (compostos), Tailwind CSS v4 (styling), Svelte 5 runes ($state, $derived). Componentes vivem em `$lib/components/ui/`.

## Requirements

### Requirement 1: Segmentação de estatísticas por fonte no Analyzer

**User Story:** As a user, I want the Analyzer to return statistics segmented by source (store vs keyword search), so that I can understand the performance of each collection type separately.

#### Acceptance Criteria

1. WHEN the Frontend requests statistics with a `dias` parameter, THE Analyzer SHALL return a response containing `por_fonte` with two sub-objects: `lojas` and `keywords`, each with `total_produtos`, `preco_medio`, `comissao_media`, and `total_coletas`.
2. THE Analyzer SHALL classify a snapshot as Busca_Loja when the `keyword` field starts with the prefix `loja-`, and as Busca_Keyword otherwise.
3. THE Analyzer SHALL maintain backward compatibility by continuing to return the existing top-level fields (`total_amostras`, `resumo`) alongside the new `por_fonte` segmentation.
4. WHEN no snapshots exist for a given fonte within the time window, THE Analyzer SHALL return zeroed values for that fonte (`total_produtos: 0`, `preco_medio: 0`, `comissao_media: 0`, `total_coletas: 0`).

### Requirement 2: Evolução temporal de buscas por keyword

**User Story:** As a user, I want to see price evolution charts for my keyword searches (not just monitored stores), so that I can track price trends for products found via keyword.

#### Acceptance Criteria

1. WHEN the Frontend requests evolution data, THE Analyzer SHALL return two top-level arrays: `lojas` (existing behavior) and `keywords` (new), each containing series with `busca_id`, `pontos[]`, `variacao_media_pct`, and `produtos`.
2. THE Analyzer SHALL use the Fonte classification (prefix `loja-`) to separate entries into the `lojas` array or the `keywords` array.
3. WHEN a keyword search has fewer than 2 data points in the time window, THE Analyzer SHALL still include the keyword in the `keywords` array with the available data points.
4. THE Analyzer SHALL include a `resumo_keywords` object with `total_quedas` and `total_altas` computed exclusively from Busca_Keyword snapshots, separate from the existing `resumo` (which remains scoped to all sources combined).

### Requirement 3: Métricas de keyword no dashboard

**User Story:** As a user, I want to see keyword search metrics prominently on the dashboard, so that I have visibility into how many keyword collections ran and their results.

#### Acceptance Criteria

1. THE Frontend SHALL display a MetricCard labeled "Buscas por keyword" showing the count of distinct Busca_Keyword entries from the `por_fonte.keywords.total_coletas` field.
2. THE Frontend SHALL display a MetricCard labeled "Produtos (keywords)" showing `por_fonte.keywords.total_produtos`.
3. THE Frontend SHALL display the existing "Lojas" MetricCard using the count from `buscas` (PostgreSQL) as it does today, without change.
4. WHEN `por_fonte.keywords.total_coletas` is zero, THE Frontend SHALL display the keyword MetricCards with value "0" and muted visual styling (Tailwind `text-tinta-suave` variant).
5. THE Frontend SHALL use existing UI_Library components (MetricCard, DashPanel, MiniChart from `$lib/components/ui/`) and Tailwind CSS v4 utilities for all new visual elements — no custom CSS.

### Requirement 4: Painel de evolução de preço para keywords

**User Story:** As a user, I want to see a price evolution chart for keyword searches in the dashboard, so that I can monitor price trends alongside store monitoring data.

#### Acceptance Criteria

1. THE Frontend SHALL render a DashPanel titled "📈 Preço médio (keywords)" that displays MiniChart components for keyword searches, positioned below the existing store price evolution panel. The DashPanel SHALL use the existing Card-based component with Tailwind styling.
2. THE Frontend SHALL display up to 3 keyword entries from the `keywords` array, sorted by highest absolute `variacao_media_pct`.
3. WHEN the `keywords` array is empty, THE Frontend SHALL display an inline message "Sem dados de buscas por keyword no período." inside the DashPanel using `text-sm italic text-tinta-suave` styling.
4. WHEN a keyword entry has only 1 data point, THE Frontend SHALL display the single-point value as a Badge (shadcn-svelte) with the price, without rendering a MiniChart line.

### Requirement 5: Resumo de quedas e altas por fonte

**User Story:** As a user, I want to see price drops and rises separated by source (stores vs keywords), so that I can understand where the most price movement is happening.

#### Acceptance Criteria

1. THE Frontend SHALL display the existing "↓ Quedas" and "↑ Altas" MetricCards using values from the combined `resumo` (all sources), maintaining current behavior.
2. THE Frontend SHALL display additional secondary indicators below the main MetricCards showing the breakdown: "{N} de lojas · {M} de keywords" using data from `resumo` and `resumo_keywords`.
3. WHEN `resumo_keywords` is absent or has zero values, THE Frontend SHALL omit the secondary breakdown indicator and display only the combined total.

### Requirement 6: Documentação do fluxo no diagrama de sequência

**User Story:** As a developer, I want the sequence diagram documentation to reflect the new segmented statistics flow, so that future developers understand the data flow and ownership model.

#### Acceptance Criteria

1. THE Documentation SHALL contain an updated Fluxo 9 (Dashboard) in `docs/08-fluxos-sequencia.md` showing the new `por_fonte` and `keywords` fields in the Analyzer response.
2. THE Documentation SHALL describe the Fonte classification logic (prefix `loja-`) in the technical details section of Fluxo 9.
3. THE Documentation SHALL maintain consistency with the existing data ownership model (Analyzer reads BigQuery, C# API proxies without transformation, Frontend composes the view).

### Requirement 7: Proxy transparente no C# API

**User Story:** As a developer, I want the C# API to proxy the new Analyzer response fields without transformation, so that the data ownership model is preserved.

#### Acceptance Criteria

1. THE C#_API SHALL forward the complete Analyzer response (including new `por_fonte`, `keywords`, and `resumo_keywords` fields) to the Frontend without modification.
2. IF the Analyzer returns an error or is unavailable, THEN THE C#_API SHALL return a graceful fallback response with empty/zeroed statistics, consistent with existing error handling behavior.
3. THE C#_API SHALL require no schema changes or new endpoints for this feature — the existing `/api/estatisticas` and `/api/lojas/evolucao` proxy routes are sufficient.
