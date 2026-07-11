# Technical Design — Dashboard Realtime

## Overview

Smart polling com detecção de mudanças, refresh seletivo por seção, animações de transição, e indicador de freshness. Zero infra nova — usa os endpoints existentes do Analyzer v2 + um endpoint leve de change-detection no C# API.

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    Frontend (Svelte 5 / Cloudflare Pages)          │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Dashboard Page (/estatisticas)                   │ │
│  │                                                              │ │
│  │  ┌────────────┐  ┌─────────────┐  ┌─────────────────────┐  │ │
│  │  │PollingTimer│  │FreshnessBar │  │ AnimatedMetric      │  │ │
│  │  │ (30s)      │  │ countdown   │  │ (tweened values)    │  │ │
│  │  └─────┬──────┘  └─────────────┘  └─────────────────────┘  │ │
│  │        │                                                     │ │
│  │        ▼  GET /api/dashboard/changes                         │ │
│  │  ┌─────────────┐                                            │ │
│  │  │ Change      │ → compara timestamps → refetch seletivo    │ │
│  │  │ Detector    │                                            │ │
│  │  └─────────────┘                                            │ │
│  └──────────────────────────────────────────────────────────────┘ │
└───────────────────────────────┬───────────────────────────────────┘
                                │ HTTP (authenticated)
                                ▼
┌───────────────────────────────────────────────────────────────────┐
│                         C# API (Cloud Run)                         │
│                                                                    │
│  GET /api/dashboard/changes  →  3 cheap BigQuery MAX() queries     │
│       (scoped by owner_uid from JWT)                               │
│                                                                    │
│  Existing endpoints (refetched selectively):                       │
│    /api/coletas/saude  /api/oportunidades/agora                    │
│    /api/conversoes/resumo  /api/alertas/eficacia                   │
└───────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Change Detection Endpoint (`GET /api/dashboard/changes`)

Endpoint leve no C# API que retorna 3 timestamps (um por seção do dashboard):

```json
{
  "saude_updated_at": "2026-07-10T22:15:00Z",
  "oportunidades_updated_at": "2026-07-10T22:15:00Z",
  "performance_updated_at": "2026-07-10T21:30:00Z"
}
```

Implementação: proxy para o Analyzer Python com uma query única:

```python
GET /dashboard/changes
→ SELECT
    MAX(CASE WHEN TRUE THEN coletado_em END) AS saude,
    MAX(criada_em) AS performance
  FROM snapshots CROSS JOIN conversoes
```

Alternativa mais simples: 3 queries baratas em paralelo (MAX de cada tabela).

### 2. PollingTimer (Svelte module)

Novo módulo `$lib/polling.svelte.js`:

```javascript
export function createPollingTimer({ interval = 30000, onTick, immediate = false }) {
  let timer = $state(null);
  let countdown = $state(interval);
  let paused = $state(false);

  function start() { ... }
  function stop() { ... }
  function pause() { ... }
  function resume() { ... }

  // Visibility API integration
  $effect(() => {
    document.addEventListener('visibilitychange', onVisibility);
    return () => document.removeEventListener('visibilitychange', onVisibility);
  });

  return { countdown, paused, start, stop, pause, resume };
}
```

### 3. FreshnessBar (component)

Novo componente `$lib/components/ui/FreshnessBar.svelte`:

```svelte
<FreshnessBar lastUpdate={timestamp} countdown={18} status="live|paused|offline" />
```

Exibe: dot pulsante + "Atualizado há 12s" + "próxima em 18s"

### 4. AnimatedMetric (component)

Novo componente `$lib/components/ui/AnimatedMetric.svelte`:

```svelte
<AnimatedMetric value={127.40} format={brl} duration={600} />
```

Usa Svelte `tweened` store para interpolar valores numéricos com easing.

### 5. Dashboard Integration

A página `/estatisticas` integra os 3 componentes:

```javascript
// On mount: initial fetch (existing)
// After initial fetch: start polling timer
// On each tick: call /api/dashboard/changes
// Compare timestamps → selective refetch
// On data change: animate transitions + highlight section
```

## Data Models

### Change Detection Response

```typescript
interface DashboardChanges {
  saude_updated_at: string | null;       // ISO 8601 UTC
  oportunidades_updated_at: string | null;
  performance_updated_at: string | null;
}
```

### Local State (frontend)

```typescript
interface PollState {
  lastFetched: {
    saude: string | null;
    oportunidades: string | null;
    performance: string | null;
  };
  lastPollAt: number;        // Date.now()
  consecutiveErrors: number;
  interval: number;          // ms (30000 default)
}
```

## Error Handling

### Network failure during polling

- Retain current data, increment `consecutiveErrors`
- After 3 failures: double interval (30s → 60s)
- On success after backoff: reset to default interval
- FreshnessBar shows "sem conexão" (yellow dot)
- No toast/modal — errors are ambient, not disruptive

### Tab hidden

- `document.visibilitychange` → pause timer
- On return: immediate tick + resume
- If hidden >5min: full refetch all sections

### Auth expired during polling

- 401 from change-detection → stop timer
- Firebase `onAuthStateChanged` handles re-auth
- On re-auth: clear local timestamps, restart timer

## Correctness Properties

### Property 1: Polling never fires when tab is hidden

The Visibility API integration pauses the timer. No wasted requests when user isn't looking.

**Validates: Requirements 3.1, 3.2**

### Property 2: Sections only refetch when actually stale

The change-detection endpoint returns server timestamps. Frontend compares against locally stored last-fetched timestamps. A section is refetched only when `server_ts > local_ts`.

**Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5**

### Property 3: Polling failure never breaks the UI

Errors are contained in the FreshnessBar indicator. Existing data remains displayed. No modals, no toasts, no blank screens.

**Validates: Requirements 7.1, 7.2, 7.5**

### Property 4: Multi-tenant isolation guaranteed by auth

Every polling request includes the Firebase JWT. The change-detection endpoint filters by owner_uid from the token. No cross-tenant data leakage possible.

**Validates: Requirements 8.1, 8.2, 8.3**

## File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `services/analyzer/routes/changes.py` | Create | `/dashboard/changes` endpoint (3 MAX queries) |
| `services/analyzer/main.py` | Modify | Register changes router |
| `src/Garimpei.Api/Endpoints/AnalyzerProxyEndpoints.cs` | Modify | Add `/api/dashboard/changes` proxy |
| `web/src/lib/polling.svelte.js` | Create | PollingTimer reactive module |
| `web/src/lib/components/ui/FreshnessBar.svelte` | Create | Live indicator + countdown |
| `web/src/lib/components/ui/AnimatedMetric.svelte` | Create | Tweened numeric values |
| `web/src/lib/components/ui/index.js` | Modify | Export new components |
| `web/src/lib/api.js` | Modify | Add `buscarDashboardChanges()` |
| `web/src/routes/estatisticas/+page.svelte` | Modify | Integrate polling + animations |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| BigQuery MAX() queries on every poll add cost | MAX on timestamp column with partition pruning is near-free. Estimate: <$0.01/month for 1 user polling every 30s. |
| Polling 30s still feels laggy for "real-time" | FreshnessBar countdown + animated transitions create perception of liveness. Configurable down to 10s for power users. |
| Multiple tabs open = multiplied polls | Each tab polls independently. Acceptable — change-detection is cheap. Future: SharedWorker for tab coordination. |
| Tweened animations on many values = performance | Tweened uses requestAnimationFrame. 4-8 metrics animating simultaneously is negligible. |

## Testing Strategy

- **Unit**: PollingTimer lifecycle (start/stop/pause/resume/visibility)
- **Unit**: AnimatedMetric interpolation (value changes animate)
- **Unit**: Change comparison logic (stale detection)
- **Integration**: Tab visibility mock (jsdom visibilityState)
- **E2E**: `mise run test:e2e-services` validates `/api/dashboard/changes` responds
- **Manual**: Open dashboard, trigger collection via `mise run test:e2e-coleta`, observe update within 30s
