# Implementation Plan: Coupon Monitoring

## Overview

This plan implements the coupon monitoring feature across 14 tasks, progressing from data models → collection infrastructure → detection → alerting → integration. Tasks follow the dependency graph to allow incremental delivery and testing at each stage.

## Tasks

- [x] 1. Create coupon proto definition and domain models (Go + C#)
  - Create `protos/coupon/v1/coupon.proto` with CouponCollectorService, FetchCouponsRequest/Response, CouponProto, DiscountType enum
  - Run `buf generate` to produce Go and C# code
  - Create `internal/domain/coupon.go` with Coupon struct, DetectionStatus constants, CouponDetection struct
  - Create `src/Garimpei.Domain/Interfaces/ICouponSource.cs` with ICouponSource, CouponSourceResult, CouponCandidate
  - Verify: `go build ./...` and `dotnet build` pass
  - **Implements:** R1-AC3, R2-AC3, R8-AC1

- [x] 2. Create Go CouponSource interface and registry
  - Create `internal/couponsource/source.go` with CouponSource interface, FetchConfig, CouponSourceFactory, SourceConfig
  - Create `internal/couponsource/registry.go` with thread-safe Registry and DefaultRegistry
  - Register Shopee, Amazon, ML factory stubs in init()
  - Verify: `go build ./...` passes
  - **Implements:** R1-AC1, R2-AC1, R3-AC1

- [x] 3. Implement Shopee coupon adapter (Go)
  - Create `internal/couponsource/shopee_adapter.go` implementing CouponSource
  - Fetch via Shopee productOfferV2 with coupon fields, paginate at 500/page with 200ms delay
  - HMAC-SHA256 auth, 30s timeout, 2 retries with 5s backoff
  - Mark coupons with end_time in past as "expired"
  - Write unit test with mock HTTP server
  - Verify: `go test ./internal/couponsource/...` passes
  - **Implements:** R1-AC1, R1-AC2, R1-AC4, R1-AC5, R1-AC6, R1-AC7

- [x] 4. Implement Amazon coupon adapter (Go)
  - Create `internal/couponsource/amazon_adapter.go` implementing CouponSource
  - Normalize Amazon offer fields to unified Coupon model
  - Rate limit 1 req/s, HTTP 429 → 60s wait + 2 retries, HTTP 5xx → 5s backoff + 2 retries
  - Skip silently if no credentials
  - Write unit test with mock HTTP server
  - Verify: `go test ./internal/couponsource/...` passes
  - **Implements:** R2-AC1, R2-AC2, R2-AC3, R2-AC4, R2-AC5, R2-AC6, R2-AC7, R2-AC8, R2-AC9

- [x] 5. Create BigQuery coupon schema and Go writer
  - Create `deploy/bigquery_coupon_schema.sql` (coupon_snapshots partitioned by collected_at, 90-day expiration)
  - Create `internal/couponsource/bqwriter.go` for append-only BigQuery inserts
  - Write integration test using BigQuery emulator
  - Verify: `go test ./internal/couponsource/...` passes
  - **Implements:** R8-AC1, R8-AC2, R8-AC3, R8-AC5

- [x] 6. Create coupon-collector gRPC server
  - Create `services/coupon-collector/main.go` (reads MARKETPLACE env, creates adapter from registry)
  - Create `services/coupon-collector/server.go` implementing FetchCoupons RPC (delegates to CouponSource, writes BigQuery)
  - Create `services/coupon-collector/Dockerfile`
  - Write server test with fake source
  - Verify: `go build ./services/coupon-collector/` and `go test` pass
  - **Implements:** R1-AC1, R1-AC3, R2-AC1, R3-AC1

- [ ] 7. Extend scheduler with coupon collection jobs
  - Add CouponCollectorService gRPC client to SchedulerServer (3 marketplace addresses from env)
  - Add executeCouponCollectionJob: sequential Shopee→Amazon→ML, 3min timeout, POST to analyzer on success
  - Support job type "coupon_collection" with default cron "0 */2 * * *"
  - Skip marketplaces without credentials, continue on failure
  - Write test for sequential execution and skip behavior
  - Verify: `go test ./services/scheduler/` passes
  - **Implements:** R4-AC1, R4-AC2, R4-AC3, R4-AC4, R4-AC5

- [x] 8. Create C# domain entities and EF Core migration
  - Create CouponAlertRule entity (IOwnedEntity: DiscountType, MinDiscount, Marketplaces, Categories, Channel, IsActive)
  - Create CouponAlertHistory entity (IOwnedEntity: CouponId, AlertRuleId, AlertedDiscountValue, AlertedAt, ExpiresAt)
  - Add DbSets and entity configs with multi-tenant query filter to AppDbContext
  - Generate migration AddCouponAlertRulesAndHistory
  - Verify: `dotnet build` and `dotnet test` pass
  - **Implements:** R6-AC2, R9-AC3

- [x] 9. Implement coupon alert rules CRUD endpoints (C# API)
  - Create CouponRulesEndpoints.cs with POST/GET/PUT/DELETE/PATCH for /api/v2/cupons/regras
  - Validate: max 20 active rules, max 10 categories, valid discount range, at least one marketplace
  - Return HTTP 409 on max rules exceeded
  - Reset dedup state on rule update
  - Register endpoints in Program.cs
  - Verify: `dotnet build` and `dotnet test` pass
  - **Implements:** R6-AC1, R6-AC5, R6-AC6, R6-AC9

- [ ] 10. Implement coupon detection in Python analyzer
  - Add POST /detect-coupons endpoint to services/analyzer/ (FastAPI)
  - BigQuery diff query: compare current vs previous snapshot, classify newly_discovered/modified/expired_or_removed
  - Safety: skip if current snapshot has 0 rows, log warning
  - On error: log failure, discard partial results
  - POST detection results to C# API POST /internal/coupon-alerts/evaluate
  - Write pytest for detection logic
  - Verify: pytest passes
  - **Implements:** R5-AC1, R5-AC2, R5-AC3, R5-AC4, R5-AC5, R5-AC6, R5-AC7

- [x] 11. Implement alert matcher and deduplication service (C# API)
  - Create POST /internal/coupon-alerts/evaluate endpoint
  - Create CouponDeduplicationService with ShouldAlertAsync (24h window, discount increase bypass)
  - Match: discount_type match, discount_value >= threshold, marketplace match, category intersection
  - Consolidate: group by (owner_uid, channel) for single message per channel
  - Record alert in CouponAlertHistory, reset dedup on rule edit
  - Dispatch via Alerter gRPC SendCouponAlert
  - Write unit tests for matching, dedup, consolidation
  - Verify: `dotnet build` and `dotnet test` pass
  - **Implements:** R6-AC3, R6-AC4, R6-AC7, R6-AC8, R6-AC10, R9-AC1, R9-AC2, R9-AC4, R9-AC5

- [ ] 12. Extend alerter with coupon notification formatting
  - Add SendCouponAlert RPC to protos/alerter/v1/alerter.proto, run buf generate
  - Implement: Telegram Markdown (bold discount, categories, link), WhatsApp plain text (emojis)
  - Prepend "⚡ Expira em breve!" when end_time < 24h
  - Include tenant's affiliate tag in coupon URL
  - Retry once after 30s on failure, log on second failure
  - Write test for message formatting
  - Verify: `go test ./services/alerter/` passes
  - **Implements:** R7-AC1, R7-AC2, R7-AC3, R7-AC4, R7-AC5, R7-AC6

- [ ] 13. Docker Compose and integration wiring
  - Add coupon-collector-shopee/amazon/ml services to docker-compose.yml (ports 50061-50063)
  - Add scheduler env vars for coupon collector addresses and analyzer URL
  - Verify: `docker compose config` validates
  - Smoke test: scheduler triggers → collection → BigQuery → detection → alert evaluation
  - **Implements:** R4-AC2, R4-AC3

- [ ] 14. Coupon listing and analytics endpoints (C# API)
  - GET /api/v2/cupons — list active coupons from latest BigQuery snapshot (marketplace/category filter)
  - GET /api/v2/cupons/historico — query coupon history (time range, marketplace, discount range)
  - Register endpoints in Program.cs
  - Verify: `dotnet build` passes
  - **Implements:** R8-AC4

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": [1] },
    { "wave": 2, "tasks": [2, 8] },
    { "wave": 3, "tasks": [3, 4, 5, 9] },
    { "wave": 4, "tasks": [6, 14] },
    { "wave": 5, "tasks": [7, 10] },
    { "wave": 6, "tasks": [11] },
    { "wave": 7, "tasks": [12] },
    { "wave": 8, "tasks": [13] }
  ]
}
```

```
T1 (proto + domain)
├── T2 (interface + registry)
│   ├── T3 (Shopee adapter)
│   ├── T4 (Amazon adapter)
│   └── T5 (BigQuery schema + writer)
│       └── T6 (coupon-collector server) ← depends on T3, T4, T5
│           └── T7 (scheduler extension) ← depends on T6
│               └── T10 (Python detector) ← depends on T5, T7
│                   └── T11 (alert matcher) ← depends on T8, T9, T10
│                       └── T12 (alerter formatting) ← depends on T11
│                           └── T13 (Docker + integration) ← depends on T6, T7, T12
├── T8 (C# entities + migration) ← depends on T1
│   └── T9 (CRUD endpoints) ← depends on T8
└── T14 (listing endpoints) ← depends on T5, T8
```

## Notes

- Tasks 3 and 4 can be parallelized (independent marketplace adapters)
- Tasks 8 and 9 (C# side) can be developed in parallel with Tasks 3-6 (Go side)
- Task 10 (Python) can start once Task 5 (BigQuery schema) is defined
- Task 13 (integration) is the final verification that everything works end-to-end
- Mercado Livre adapter (Task 2 stub) will be implemented as a future task when ML credentials are available
