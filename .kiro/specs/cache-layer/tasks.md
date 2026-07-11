# Implementation Plan

## Overview

Implementação do cache em duas camadas (L1 edge + L2 sidecar) para o Garimpei. O Cache Sidecar (Go) fica entre a C# API e o Collector, validando BuscaContract e protegendo contra thundering herd. O Cloudflare Workers Cache (L1) absorve ~90% das leituras repetitivas na edge sem atingir Cloud Run.

## Tasks

- [x] 1. Create `protos/cache/v1/cache.proto` with CacheService definition (Get, Invalidate, Healthz). Import `collector/v1/collector.proto` for Product and Marketplace types. Run `mise run proto:generate` to generate Go and C# stubs.
- [x] 2. Create `services/cache/` directory with Go sidecar implementation: `main.go` (gRPC server startup, health port, config from env), `server.go` (CacheService implementation with LRU store, reverse index, singleflight, BuscaContract validation), `lru.go` (LRU cache with size-based eviction, TTL, thread-safe via sync.RWMutex), `config.go` (env vars: CACHE_MAX_BYTES, CACHE_TTL_SECONDS, COLLECTOR_GRPC_ADDRESS).
- [x] 3. Write unit tests for Cache Sidecar: `services/cache/server_test.go` (TestGet_CacheHit, TestGet_CacheMiss, TestGet_ValidationFailure, TestInvalidate_RemovesByBuscaId, TestInvalidate_Idempotent, TestSingleflight_Coalesces), `services/cache/lru_test.go` (TestLRU_PutGet, TestLRU_Eviction, TestLRU_TTLExpiry, TestLRU_SizeLimit).
- [x] 4. Modify `services/collector/server.go` Collect method: after successful persist, compute SHA-256 hash of collected products, call `CacheService.Invalidate` via local gRPC if divergence detected, call Cloudflare Purge API with tag `busca:{busca_id}` (with 1 retry). Add env vars `CACHE_GRPC_ADDRESS`, `CF_PURGE_TOKEN`, `CF_ZONE_ID`. Add tests in `services/collector/divergence_test.go`.
- [x] 5. Modify C# API to call CacheService.Get instead of CollectorService.Fetch/FetchShop for read operations: update `CuradoriaEndpoints.cs` (ranking, ranking/shop), update `src/Garimpei.Api/Endpoints/CandidatosEndpoints.cs`. Add gRPC client registration for `CacheService.CacheServiceClient`. Implement circuit breaker (3 failures → open, 10s → half-open). Add `X-Cache-Source` header. Add tests.
- [x] 6. Update `cloudflare-worker/worker.js`: add caching logic for GET requests on cacheable routes (`/api/v2/curadoria/ranking`, `/api/candidatos`, `/api/lojas/novidades`). Set `Cache-Control: public, max-age=300`, `Cache-Tag: busca:{busca_id}`, `X-Cache-Status: HIT|MISS|BYPASS`. Non-cacheable routes bypass.
- [x] 7. Create `services/cache/Dockerfile`. Add cache-sidecar to `deploy/cloud-run-deploy-now.yaml` (port 50055, 512Mi memory, startup probe). Add to CI pipeline (`.github/workflows/ci.yml` build step). Update `contracts/registry.yaml` with cache-sidecar service and boundaries.
- [x] 8. Add `mise run test:integration:cache` task that starts cache sidecar + collector mock, validates Get miss→fetch→hit cycle, validates Invalidate→next Get refetches. Run full test suite (`mise run test:go`, `dotnet test`, `bunx vitest run`). Verify `mise run check:service-contracts` passes.
- [x] 9. Run E2E validation: `mise run test:e2e-novos` confirms pipeline works with cache layer. Verify cache headers in responses (`X-Cache-Source`, `X-Cache-Status`). Confirm `grep -r "CACHE_GRPC_ADDRESS" deploy/` shows correct wiring. Document in ADR-0031.

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": [1] },
    { "wave": 2, "tasks": [2, 6] },
    { "wave": 3, "tasks": [3, 4, 5] },
    { "wave": 4, "tasks": [7] },
    { "wave": 5, "tasks": [8, 9] }
  ]
}
```

## Notes

- Nenhuma dependência externa nova (LRU é stdlib, singleflight já está no go.mod transitivamente)
- Cache keys incluem owner_uid como prefixo para tenant isolation
- Sidecar comunica via localhost gRPC (sem network hop)
- Workers Cache é feature nativa do Cloudflare (não requer pacote)
- Purge API usa token com escopo mínimo (Zone:Cache Purge only)
- Se o cache sidecar falha, C# API faz fallback transparente para Collector
- TTL L1 = 5 minutos, TTL L2 = 30 minutos
- Invalidação é bilateral: L2 (gRPC local) + L1 (Cloudflare Purge API por tag)
