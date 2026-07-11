# Implementation Plan: Analyzer Dashboard v2

## Overview

Four new Python Analyzer endpoints and a frontend dashboard restructure. The Analyzer adds `/coletas/saude`, `/oportunidades/agora`, `/conversoes/resumo`, and `/alertas/eficacia`. The C# API proxies them. The frontend `/estatisticas` page is reorganized into 3 question-oriented sections with existing charts moved to collapsible panels.

Implementation order: Analyzer routes first (backend independently testable), then C# proxies, then frontend API client functions, then dashboard rewrite.

## Tasks

- [ ] 1. Implement Collection Health endpoint
  - [ ] 1.1 Create `services/analyzer/routes/saude.py` with `GET /coletas/saude`
    - Query BigQuery `snapshots` table for: MAX(coletado_em), COUNT(DISTINCT keyword) in last 24h, distinct keywords from last 7 days (expected count)
    - Compute `minutos_desde_ultima` from current time minus last collection
    - Set status: "sem_dados" if no rows, "atrasado" if >360 min, "ok" otherwise
    - Compute `keywords_atrasadas`: keywords seen in last 7 days but NOT in last 24h
    - Wrap in try/except: table not found → return structured empty response
    - Use `ScalarQueryParameter` for all values
    - Return JSON matching the schema defined in design
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 6.1, 6.3_

  - [ ] 1.2 Register saude router in `services/analyzer/main.py`
    - Import `from routes import saude` and add `app.include_router(saude.router)`
    - Do NOT modify any existing router registrations
    - _Requirements: 7.2_

- [ ] 2. Implement Opportunities endpoint
  - [ ] 2.1 Create `services/analyzer/routes/oportunidades.py` with `GET /oportunidades/agora`
    - Accept `dias` query param (int, 1-30, default 7)
    - Query 1 — Active price drops: compare FIRST_VALUE vs LAST_VALUE of price within window, WHERE variacao <= -0.10, ORDER BY variacao ASC, LIMIT 10
    - Query 2 — New products: WHERE aparicoes = 1 AND coletado_em >= 48h ago, LIMIT 10
    - Query 3 — High-value unpublished: commission > P75 AND sales > median, LEFT JOIN publicacoes to exclude published, LIMIT 5
    - If `publicacoes` table not found: skip JOIN, set `filtro_publicacoes: false`
    - All queries use `ScalarQueryParameter` for `dias`
    - Return JSON with `quedas`, `novos`, `alto_valor` lists + totals
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 6.1, 6.3_

  - [ ] 2.2 Register oportunidades router in `services/analyzer/main.py`
    - Import and register without modifying existing routers
    - _Requirements: 7.2_

- [ ] 3. Implement Revenue Summary endpoint
  - [ ] 3.1 Create `services/analyzer/routes/resumo_conversoes.py` with `GET /conversoes/resumo`
    - Accept `dias` query param (int, 1-180, default 30)
    - Query BigQuery `conversoes` table: SUM(comissao) as comissao_total, COUNT(*) as conversoes, COUNT(DISTINCT produto_id), GROUP BY canal
    - Compute `melhor_canal` as canal with highest comissao
    - If table not found: return `status: "sem_dados"` with zeroed metrics
    - Return JSON matching design schema
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.3_

  - [ ] 3.2 Register resumo_conversoes router in `services/analyzer/main.py`
    - _Requirements: 7.2_

- [ ] 4. Implement Alert Efficacy endpoint
  - [ ] 4.1 Create `services/analyzer/routes/eficacia.py` with `GET /alertas/eficacia`
    - Accept `dias` query param (int, 1-180, default 30)
    - Query 1 — Price drops detected: COUNT products with variacao <= -0.15 in snapshots within window
    - Query 2 — Alerts sent: COUNT publications with estrategia containing "alerta" or triggered by drops (heuristic: publicacoes created within 1h of a drop detection)
    - Query 3 — Conversions attributed: JOIN conversoes.produto_id with dropped products
    - Compute `taxa_deteccao` = alertas_enviados / quedas_detectadas * 100
    - Compute `taxa_conversao` = conversoes_atribuidas / alertas_enviados * 100
    - Compute `melhor_keyword` = keyword with most conversions
    - If no drops: return all zeros with taxa_deteccao as null
    - If conversoes table missing: return drop/alert counts, set `conversoes_disponiveis: false`
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 6.1, 6.3_

  - [ ] 4.2 Register eficacia router in `services/analyzer/main.py`
    - _Requirements: 7.2_

- [ ] 5. Checkpoint - Verify Analyzer builds and existing tests pass
  - Run `uvx ruff check services/analyzer/` — no lint errors
  - Verify `python -c "from services.analyzer.main import app"` imports without error (syntax check)
  - Verify existing endpoints still respond (manual curl or `mise run test:e2e-analyzer local`)

- [ ] 6. Add C# API proxies for new Analyzer endpoints
  - [ ] 6.1 Create `src/Garimpei.Api/Endpoints/AnalyzerProxyEndpoints.cs`
    - Map `GET /api/coletas/saude` → proxy to `{analyzerUrl}/coletas/saude`
    - Map `GET /api/oportunidades/agora?dias=N` → proxy to `{analyzerUrl}/oportunidades/agora?dias=N`
    - Map `GET /api/conversoes/resumo?dias=N` → proxy to `{analyzerUrl}/conversoes/resumo?dias=N`
    - Map `GET /api/alertas/eficacia?dias=N` → proxy to `{analyzerUrl}/alertas/eficacia?dias=N`
    - All require authorization (`.RequireAuthorization()`)
    - On analyzer error: return empty-state JSON (same pattern as existing `LojasEndpoints.cs` proxy)
    - _Requirements: 5.2, 5.3, 5.4, 6.1_

  - [ ] 6.2 Register proxy endpoints in `src/Garimpei.Api/Program.cs`
    - Add `app.MapAnalyzerProxyEndpoints();` after existing endpoint registrations
    - _Requirements: 7.2_

- [ ] 7. Checkpoint - Verify C# builds
  - Run `dotnet build --nologo src/Garimpei.Api/Garimpei.Api.csproj` — 0 errors, 0 warnings
  - Run `dotnet test --nologo src/Garimpei.sln` — all tests pass

- [ ] 8. Add frontend API client functions
  - [ ] 8.1 Add 4 new functions to `web/src/lib/api.js`
    - `buscarSaudeColetas()` → `pegar('/api/coletas/saude')`
    - `buscarOportunidadesAgora({ dias = 7 } = {})` → `pegar('/api/oportunidades/agora?dias=${dias}')`
    - `buscarResumoConversoes({ dias = 30 } = {})` → `pegar('/api/conversoes/resumo?dias=${dias}')`
    - `buscarEficaciaAlertas({ dias = 30 } = {})` → `pegar('/api/alertas/eficacia?dias=${dias}')`
    - _Requirements: 5.2, 5.3, 5.4_

- [ ] 9. Restructure Dashboard frontend
  - [ ] 9.1 Rewrite `web/src/routes/estatisticas/+page.svelte`
    - Load all data in `onMount` via `Promise.all([buscarSaudeColetas(), buscarOportunidadesAgora({dias}), buscarResumoConversoes({dias}), buscarEficaciaAlertas({dias}), buscarEstatisticas({dias}), buscarEvolucaoLojas({dias})])`
    - Section 1 "🔄 Saúde": Badge with status color (green/yellow/gray), relative time since last collection, "X/Y coletas 24h", list of stale keywords (if any)
    - Section 2 "💰 Oportunidades": RankList of top 5 drops (name + variacao% + price), Badge with new products count, Badge with high-value count
    - Section 3 "📊 Performance": MetricCard for comissao_total (BRL), MetricCard for conversoes count, MetricCard for melhor_canal, MetricCard for taxa_deteccao %
    - Collapsible panel "📈 Evolução de preços" with existing MiniChart code (lojas + keywords)
    - Collapsible panel "📋 Histórico de coletas" with existing metrics cards (moved from current row 1 and row 2)
    - Each section independently handles errors (try/catch per section or per-section error state)
    - Use existing components only: MetricCard, DashPanel, MiniChart, RankList, Badge, Alert, Select
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

- [ ] 10. Final checkpoint - Full validation
  - Run `cd web && bun run check` — svelte-check passes
  - Run `cd web && npx vitest run` — unit tests pass
  - Run `uvx ruff check services/analyzer/` — lint passes
  - Run `dotnet test --nologo src/Garimpei.sln` — C# tests pass
  - Manual: open `/estatisticas` locally and verify 3 sections render (empty-state OK)
  - Verify existing `/coletas` page still works unchanged

## Notes

- The Analyzer endpoints (tasks 1-4) are fully independent of each other and can be implemented in any order
- C# proxies (task 6) depend on endpoints existing but can be written before deploy (they gracefully handle analyzer being down)
- Frontend (tasks 8-9) can be developed in parallel with proxy placeholders returning empty JSON
- The `publicacoes` table in BigQuery may not exist for all users — the `/oportunidades/agora` query must handle this gracefully
- BigQuery query timeout uses `QueryJobConfig(timeout_ms=5000)` from the python client
- No new Python packages required (FastAPI + google-cloud-bigquery already installed)
- Steering rules respected: no E2E tests added to CI, only lint and build checks

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "2.1", "3.1", "4.1"] },
    { "id": 1, "tasks": ["1.2", "2.2", "3.2", "4.2"] },
    { "id": 2, "tasks": ["6.1"] },
    { "id": 3, "tasks": ["6.2"] },
    { "id": 4, "tasks": ["8.1"] },
    { "id": 5, "tasks": ["9.1"] }
  ]
}
```
