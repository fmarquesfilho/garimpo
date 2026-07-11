# Requirements Document

## Introduction

O sistema Garimpei atualmente faz requests diretamente ao Collector (Go gRPC sidecar) para cada leitura de dados de produto. Com dados que mudam a cada 8h e frontend polling a cada 30s, a carga é desnecessária. Este feature introduz uma camada de cache em dois níveis:

- **L1 (Cloudflare Workers Cache)**: Cache HTTP no edge, absorvendo ~90% das leituras repetitivas sem custo adicional.
- **L2 (Go Cache Sidecar)**: Cache in-memory interno ao Cloud Run, protegendo o Collector contra thundering herd quando L1 expira.

Ambas as camadas validam o BuscaContract, garantindo consistência end-to-end do schema canônico. A invalidação é event-driven: o Collector, após completar uma coleta, compara os dados novos com o cache e invalida seletivamente quando há divergência.

## Glossary

- **Cache_Sidecar**: Novo serviço Go deployado como sidecar no Cloud Run (mesma pod que C# API e Collector). Responsável pelo cache L2 in-memory e validação de BuscaContract.
- **Workers_Cache**: Feature nativa do Cloudflare Workers (2025) que permite cache HTTP de responses antes da execução do Worker. Camada L1 no edge.
- **Collector**: Serviço Go gRPC existente que busca produtos em marketplaces externos (Shopee, Amazon, etc.) e persiste snapshots no BigQuery.
- **C#_API**: Gateway HTTP principal (ASP.NET) que orquestra chamadas gRPC aos sidecars.
- **Cloudflare_Worker**: Worker JS existente (`garimpei-proxy`) que roteia tráfego entre frontend e Cloud Run.
- **BuscaContract**: Schema canônico JSON (contracts/schemas/busca-contract.json) que define a estrutura de uma Busca cross-service.
- **collection_keys**: Array derivado deterministicamente dos campos da busca (shop_ids, keywords, categorias). Usado como chave de cache no L2.
- **busca_id**: UUID estável que identifica uma busca. Usado como tag para purge granular no L1.
- **Purge_API**: API HTTP do Cloudflare para invalidar cache por tag (Cache-Tag header).
- **thundering_herd**: Padrão onde múltiplas requests simultâneas atingem o backend quando o cache expira ao mesmo tempo.

## Requirements

### Requirement 1: L1 — Cloudflare Workers Cache

**User Story:** As a system operator, I want HTTP responses cached at the Cloudflare edge, so that repeated frontend polls are served without reaching Cloud Run.

#### Acceptance Criteria

1. WHEN the Cloudflare_Worker receives a GET request to `/api/*` that matches a cacheable route, THE Cloudflare_Worker SHALL set a `Cache-Control: public, max-age=300` header on the upstream response before returning it to the client.
2. WHEN a response includes a `Cache-Control` header with `max-age > 0`, THE Workers_Cache SHALL serve subsequent identical requests from cache without forwarding to Cloud Run until the TTL expires.
3. THE Cloudflare_Worker SHALL include a `Cache-Tag` header with the value `busca:{busca_id}` on all responses that contain product data associated with a specific busca, regardless of whether the response is served from cache or passed through directly.
4. WHEN the Cloudflare_Worker receives a request for a non-cacheable route (POST, PUT, DELETE, or routes without busca_id context), THE Cloudflare_Worker SHALL bypass the Workers_Cache, forward the request directly to Cloud Run, and SHALL NOT serve any previously cached response for that route.
5. THE Cloudflare_Worker SHALL include a `X-Cache-Status` response header indicating `HIT`, `MISS`, or `BYPASS` so that observability tooling can track cache effectiveness.

### Requirement 2: L2 — Go Cache Sidecar Service

**User Story:** As a system operator, I want an in-memory cache sidecar between the C# API and the Collector, so that L1 misses are served from memory without calling external marketplace APIs.

#### Acceptance Criteria

1. THE Cache_Sidecar SHALL expose a gRPC service (`cache.v1.CacheService`) with methods `Get`, `Invalidate`, and `Healthz`.
2. WHEN the Cache_Sidecar receives a `Get` request with a set of collection_keys, THE Cache_Sidecar SHALL return the cached product data if all requested keys exist in the in-memory store.
3. WHEN the Cache_Sidecar receives a `Get` request and one or more collection_keys are not present in the in-memory store (cache miss), THE Cache_Sidecar SHALL call the Collector's `Fetch` or `FetchShop` gRPC method to retrieve the data, store the result in memory, and return it to the caller.
4. THE Cache_Sidecar SHALL store cached entries keyed by each individual collection_key derived from BuscaContract using the `DeriveCollectionKeys` function from the `internal/busca` package.
5. THE Cache_Sidecar SHALL be deployed as a sidecar container in the same Cloud Run multi-container pod as the C# API and Collector, communicating over localhost gRPC (port 50055).
6. THE Cache_Sidecar SHALL limit total in-memory cache size to a configurable maximum (default 256 MB) and evict existing entries using LRU policy to make room for new data when the limit is reached.

### Requirement 3: BuscaContract Validation in Cache Layer

**User Story:** As a developer, I want the cache layer to validate BuscaContract on every cached entry, so that invalid or stale schema data never reaches consumers.

#### Acceptance Criteria

1. WHEN the Cache_Sidecar stores a new entry in the in-memory cache, THE Cache_Sidecar SHALL validate the entry against the BuscaContract JSON Schema (contracts/schemas/busca-contract.json) before insertion.
2. IF a cached entry fails BuscaContract validation, THEN THE Cache_Sidecar SHALL reject the request entirely, log a structured error with the validation details, and return an InvalidArgument gRPC status to the caller.
3. THE Cache_Sidecar SHALL use the same `DeriveCollectionKeys` function (internal/busca package) as all other Go services to compute cache keys, ensuring cross-service consistency.
4. WHEN the Cache_Sidecar returns a cached entry, THE Cache_Sidecar SHALL include a `schema_version` field in the gRPC response metadata indicating the BuscaContract schema version used for validation.

### Requirement 4: Cache Invalidation Flow

**User Story:** As a system operator, I want caches invalidated automatically when the Collector detects data divergence, so that users always see fresh data after a collection completes.

#### Acceptance Criteria

1. WHEN the Collector completes a `Collect` operation and the collected data diverges from the L2 cached data for the same collection_keys, THE Collector SHALL call `CacheService.Invalidate` on the Cache_Sidecar via local gRPC with the affected busca_id.
2. WHEN the Collector completes a `Collect` operation and the collected data diverges from the L2 cached data, THE Collector SHALL call the Cloudflare Purge API with tag `busca:{busca_id}` to invalidate L1 cache entries.
3. WHEN the Cache_Sidecar receives an `Invalidate` request with a busca_id, THE Cache_Sidecar SHALL remove all cached entries whose collection_keys are associated with that busca_id.
4. IF the Cloudflare Purge API call fails (network error, timeout, or non-2xx response), THEN THE Collector SHALL log the failure with structured context, increment retry_count to 1, and retry once after a 1-second delay.
5. IF the Cloudflare Purge API retry also fails (retry_count = 1), THEN THE Collector SHALL log the failure as a warning, skip the L1 purge entirely, and proceed without blocking the Collect response (L1 will expire naturally via TTL).
6. THE Collector SHALL complete the L2 invalidation (local gRPC, target latency under 1ms) before initiating the L1 purge (HTTP call, target latency under 100ms).

### Requirement 5: C# API Integration with Cache Sidecar

**User Story:** As a developer, I want the C# API to read product data from the Cache Sidecar instead of calling the Collector directly, so that cached data is used transparently.

#### Acceptance Criteria

1. WHEN the C#_API needs to read product data for a busca, THE C#_API SHALL call `CacheService.Get` on the Cache_Sidecar gRPC service instead of calling `CollectorService.Fetch` or `CollectorService.FetchShop` directly. All product read requests SHALL go through the Cache_Sidecar when it is reachable.
2. THE C#_API SHALL derive collection_keys from the BuscaContract using the same deterministic logic as the Go implementation (`DeriveCollectionKeys`), ensuring cache key consistency.
3. WHEN the Cache_Sidecar is unreachable (connection refused, timeout after 500ms), THE C#_API SHALL fall back to calling the Collector directly, log the fallback event, and continue serving the request. IF the Collector fallback also fails, THE C#_API SHALL return partial data or an empty structured response (never HTTP 500).
4. THE C#_API SHALL include a `X-Cache-Source` header in HTTP responses with value `l2-hit`, `l2-miss`, or `l2-bypass` to indicate cache utilization at the L2 layer.

### Requirement 6: Graceful Degradation

**User Story:** As a system operator, I want the system to continue serving requests even when cache layers are unavailable, so that a cache failure never causes a full outage.

#### Acceptance Criteria

1. WHEN a read request arrives and the Cache_Sidecar is unavailable (health check fails or gRPC connection refused), THE C#_API SHALL route that request directly to the Collector and log a degradation warning (rate-limited to once per minute).
2. IF the Workers_Cache is unavailable or returns errors, THEN THE Cloudflare_Worker SHALL pass requests through to Cloud Run without modification (transparent proxy behavior).
3. WHEN the Cache_Sidecar recovers after being unavailable, THE C#_API SHALL resume sending read requests to the Cache_Sidecar within 10 seconds of health recovery (circuit breaker reset).
4. THE Cache_Sidecar SHALL expose a `Healthz` gRPC method that returns the current cache size, hit/miss counts, and a ready status for use by the C#_API circuit breaker and Cloud Run health probes.

### Requirement 7: Cache Observability

**User Story:** As a system operator, I want visibility into cache hit rates, latencies, and invalidation events, so that I can monitor cache effectiveness and debug issues.

#### Acceptance Criteria

1. THE Cache_Sidecar SHALL emit structured logs (JSON format) for every cache hit, cache miss, invalidation event, and validation failure, including busca_id and collection_keys in each log entry.
2. THE Cache_Sidecar SHALL expose Prometheus-compatible metrics: `cache_hits_total`, `cache_misses_total`, `cache_invalidations_total`, `cache_size_bytes`, and `cache_latency_seconds` histogram.
3. THE Cloudflare_Worker SHALL log cache status (HIT/MISS/BYPASS) for each request to enable L1 hit-rate calculation from worker analytics.
4. WHEN an invalidation event occurs, THE Cache_Sidecar SHALL log the time delta between the last cache write and the invalidation trigger, enabling staleness measurement.

### Requirement 8: Cache Key Derivation Consistency

**User Story:** As a developer, I want cache keys derived identically across all services, so that cache lookups and invalidations always target the correct entries.

#### Acceptance Criteria

1. THE Cache_Sidecar SHALL use the `DeriveCollectionKeys` function from `internal/busca` package to compute cache keys from shop_ids, keywords, and categorias.
2. THE C#_API SHALL implement a `DeriveCollectionKeys` function that produces output identical to the Go implementation for the same inputs (validated by cross-language fixture tests in CI).
3. FOR ALL valid BuscaContract instances, calling `DeriveCollectionKeys` with the same inputs in Go and C# SHALL produce identical sorted arrays (round-trip property).
4. WHEN the BuscaContract schema is updated, THE CI pipeline SHALL validate that all `DeriveCollectionKeys` implementations (Go, C#) produce consistent output for the shared test fixtures (`fixtures/buscas.json`).

### Requirement 9: Proto Definition for Cache Service

**User Story:** As a developer, I want a well-defined gRPC proto for the Cache service, so that the C# API and Collector can communicate with the Cache Sidecar using typed contracts.

#### Acceptance Criteria

1. THE Cache_Sidecar proto (`protos/cache/v1/cache.proto`) SHALL define a `CacheService` with methods: `Get(GetRequest) returns (GetResponse)`, `Invalidate(InvalidateRequest) returns (InvalidateResponse)`, and `Healthz(HealthzRequest) returns (HealthzResponse)`.
2. THE `GetRequest` message SHALL include fields: `repeated string collection_keys`, `string busca_id`, and `collector.v1.Marketplace marketplace`.
3. THE `GetResponse` message SHALL include fields: `repeated collector.v1.Product products`, `bool cache_hit`, `string fetched_at`, and `string schema_version`.
4. THE `InvalidateRequest` message SHALL include fields: `string busca_id` and `repeated string collection_keys`.
5. THE `InvalidateResponse` message SHALL include fields: `int32 keys_removed` and `bool success`.
6. THE `HealthzResponse` message SHALL include fields: `bool ready`, `int64 cache_size_bytes`, `int64 hits_total`, `int64 misses_total`, and `int64 entries_count`.
