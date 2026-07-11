# Implementation Plan: Collector Collect RPC

## Overview

Add a `Collect` RPC to the CollectorService that combines product searching with asynchronous BigQuery persistence. Migrate the Scheduler to use `Collect` instead of `Fetch`/`FetchShop` + local persistence, removing its dependency on `internal/store`. Implementation proceeds proto-first, then Collector exporter infrastructure, then the Collect method, then Scheduler migration, and finally Dockerfile/arch-go updates.

## Tasks

- [ ] 1. Define Collect RPC in proto and regenerate stubs
  - [ ] 1.1 Add CollectRequest, CollectResponse messages and Collect RPC to `protos/collector/v1/collector.proto`
    - Add `CollectRequest` message with `oneof target { string keyword = 1; int64 shop_id = 6; }`, plus fields `limit`, `sort_by`, `owner_uid`, `marketplace`
    - Add `CollectResponse` message with `repeated Product products`, `int32 total_found`, `string fetched_at`, `bool persisted`
    - Add `rpc Collect(CollectRequest) returns (CollectResponse)` to `CollectorService`
    - Existing RPCs (Fetch, FetchShop, ResolveShop, GenerateAffiliateLink) remain unchanged
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 8.3, 8.4_

  - [ ] 1.2 Run `buf generate` from `protos/` directory to regenerate Go and C# stubs
    - Verify `gen/go/collector/v1/` contains the new `Collect` method on the client/server interfaces
    - Verify `src/Garimpei.Protos/` regenerates without errors
    - _Requirements: 1.1, 8.4_

- [ ] 2. Implement Collector exporter infrastructure
  - [ ] 2.1 Create `services/collector/store_gcp.go` with build tag `gcp`
    - Implement `initSnapshots(ctx context.Context, cfg BigQueryExporterConfig, logger *slog.Logger) store.SnapshotRepo`
    - Resolve `ProjectEnv` via `ResolveCredentialEnv`, check `Dataset` and `ProductsTable` non-empty
    - Return `store.NopSnapshots()` with warning log if config incomplete
    - Call `store.NovoBigQueryStore(ctx, project, cfg.Dataset, "eventos", cfg.ProductsTable)` on success
    - Return `store.NopSnapshots()` with warning log on init error
    - _Requirements: 5.1, 5.2_

  - [ ] 2.2 Create `services/collector/store_nop.go` with build tag `!gcp`
    - Implement `initSnapshots(_ context.Context, _ BigQueryExporterConfig, logger *slog.Logger) store.SnapshotRepo`
    - Log info "snapshot export disabled (build without -tags gcp)" and return `store.NopSnapshots()`
    - _Requirements: 5.2_

- [ ] 3. Implement Collect RPC handler and export goroutine
  - [ ] 3.1 Add exporter fields and initialization to `services/collector/server.go`
    - Add `snapshots store.SnapshotRepo` and `exportCh chan store.Snapshot` (buffer 64) fields to `UnifiedCollectorServer`
    - Update `NewUnifiedCollectorServer` to accept `store.SnapshotRepo`, initialize `exportCh`
    - Add `enqueueExport(snap store.Snapshot) bool` method using non-blocking `select` with `default` case
    - Add `runExporter(ctx context.Context)` method that drains the channel and calls `RegistrarSnapshot`, with shutdown drain loop
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ] 3.2 Implement the `Collect` method on `UnifiedCollectorServer`
    - Switch on `oneof target`: keyword → call search via pipeline (same as Fetch), shop_id → call fetch shop (same as FetchShop)
    - Validate: empty keyword → `codes.InvalidArgument`; shop_id == 0 → `codes.InvalidArgument`; neither set → `codes.InvalidArgument`
    - On successful search with products > 0: build `store.Snapshot` with keyword literal (or `strconv.FormatInt(shop_id, 10)` for shop), estrategia `"coleta-agendada"`, timestamp `time.Now().UTC()`, items mapped with position (1-based)
    - Call `enqueueExport` and set `persisted` in response based on return value
    - On search with 0 products: return empty list, `persisted=false`, do NOT enqueue
    - Return `CollectResponse` with products, total_found, fetched_at (RFC3339)
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 3.1, 3.2, 3.3, 3.4, 3.5_

  - [ ] 3.3 Wire exporter into `services/collector/main.go`
    - Call `initSnapshots(ctx, cfg.Exporters.BigQuery, logger)` at startup
    - Pass `SnapshotRepo` to `NewUnifiedCollectorServer`
    - Start `runExporter` goroutine with cancellable context
    - On shutdown: cancel context, wait for exporter goroutine to complete drain
    - _Requirements: 4.4, 4.5, 5.1, 5.2_

  - [ ]* 3.4 Write unit tests for Collect RPC in `services/collector/server_test.go`
    - Test keyword target → returns products, calls enqueue
    - Test shop_id target → returns products, calls enqueue
    - Test empty keyword → `codes.InvalidArgument`
    - Test shop_id == 0 → `codes.InvalidArgument`
    - Test neither set → `codes.InvalidArgument`
    - Test 0 products → empty response, no enqueue
    - Test buffer full → `persisted=false`, products still returned
    - Inject `store.NopSnapshots()` for exporter
    - _Requirements: 2.1, 2.5, 2.6, 2.7, 3.1, 3.3, 3.4, 4.2, 4.3, 6.3_

- [ ] 4. Checkpoint - Ensure Collector builds and tests pass
  - Ensure `go build ./services/collector/` and `go build -tags gcp ./services/collector/` both compile
  - Run `go test ./services/collector/...` — ensure all tests pass, ask the user if questions arise.

- [ ] 5. Migrate Scheduler to use Collect RPC
  - [ ] 5.1 Replace Fetch/FetchShop calls with Collect in `services/scheduler/jobs.go`
    - In `executeKeywordSearch`: replace `s.collector.Fetch(...)` with `s.collector.Collect(...)` using `CollectRequest` with keyword target
    - In `executeShopCollection`: replace `s.collector.FetchShop(...)` with `s.collector.Collect(...)` using `CollectRequest` with shop_id target
    - In `executeShopCollection` filtered keywords path: replace `s.collector.Fetch(...)` with `s.collector.Collect(...)` using keyword target
    - Remove `persistSnapshot` method entirely
    - Remove `"github.com/fmarquesfilho/garimpo/internal/store"` import from `jobs.go`
    - On gRPC error from Collect: log error with job name and keyword/shop_id, skip alert enqueueing, continue
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [ ] 5.2 Remove store dependency from `services/scheduler/server.go`
    - Remove `snapshots store.SnapshotRepo` field from `SchedulerServer` struct
    - Remove `"github.com/fmarquesfilho/garimpo/internal/store"` import
    - Remove `snapshots: initSnapshotStore(ctx, logger)` from `NewSchedulerServer`
    - _Requirements: 7.3_

  - [ ] 5.3 Delete `services/scheduler/store_gcp.go` and `services/scheduler/store_nogcp.go`
    - These files are no longer needed since the Scheduler no longer persists snapshots
    - _Requirements: 7.3_

  - [ ] 5.4 Add `shouldNotDependsOn` rule for Scheduler → `internal/store` in `arch-go.yml`
    - Add `"github.com/fmarquesfilho/garimpo/internal/store"` to the Scheduler's `shouldNotDependsOn.internal` list
    - _Requirements: 7.6_

  - [ ]* 5.5 Write unit tests for Scheduler Collect integration in `services/scheduler/server_test.go`
    - Test keyword job calls Collect with keyword target
    - Test shop job calls Collect with shop_id target
    - Test Collect error → log + skip alert + continue
    - Verify no import of `internal/store` in scheduler package
    - _Requirements: 7.1, 7.2, 7.4_

- [ ] 6. Checkpoint - Ensure Scheduler builds and arch-go passes
  - Ensure `go build ./services/scheduler/` compiles without `-tags gcp`
  - Run `go test ./services/scheduler/...` — ensure all tests pass
  - Run `arch-go` — ensure zero violations, ask the user if questions arise.

- [ ] 7. Update Dockerfiles
  - [ ] 7.1 Add `-tags gcp` to Collector Dockerfile build command
    - Change `go build -ldflags="-s -w" -o /out/garimpei-collector ./services/collector` to `go build -tags gcp -ldflags="-s -w" -o /out/garimpei-collector ./services/collector`
    - _Requirements: 5.1_

  - [ ] 7.2 Remove `-tags gcp` from Scheduler Dockerfile build command
    - Change `go build -tags gcp -ldflags="-s -w" -o /out/scheduler ./services/scheduler` to `go build -ldflags="-s -w" -o /out/scheduler ./services/scheduler`
    - _Requirements: 7.3_

- [ ] 8. Final checkpoint - Ensure full build and tests pass
  - Run `go build ./...` to verify all packages compile
  - Run `go test ./...` to ensure all tests pass
  - Run `arch-go` to validate architecture rules
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Proto changes (task 1) must run first since `buf generate` produces Go + C# stubs used everywhere
- Collector exporter infrastructure (task 2) is independent of the Collect method but must exist before wiring in main.go
- Scheduler migration (task 5) depends on proto stubs existing (task 1) but is independent of Collector implementation — only the proto contract matters
- The `persistSnapshot` method in the Scheduler moves conceptually to the Collector's background goroutine
- `NopSnapshots` is already implemented in `internal/store` — no new code needed there
- The design specifies `persisted=true` even with NopSnapshots (snapshot accepted into channel) but no data is written

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2"] },
    { "id": 2, "tasks": ["2.1", "2.2"] },
    { "id": 3, "tasks": ["3.1"] },
    { "id": 4, "tasks": ["3.2", "3.3"] },
    { "id": 5, "tasks": ["3.4", "5.1"] },
    { "id": 6, "tasks": ["5.2"] },
    { "id": 7, "tasks": ["5.3", "5.4"] },
    { "id": 8, "tasks": ["5.5", "7.1", "7.2"] }
  ]
}
```
