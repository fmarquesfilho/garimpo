# Requirements Document

## Introduction

Enable near-real-time dashboard updates (≤30 seconds delay) using a smart polling strategy with visual freshness indicators. When backend events occur — collections completing, alerts firing, publications being sent — the dashboard reflects changes automatically without requiring manual page refresh.

The chosen approach is **Smart Polling with Visual Freshness** (hybrid of approaches A and E from the evaluation). This satisfies the "tempo real com atrasos aceitáveis até 30s" requirement without introducing new infrastructure (no Durable Objects, no Firestore, no SSE). The system polls a lightweight change-detection endpoint every 15–30 seconds, refetches only stale sections, and uses animated transitions plus freshness indicators to convey liveness.

### Prerequisites

- The `analyzer-dashboard-v2` spec must be deployed — the polling mechanism depends on the four Analyzer endpoints (`/coletas/saude`, `/oportunidades/agora`, `/conversoes/resumo`, `/alertas/eficacia`) being available.

### Future Evolution Paths

- **Cloudflare Durable Objects**: If sub-second latency becomes necessary, a DO-per-tenant WebSocket architecture can replace polling without changing the frontend contract (swap polling timer for WebSocket message handler).
- **Server-Sent Events**: If Cloud Run timeouts are extended or the API migrates to a persistent host, SSE provides push semantics with minimal client complexity.
- **Firebase Realtime Database**: Already using Firebase Auth; Firestore listeners could supplement polling for high-priority alerts.

## Glossary

- **Dashboard**: The frontend page at `/estatisticas` (Svelte 5 SPA on Cloudflare Pages) displaying analytics data in three sections: Saúde, Oportunidades, Performance.
- **Polling_Timer**: A recurring interval in the frontend that triggers change detection requests at a configured cadence (default 30 seconds).
- **Change_Detector**: A lightweight API endpoint that returns per-section timestamps indicating when each data domain last changed, allowing the frontend to selectively refetch only stale sections.
- **Freshness_Indicator**: A UI element displaying how recently data was fetched and when the next automatic refresh will occur.
- **Section**: One of the three primary dashboard areas (Saúde, Oportunidades, Performance) that can be independently refreshed.
- **Visibility_State**: The browser's document visibility status (visible or hidden), used to pause/resume polling when the tab is backgrounded.
- **Animated_Transition**: A visual interpolation between old and new metric values (numbers, badges, lists) applied when data changes, preventing jarring jumps.
- **Stale_Section**: A Section whose data has changed on the backend since the frontend last fetched it, as indicated by the Change_Detector response.
- **Owner_UID**: The Firebase Auth user identifier that scopes all data queries to the authenticated tenant.

## Requirements

### Requirement 1: Change Detection Endpoint

**User Story:** As a user, I want the dashboard to know when new data is available without re-querying all BigQuery tables, so that polling is cheap and responsive.

#### Acceptance Criteria

1. WHEN a GET request is made to `/api/dashboard/changes`, THE C#_API SHALL return a JSON response containing per-section timestamps: `saude_updated_at`, `oportunidades_updated_at`, and `performance_updated_at`, each as ISO 8601 UTC strings or null if no data exists.
2. THE C#_API SHALL compute `saude_updated_at` as the MAX `coletado_em` from the most recent snapshot in BigQuery for the authenticated user.
3. THE C#_API SHALL compute `oportunidades_updated_at` as the greater of: the latest snapshot timestamp and the latest publication timestamp for the authenticated user.
4. THE C#_API SHALL compute `performance_updated_at` as the MAX timestamp from the `conversoes` table for the authenticated user.
5. THE C#_API SHALL complete the `/api/dashboard/changes` response within 2 seconds for datasets spanning 180 days.
6. THE C#_API SHALL scope all timestamp queries by the Owner_UID extracted from the Firebase Auth token in the Authorization header.
7. WHEN any underlying table does not exist or contains no rows for the user, THE C#_API SHALL return null for that section's timestamp rather than an error.
8. THE C#_API SHALL return the response with a `Cache-Control: no-store` header to prevent intermediate caching of change-detection data.

### Requirement 2: Polling Timer Lifecycle

**User Story:** As a user, I want the dashboard to check for new data periodically without me refreshing the page, so that I always see current information.

#### Acceptance Criteria

1. WHEN the Dashboard mounts, THE Polling_Timer SHALL start with a default interval of 30 seconds.
2. WHEN the Polling_Timer fires, THE Dashboard SHALL call the Change_Detector endpoint and compare returned timestamps against locally stored last-fetched timestamps for each Section.
3. WHEN a Section's server timestamp is newer than the locally stored timestamp, THE Dashboard SHALL refetch only that Section's data from the corresponding Analyzer endpoint.
4. WHEN no Section timestamps have changed, THE Dashboard SHALL not make any additional API calls beyond the Change_Detector request.
5. WHEN the Dashboard unmounts (navigation away), THE Polling_Timer SHALL be cleared to prevent orphaned requests.
6. THE Polling_Timer SHALL use the existing `authHeaders()` function from `$lib/api.js` to authenticate Change_Detector requests.

### Requirement 3: Tab Visibility Management

**User Story:** As a user, I want polling to pause when I'm not looking at the dashboard tab, so that unnecessary requests don't consume bandwidth and API quota.

#### Acceptance Criteria

1. WHEN the browser tab becomes hidden (document.visibilityState changes to "hidden"), THE Polling_Timer SHALL pause and not fire any Change_Detector requests.
2. WHEN the browser tab becomes visible again (document.visibilityState changes to "visible"), THE Dashboard SHALL immediately fire one Change_Detector request and resume the Polling_Timer at the configured interval.
3. WHEN the tab has been hidden for more than 5 minutes and becomes visible again, THE Dashboard SHALL perform a full refetch of all sections regardless of Change_Detector timestamps.

### Requirement 4: Freshness Indicator Display

**User Story:** As a user, I want to see when data was last refreshed and when the next update will happen, so that I trust the dashboard is live without needing to manually refresh.

#### Acceptance Criteria

1. THE Dashboard SHALL display a Freshness_Indicator in the page header showing: the time of the last successful data refresh in relative format (e.g., "há 12s"), and a countdown to the next poll (e.g., "próxima atualização em 18s").
2. WHILE the Polling_Timer is active and data was fetched within the last 60 seconds, THE Freshness_Indicator SHALL display a green pulsing dot indicating the dashboard is live.
3. WHILE the Polling_Timer is paused (tab hidden), THE Freshness_Indicator SHALL display a gray dot with the text "pausado".
4. WHEN the last Change_Detector request fails (network error or HTTP error), THE Freshness_Indicator SHALL display a yellow dot with the text "sem conexão" and THE Polling_Timer SHALL continue attempting at the configured interval.
5. THE Freshness_Indicator SHALL use the existing StatusIndicator component from the UI library.

### Requirement 5: Animated Value Transitions

**User Story:** As a user, I want metric values to animate smoothly when they change, so that I notice updates without the page feeling jarring.

#### Acceptance Criteria

1. WHEN a numeric metric value changes (commission total, conversion count, collection count, drop percentage), THE Dashboard SHALL animate the transition from the old value to the new value over 600 milliseconds using an ease-out timing function.
2. WHEN a Section receives new data, THE Dashboard SHALL apply a subtle background highlight (fade from accent color to transparent over 1 second) to the Section container to draw attention.
3. WHEN a new item appears in the opportunities list (a new price drop or new product), THE Dashboard SHALL animate the item's entrance with a slide-in-from-top transition over 300 milliseconds.
4. WHEN the health status badge changes value (e.g., "ok" to "atrasado"), THE Dashboard SHALL apply a scale-pulse animation to the badge over 400 milliseconds.
5. THE Dashboard SHALL implement numeric transitions using a Svelte tweened store or equivalent reactive animation primitive, not CSS-only transitions on text content.

### Requirement 6: Selective Section Refetch

**User Story:** As a developer, I want the polling mechanism to only refetch sections whose underlying data changed, so that BigQuery costs remain low and responses are fast.

#### Acceptance Criteria

1. THE Dashboard SHALL maintain a local record of the last-fetched timestamp for each Section, updated after each successful API response.
2. WHEN the Change_Detector indicates only `saude_updated_at` has changed, THE Dashboard SHALL refetch only the `/api/coletas/saude` endpoint and leave Oportunidades and Performance sections untouched.
3. WHEN the Change_Detector indicates only `oportunidades_updated_at` has changed, THE Dashboard SHALL refetch only the `/api/oportunidades/agora` endpoint.
4. WHEN the Change_Detector indicates only `performance_updated_at` has changed, THE Dashboard SHALL refetch only the `/api/conversoes/resumo` and `/api/alertas/eficacia` endpoints.
5. WHEN multiple Section timestamps have changed simultaneously, THE Dashboard SHALL refetch all affected sections in parallel.
6. THE Dashboard SHALL not refetch the evolution charts (collapsible section) during polling — those remain manual-refresh only via the time window selector.

### Requirement 7: Error Resilience During Polling

**User Story:** As a user, I want polling failures to not break my dashboard experience, so that temporary network issues don't require a page reload.

#### Acceptance Criteria

1. WHEN a Change_Detector request fails due to network error, THE Dashboard SHALL retain the currently displayed data and retry on the next polling interval.
2. WHEN a Section refetch fails after the Change_Detector indicated new data, THE Dashboard SHALL retain the previous Section data, display the error in the Freshness_Indicator, and retry the Section refetch on the next successful Change_Detector cycle.
3. IF three consecutive Change_Detector requests fail, THEN THE Dashboard SHALL double the polling interval (from 30s to 60s) to reduce load during sustained outages.
4. WHEN a Change_Detector request succeeds after a backoff period, THE Dashboard SHALL reset the polling interval to the default 30 seconds.
5. THE Dashboard SHALL not display error toasts or modal dialogs for polling failures — errors are conveyed solely through the Freshness_Indicator.

### Requirement 8: Multi-Tenant Data Isolation

**User Story:** As a user, I want the real-time updates to show only my data, so that I never see another tenant's events.

#### Acceptance Criteria

1. THE Change_Detector endpoint SHALL extract the Owner_UID from the validated Firebase Auth JWT token and use it as the sole tenant filter for all timestamp queries.
2. WHEN the Firebase Auth token is expired or missing, THE Change_Detector endpoint SHALL return HTTP 401 and THE Dashboard SHALL stop the Polling_Timer until re-authentication succeeds.
3. THE Dashboard SHALL not cache Change_Detector responses across different authenticated sessions — when the user logs out and another user logs in, all locally stored timestamps SHALL be cleared.

### Requirement 9: Configurable Polling Interval

**User Story:** As a developer, I want the polling interval to be configurable without code changes, so that I can tune it based on observed load and user experience.

#### Acceptance Criteria

1. THE Dashboard SHALL read the polling interval from a Svelte environment variable `VITE_POLL_INTERVAL_MS` with a default of 30000 milliseconds.
2. THE Dashboard SHALL enforce a minimum polling interval of 10000 milliseconds and a maximum of 120000 milliseconds, clamping any configured value to this range.
3. THE Dashboard SHALL accept an optional `intervalo` query parameter on the `/estatisticas` page URL that overrides the environment variable for debugging purposes (e.g., `?intervalo=5000`).

### Requirement 10: Initial Load Compatibility

**User Story:** As a user, I want the dashboard to load fully on first visit exactly as before, with polling enhancing the experience after the initial render.

#### Acceptance Criteria

1. THE Dashboard SHALL perform the initial data fetch for all sections on mount (existing `carregar()` behavior) without waiting for the Polling_Timer.
2. THE Dashboard SHALL start the Polling_Timer only after the initial load completes (all sections have either resolved or errored).
3. THE Dashboard SHALL preserve the existing time window selector (`dias` state) behavior — changing the time window triggers a full refetch of all sections and resets the locally stored timestamps.
4. THE Dashboard SHALL preserve the existing parallel loading pattern where each Section resolves independently.
