# ADR-0031: Cache Layer L1 + L2 — Edge + Sidecar

**Data:** 2026-07-12
**Status:** Aceito
**Contexto:** Frontend faz polling a cada 30s; dados mudam a cada 8h. Sem cache, cada poll atinge o Collector que chama APIs externas.

## Decisão

Implementar cache em duas camadas:

### L1 — Cloudflare Workers Cache (Edge)
- Cache HTTP nativo no Worker existente (`garimpei-proxy`)
- TTL 5 minutos, invalidação por Cache-Tag (`busca:{busca_id}`)
- Absorve ~90% dos reads (30s poll / 300s TTL)
- Zero custo adicional (Workers Cache é free tier)
- Headers: `X-Cache-Status: HIT|MISS|BYPASS`, `Cache-Tag: busca:{id}`

### L2 — Go Cache Sidecar (Cloud Run)
- Serviço gRPC (porta 50055) no mesmo pod
- In-memory LRU (256 MB, ~50k entries)
- TTL 30 minutos
- Singleflight protege contra thundering herd
- Valida BuscaContract antes de armazenar

### Invalidação
- Collector detecta divergência via hash SHA-256 após cada `Collect`
- Se diverge: Invalidate L2 (gRPC local, <1ms) → Purge L1 (Cloudflare API, <100ms)
- Retry: 1x para L1. Se falha 2x, skip (TTL expira naturalmente)

### Circuit Breaker (C# API)
- 3 falhas → OPEN → fallback direto ao Collector
- 10s → HALF_OPEN → tenta 1 request → CLOSED se OK
- Header `X-Cache-Source: l2-hit|l2-miss|l2-bypass`

## Consequências

- **Positivo:** Redução de ~95% da carga no Collector; latência p50 de ~500ms para ~5ms
- **Positivo:** Sem dependências novas (LRU stdlib, singleflight já transitiva)
- **Positivo:** Graceful degradation — se cache falha, sistema funciona como antes
- **Negativo:** Mais um container no pod (512Mi RAM extra)
- **Negativo:** Dados podem ficar stale por até 5min (L1) ou 30min (L2) sem invalidação

## Arquivos Principais

| Arquivo | Propósito |
|---------|-----------|
| `protos/cache/v1/cache.proto` | CacheService (Get, Invalidate, Healthz) |
| `services/cache/` | Go sidecar (server, lru, config) |
| `services/collector/divergence.go` | Divergence detection + invalidation |
| `src/Garimpei.Infrastructure/Sources/CacheCircuitBreaker.cs` | Circuit breaker |
| `cloudflare-worker/worker.js` | L1 caching logic |
| `deploy/cloud-run-deploy-now.yaml` | Sidecar deployment |

## Referências

- Spec completa: `.kiro/specs/cache-layer/`
- Design: `.kiro/specs/cache-layer/design.md` (12 correctness properties)
- ADR-0029 (Dashboard Realtime) definiu o polling a cada 30s
- ADR-0030 (BuscaContract) define collection_keys derivadas
