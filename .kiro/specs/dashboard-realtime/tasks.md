# Implementation Plan: Dashboard Realtime

## Overview

Smart polling with change detection, selective section refresh, animated transitions, and a freshness indicator. Builds on top of the analyzer-dashboard-v2 endpoints. Zero new infrastructure — uses existing Analyzer + C# proxy pattern.

## Tasks

- [ ] 1. Create Change Detection endpoint
  - [ ] 1.1 Create `services/analyzer/routes/changes.py` with `GET /dashboard/changes`
    - Query BigQuery for MAX(coletado_em) from snapshots → `saude_updated_at`
    - Query MAX(coletado_em) OR MAX of last publication → `oportunidades_updated_at`
    - Query MAX(convertido_em) from conversoes → `performance_updated_at`
    - All queries scoped by owner_uid (when available) or global if no multi-tenant filter yet
    - On table not found: return null for that section's timestamp
    - Response must complete within 2 seconds
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.7_

  - [ ] 1.2 Register changes router in `services/analyzer/main.py`
    - _Requirements: 1.1_

  - [ ] 1.3 Add C# proxy `GET /api/dashboard/changes` in `AnalyzerProxyEndpoints.cs`
    - Require authorization, pass through to analyzer
    - Set `Cache-Control: no-store` header on response
    - On analyzer error: return all nulls (graceful degradation)
    - _Requirements: 1.6, 1.8, 8.1_

- [ ] 2. Create PollingTimer module
  - [ ] 2.1 Create `web/src/lib/polling.svelte.js`
    - Export `createPollingTimer({ interval, onTick, immediate })` factory
    - Internal: `$state` for countdown, paused, consecutiveErrors
    - `start()`: sets interval, begins countdown decrement (every 1s)
    - `stop()`: clears interval and countdown
    - `pause()` / `resume()`: for tab visibility
    - On tick: call `onTick()`, reset countdown
    - Visibility API: `visibilitychange` listener auto-pauses/resumes
    - If hidden >5 min on resume: signal full refetch needed
    - Read interval from `import.meta.env.VITE_POLL_INTERVAL_MS` (default 30000)
    - Clamp to 10000-120000ms range
    - Accept `?intervalo=` URL param override for debugging
    - _Requirements: 2.1, 2.2, 2.5, 3.1, 3.2, 3.3, 9.1, 9.2, 9.3_

  - [ ] 2.2 Implement error backoff logic in PollingTimer
    - Track `consecutiveErrors` counter
    - After 3 consecutive errors: double the interval
    - On first success after backoff: reset interval to default
    - _Requirements: 7.3, 7.4_

- [ ] 3. Create UI components
  - [ ] 3.1 Create `web/src/lib/components/ui/FreshnessBar.svelte`
    - Props: `lastUpdate` (Date), `countdown` (number, seconds), `status` ('live' | 'paused' | 'offline')
    - Display: StatusIndicator (reuse existing) + relative time + countdown text
    - Green pulsing dot when live and data <60s old
    - Gray dot + "pausado" when paused
    - Yellow dot + "sem conexão" when offline
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ] 3.2 Create `web/src/lib/components/ui/AnimatedMetric.svelte`
    - Props: `value` (number), `format` (fn), `duration` (ms, default 600)
    - Use Svelte `tweened` store with `cubicOut` easing
    - On value change: tween from old to new over duration
    - Display formatted tweened value
    - _Requirements: 5.1, 5.5_

  - [ ] 3.3 Export new components in `web/src/lib/components/ui/index.js`
    - Add FreshnessBar and AnimatedMetric exports
    - _Requirements: 4.5, 5.5_

- [ ] 4. Add API client function
  - [ ] 4.1 Add `buscarDashboardChanges()` to `web/src/lib/api.js`
    - `return pegar('/api/dashboard/changes')`
    - _Requirements: 2.2_

- [ ] 5. Integrate polling into Dashboard page
  - [ ] 5.1 Add polling logic to `web/src/routes/estatisticas/+page.svelte`
    - Import `createPollingTimer` and `buscarDashboardChanges`
    - After initial `carregar()` completes: start polling timer
    - On each tick: call buscarDashboardChanges, compare timestamps
    - If `saude_updated_at` changed: refetch buscarSaudeColetas
    - If `oportunidades_updated_at` changed: refetch buscarOportunidadesAgora
    - If `performance_updated_at` changed: refetch buscarResumoConversoes + buscarEficaciaAlertas
    - Store local timestamps in reactive state, update after each successful fetch
    - On network error: retain current data, increment error count
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 7.1, 7.2, 10.1, 10.2_

  - [ ] 5.2 Add FreshnessBar to dashboard header
    - Position between h1 title and time-window Select
    - Pass lastUpdate, countdown, and status from polling state
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ] 5.3 Replace static MetricCard values with AnimatedMetric
    - Commission total, conversion count, drop counts, collection counts
    - Format with existing brl/num formatters
    - _Requirements: 5.1, 5.5_

  - [ ] 5.4 Add section highlight animation on data change
    - When a section refetches new data: apply CSS class `animate-highlight` (subtle bg flash)
    - Define keyframe in app.css or inline: background accent → transparent over 1s
    - _Requirements: 5.2_

  - [ ] 5.5 Add entrance animation for new opportunity items
    - When opportunities list changes: new items slide in from top over 300ms
    - Use Svelte `transition:slide` or `transition:fly` on OpportunityCard
    - _Requirements: 5.3_

  - [ ] 5.6 Preserve time-window selector behavior
    - Changing `dias` triggers full refetch of all sections
    - Reset local timestamps (force all sections to refetch on next poll)
    - _Requirements: 10.3, 10.4_

  - [ ] 5.7 Handle auth expiry during polling
    - On 401 from change-detection: stop polling timer
    - Clear local timestamps on user change
    - _Requirements: 8.2, 8.3_

- [ ] 6. Checkpoint — Verify everything works
  - Run `cd web && bun run check` — svelte-check passes
  - Run `cd web && npx vitest run` — all unit tests pass
  - Run `uvx ruff check services/analyzer/` — Python lint passes
  - Run `dotnet build --nologo src/Garimpei.Api/Garimpei.Api.csproj` — C# builds
  - Manual: open `/estatisticas`, observe FreshnessBar counting down, trigger a collection, verify dashboard updates within 30s

## Notes

- The evolution charts (collapsible section) do NOT participate in polling — they refresh only on time-window change (Requirement 6.6)
- The AnimatedMetric component must not animate on initial mount (only on subsequent value changes)
- The `tweened` store from Svelte is imported from `svelte/motion`
- Polling starts AFTER initial load (Requirement 10.2) to avoid race conditions
- The `?intervalo=5000` debug param is useful for E2E testing of the polling mechanism

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "2.1", "3.1", "3.2"] },
    { "id": 1, "tasks": ["1.2", "1.3", "2.2", "3.3", "4.1"] },
    { "id": 2, "tasks": ["5.1", "5.2", "5.3"] },
    { "id": 3, "tasks": ["5.4", "5.5", "5.6", "5.7"] }
  ]
}
```
