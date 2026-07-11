# Requirements Document

## Introduction

The Collector service needs a `Collect` RPC that combines product searching with BigQuery persistence. Currently, the Scheduler calls `Fetch`/`FetchShop` (pure reads) and then persists snapshots itself via `internal/store` — violating the architecture rule that the Scheduler should not depend on `internal/store`. The `Collect` RPC moves persistence responsibility to the Collector, which already has a `BigQueryExporterConfig` in its YAML config (ADR-0018). This enables the Python Analyzer to query BigQuery snapshots for "Novidades" (new products) and "Quedas" (price drops), which are currently empty because no snapshots are being written from the Collector path.

### Prerequisites

- The frontend fix (sending literal keyword/shop_id as `busca_id` to the Analyzer instead of a UUID) is already deployed.

## Glossary

- **Collector**: The unified gRPC service (`services/collector`) that searches products across marketplaces via config-driven receivers (ADR-0018).
- **Collect_RPC**: The new gRPC method on CollectorService that searches products AND persists the results as BigQuery snapshots. Distinct from `Fetch` which remains side-effect-free.
- **Fetch_RPC**: The existing gRPC method on CollectorService that searches products and returns them without any side-effects.
- **FetchShop_RPC**: The existing gRPC method on CollectorService that fetches products from a specific shop without any side-effects.
- **BigQuery_Exporter**: The component within the Collector responsible for writing snapshot rows to BigQuery. Configured via the `exporters.bigquery` section of the Collector YAML config.
- **Snapshot**: A point-in-time capture of product search results, stored as rows in BigQuery's `snapshots` table. Contains keyword, strategy, timestamp, and a list of product items with position, price, commission, sales, rating, image, link, and shop name.
- **Scheduler**: The orchestration service that triggers periodic collection jobs via cron. Must not depend on `internal/store`.
- **Pipeline**: The Collector's internal orchestrator that manages receivers (product sources) by marketplace.
- **Analyzer**: The Python service that queries BigQuery snapshots to detect new products ("Novidades") and price drops ("Quedas").
- **Graceful_Degradation**: Behavior where the Collect_RPC still returns products successfully even when the BigQuery_Exporter is not configured or encounters a transient error during persistence.

## Requirements

### Requirement 1: Define the Collect RPC in the Proto

**User Story:** As a developer, I want a single `Collect` RPC defined in `collector.proto`, so that clients can call one method to search (by keyword or shop) and persist products.

#### Acceptance Criteria

1. THE CollectorService proto SHALL define a `Collect` RPC method that accepts a `CollectRequest` and returns a `CollectResponse`.
2. THE `CollectRequest` message SHALL contain a `oneof target` with two options: `keyword` (string) for keyword-based search, and `shop_id` (int64) for shop-based fetch.
3. THE `CollectRequest` message SHALL contain fields for limit (int32), sort_by (string), owner_uid (string), and marketplace (enum Marketplace).
4. THE `CollectResponse` message SHALL contain fields for products (repeated Product), total_found (int32), fetched_at (string RFC3339), and persisted (bool indicating snapshot was accepted for export).
5. THE `Collect` RPC SHALL be additive to the existing proto definition — existing RPCs (Fetch, FetchShop, ResolveShop, GenerateAffiliateLink) SHALL remain unchanged.

### Requirement 2: Implement Collect with keyword target as Search Plus Persist

**User Story:** As the Scheduler, I want to call `Collect` with a keyword to get products and have them persisted to BigQuery, so that the Analyzer can detect new products and price drops.

#### Acceptance Criteria

1. WHEN a CollectRequest with `keyword` target is received, THE Collector SHALL search products using the same logic as the Fetch RPC for the specified marketplace and keyword.
2. WHEN product search succeeds AND the BigQuery exporter is configured, THE Collector SHALL enqueue the results for asynchronous persistence as a Snapshot to BigQuery.
3. THE Snapshot SHALL use the literal keyword from the request as the `keyword` field in BigQuery — not a UUID or internal identifier.
4. THE Snapshot SHALL include position (1-based), produto_id, nome, preco, comissao, vendas, nota, imagem, link, and loja for each product item.
5. WHEN the keyword is empty, THE Collector SHALL return an InvalidArgument error without searching or persisting.
6. WHEN product search succeeds but returns zero products, THE Collector SHALL return an empty product list and SHALL NOT enqueue a Snapshot for export.
7. WHEN product search succeeds, THE Collector SHALL return the products in the CollectResponse regardless of whether export was accepted or not.

### Requirement 3: Implement Collect with shop_id target as Shop Fetch Plus Persist

**User Story:** As the Scheduler, I want to call `Collect` with a shop_id to fetch shop products and persist them, so that shop monitoring also produces BigQuery snapshots.

#### Acceptance Criteria

1. WHEN a CollectRequest with `shop_id` target is received, THE Collector SHALL fetch shop products using the same logic as the FetchShop RPC.
2. WHEN shop product fetch succeeds AND the BigQuery exporter is configured, THE Collector SHALL enqueue the results for asynchronous persistence as a Snapshot using the shop_id (base-10 string representation) as the `keyword` field.
3. WHEN the shop_id is zero, THE Collector SHALL return an InvalidArgument error without fetching or persisting.
4. WHEN neither keyword nor shop_id is set in the oneof, THE Collector SHALL return an InvalidArgument error.
5. WHEN shop product fetch succeeds, THE Collector SHALL return the products in the CollectResponse regardless of whether export was accepted or not.

### Requirement 4: Asynchronous Export via Buffered Channel

**User Story:** As an operator, I want the Collect RPC to return immediately after searching without waiting for BigQuery, so that persistence latency does not affect response time.

#### Acceptance Criteria

1. THE Collector SHALL maintain a buffered channel (capacity 64) for snapshot export.
2. WHEN a snapshot is enqueued successfully, THE Collector SHALL set `persisted` to true in the response.
3. WHEN the export channel is full, THE Collector SHALL set `persisted` to false, log a warning, and return the products normally without blocking.
4. A background goroutine SHALL drain the channel and call `RegistrarSnapshot` for each snapshot.
5. ON graceful shutdown (context cancellation), THE goroutine SHALL drain remaining snapshots from the channel before exiting.

### Requirement 5: BigQuery Exporter is Config-Driven and Optional

**User Story:** As an operator, I want BigQuery export to be controlled by the YAML config, so that environments without BigQuery (local dev, tests) work without errors.

#### Acceptance Criteria

1. WHEN the `exporters.bigquery` section is present in the Collector YAML config AND `project_env` resolves to a non-empty string AND `dataset` and `products_table` are non-empty, THE BigQuery exporter SHALL be initialized at startup.
2. WHEN the `exporters.bigquery` section is absent or incomplete, THE Collector SHALL use `store.NopSnapshots()` and log a warning indicating export is disabled.
3. WHILE using NopSnapshots, THE Collect RPC SHALL search and return products normally. The `persisted` field SHALL be true (snapshot accepted into channel) but no data is written.
4. WHEN the BigQuery exporter is configured, snapshots dequeued by the background goroutine SHALL be written via `store.RegistrarSnapshot`.

### Requirement 6: Graceful Degradation on Persistence Errors

**User Story:** As a developer, I want the Collect RPC to never fail due to a BigQuery write error, so that product data is always returned to the caller.

#### Acceptance Criteria

1. IF a BigQuery write error occurs in the background goroutine, THEN it SHALL log the error at error level with keyword, item count, and error message.
2. IF a BigQuery write error occurs, THE snapshot SHALL be discarded without retry (the next Scheduler cycle produces a fresh snapshot).
3. THE Collect RPC response is never affected by BigQuery errors since export is asynchronous — the response was already returned before persistence is attempted.

### Requirement 7: Scheduler Calls Collect Instead of Fetch

**User Story:** As a maintainer, I want the Scheduler to call `Collect` instead of `Fetch`/`FetchShop` plus local persistence, so that the Scheduler no longer depends on `internal/store`.

#### Acceptance Criteria

1. WHEN executing a keyword search job, THE Scheduler SHALL call the Collect RPC with `keyword` target, passing the same keyword, limit (50), marketplace, and owner_uid.
2. WHEN executing a shop collection job, THE Scheduler SHALL call the Collect RPC with `shop_id` target, passing the same shop_id, limit (50), marketplace, and owner_uid.
3. THE Scheduler SHALL NOT import or reference `internal/store` — the `persistSnapshot` method, `store.SnapshotRepo` field, `initSnapshotStore` function, and build-tagged store files SHALL be removed.
4. IF the Collect RPC returns a gRPC error, THEN THE Scheduler SHALL log the error with job name and keyword/shop_id, skip alert enqueueing, and continue processing remaining keywords.
5. WHEN a Collect RPC returns successfully, THE Scheduler SHALL enqueue a price alert task via Cloud Tasks with the same payload structure as current.
6. WHEN arch-go runs, THE Scheduler SHALL pass with zero violations including no dependency on `internal/store`.

### Requirement 8: Backward Compatibility

**User Story:** As a developer, I want the existing Fetch and FetchShop RPCs to remain unchanged, so that any clients depending on the pure-read behavior continue to work.

#### Acceptance Criteria

1. THE Fetch RPC SHALL continue to search and return products without touching the export channel or BigQuery.
2. THE FetchShop RPC SHALL continue to fetch shop products without touching the export channel or BigQuery.
3. THE Fetch and FetchShop request/response message schemas SHALL not change.
4. Generated Go and C# stubs SHALL remain backward compatible — existing compiled clients SHALL not require recompilation.
