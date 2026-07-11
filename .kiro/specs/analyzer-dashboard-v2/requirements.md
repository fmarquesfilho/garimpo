# Requirements Document

## Introduction

Redesign the Garimpei Analyzer API and Dashboard to answer the three user-centric questions that matter: "Are my collections running?", "What can I publish now?", and "Am I making money?". The current dashboard displays vanity metrics (total products, average price, average commission) that provide no actionable insight. This spec introduces four new Analyzer endpoints and restructures the frontend `/estatisticas` page into three goal-oriented sections with the existing evolution charts preserved as collapsible secondary content.

### Prerequisites

- The `collector-collect-rpc` spec (BigQuery snapshot persistence) must be deployed — without snapshots being written to BigQuery, the Analyzer has no data source.

## Glossary

- **Analyzer**: The Python FastAPI service (`services/analyzer`) that queries BigQuery for analytics data and serves it to the frontend.
- **Dashboard**: The frontend page at `/estatisticas` that displays analytics data from the Analyzer.
- **Snapshot**: A point-in-time capture of product search results stored in BigQuery's `snapshots` table, containing keyword, timestamp, and product details (price, commission, sales, rating, image, link, shop).
- **Collection**: A single execution of the Collector that produces a Snapshot. Scheduled via cron in the Scheduler.
- **Opportunity**: A product with an active price drop or newly detected that has not yet been published by the user.
- **Conversion**: A confirmed sale tracked via the Shopee affiliate API, stored in BigQuery's `conversoes` table with sub_id, product, channel, commission, and timestamp.
- **Publication**: A product listing published by the user to a channel (Telegram, WhatsApp). Stored with status (enviada, erro) and metadata.
- **Health_Status**: A computed assessment of collection system health — whether scheduled collections are executing on time and producing data.
- **Drop_Magnitude**: The absolute percentage decrease in a product's price between its first and latest snapshot within a time window.
- **Detection_Rate**: The ratio of price drops detected to alerts actually sent to the user.
- **Graceful_Degradation**: Behavior where endpoints return valid empty-state responses when underlying data tables do not exist or contain no rows.

## Requirements

### Requirement 1: Collection Health Endpoint

**User Story:** As a user, I want to see whether my scheduled collections are running correctly, so that I know my monitoring setup is healthy without checking logs.

#### Acceptance Criteria

1. WHEN a GET request is made to `/coletas/saude`, THE Analyzer SHALL return a JSON response containing: last collection timestamp, delay status, count of collections in the last 24 hours, expected collection count based on configured schedules, and a list of keywords with no collection in the last 24 hours.
2. WHEN the last collection timestamp is older than 6 hours, THE Analyzer SHALL set the delay status to "atrasado" (delayed).
3. WHEN the last collection timestamp is within 6 hours, THE Analyzer SHALL set the delay status to "ok".
4. WHEN no snapshots exist in BigQuery, THE Analyzer SHALL return a valid response with last collection as null, delay status "sem_dados", zero collections in 24 hours, and an empty stale keywords list.
5. THE Analyzer SHALL compute the expected collection count by counting distinct keywords in snapshots from the last 7 days and multiplying by 1 (assuming one daily collection per keyword).
6. THE Analyzer SHALL complete the `/coletas/saude` query within 5 seconds for datasets spanning 30 days of snapshots.

### Requirement 2: Opportunities Endpoint

**User Story:** As a user, I want to see products I should publish right now, so that I can act on price drops and new products without manually searching.

#### Acceptance Criteria

1. WHEN a GET request is made to `/oportunidades/agora`, THE Analyzer SHALL return a JSON response containing: top active price drops sorted by Drop_Magnitude descending (limited to 10), new products detected in the last 48 hours (limited to 10), and high-value unpublished products (limited to 5).
2. THE Analyzer SHALL identify active price drops by comparing the earliest and latest price for each product within a configurable time window (default 7 days, query parameter `dias`, range 1-30).
3. THE Analyzer SHALL exclude already-published products from the opportunities list by performing a LEFT JOIN against the BigQuery `publicacoes` table and filtering out rows where a matching publication exists.
4. THE Analyzer SHALL identify high-value unpublished products as those with commission above the 75th percentile AND sales above the median, that have not been published.
5. WHEN the `publicacoes` table does not exist in BigQuery, THE Analyzer SHALL return opportunities without the publication exclusion filter and include a field `filtro_publicacoes` set to false in the response.
6. WHEN no price drops or new products exist, THE Analyzer SHALL return empty lists with zero totals.
7. THE Analyzer SHALL complete the `/oportunidades/agora` query within 5 seconds for datasets spanning 30 days of snapshots.

### Requirement 3: Revenue Summary Endpoint

**User Story:** As a user, I want to see how much money my publications are generating, so that I can evaluate whether this effort is profitable.

#### Acceptance Criteria

1. WHEN a GET request is made to `/conversoes/resumo`, THE Analyzer SHALL return a JSON response containing: total estimated commission in the time window, number of conversions, number of distinct products that converted, and a breakdown by channel (canal).
2. THE Analyzer SHALL accept a `dias` query parameter (default 30, range 1-180) to define the time window.
3. THE Analyzer SHALL compute the best performing channel as the channel with the highest total commission in the time window.
4. WHEN the `conversoes` table does not exist or contains no rows, THE Analyzer SHALL return a valid response with zero commission, zero conversions, an empty channel breakdown, and status "sem_dados".
5. THE Analyzer SHALL complete the `/conversoes/resumo` query within 5 seconds for datasets spanning 180 days of conversions.

### Requirement 4: Alert Efficacy Endpoint

**User Story:** As a user, I want to understand whether my price-drop alerts are leading to actual conversions, so that I can tune my alert thresholds.

#### Acceptance Criteria

1. WHEN a GET request is made to `/alertas/eficacia`, THE Analyzer SHALL return a JSON response containing: total price drops detected in the time window, total alerts sent (publications triggered by drops), total conversions attributed to alerted products, Detection_Rate, and best performing keyword (keyword with highest conversion count).
2. THE Analyzer SHALL accept a `dias` query parameter (default 30, range 1-180) to define the time window.
3. THE Analyzer SHALL compute Detection_Rate as the ratio of alerts sent to price drops detected, expressed as a percentage.
4. THE Analyzer SHALL attribute a conversion to an alerted product when the conversion's `produto_id` matches a product that had a price drop detected in the same time window.
5. WHEN no price drops are detected in the time window, THE Analyzer SHALL return zero for all metrics and Detection_Rate as null.
6. WHEN the `conversoes` table does not exist, THE Analyzer SHALL return the drop and alert counts with conversion metrics set to zero and a field `conversoes_disponiveis` set to false.

### Requirement 5: Dashboard Frontend Restructure

**User Story:** As a user, I want the dashboard organized around my three key questions instead of vanity metrics, so that I immediately see what matters when I open the page.

#### Acceptance Criteria

1. THE Dashboard SHALL display three primary sections in this order: "Saúde" (health), "Oportunidades" (opportunities), "Performance" (revenue).
2. THE "Saúde" section SHALL display: the delay status as a colored badge (green for "ok", yellow for "atrasado", gray for "sem_dados"), last collection time in relative format, collections in 24h vs expected count, and a list of stale keywords if any exist.
3. THE "Oportunidades" section SHALL display: top price drops as a ranked list showing product name, drop percentage, current price, and link; new products count; and high-value unpublished products count.
4. THE "Performance" section SHALL display: total commission earned formatted as BRL currency, number of conversions, best channel name, and Detection_Rate percentage.
5. THE Dashboard SHALL preserve the existing evolution charts (price evolution by store and by keyword) and collection history as collapsible panels below the three primary sections.
6. THE Dashboard SHALL use existing UI components: MetricCard, DashPanel, MiniChart, RankList, Badge, Alert, and Select.
7. WHEN any endpoint returns an error or empty state, THE Dashboard SHALL display a contextual empty-state message within that section without breaking the other sections.

### Requirement 6: Endpoint Response Time and Graceful Degradation

**User Story:** As a user, I want the dashboard to load quickly and work even when data is sparse, so that new users and users with partial setups are not confused by error screens.

#### Acceptance Criteria

1. THE Analyzer SHALL return a valid JSON response for all new endpoints even when the underlying BigQuery tables do not exist, returning structured empty-state responses instead of HTTP 500 errors.
2. WHEN a BigQuery query exceeds 5 seconds, THE Analyzer SHALL return a partial response with the data gathered so far and a field `timeout` set to true, rather than making the client wait indefinitely.
3. THE Analyzer SHALL use parameterized queries with `@parameter` syntax for all user-provided values to prevent SQL injection.
4. THE Analyzer SHALL not introduce new Python dependencies unless the functionality cannot be achieved with the existing FastAPI, google-cloud-bigquery, and opentelemetry stack.

### Requirement 7: Backward Compatibility of Existing Endpoints

**User Story:** As a developer, I want the existing 7 Analyzer endpoints to continue working unchanged, so that any frontend pages or external consumers relying on them are not broken.

#### Acceptance Criteria

1. THE Analyzer SHALL retain all existing endpoints (`/estatisticas`, `/coletas`, `/evolucao`, `/quedas`, `/novidades`, `/conversoes`, `/cupons`) with their current request and response schemas unchanged.
2. THE Analyzer SHALL register the new endpoints as additional routes without modifying existing router registrations.
3. THE existing `/coletas` page in the frontend SHALL continue to function with its current data source and display format.
