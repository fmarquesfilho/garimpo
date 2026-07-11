# Handoff — Cache Layer L1 + L2 (2026-07-12)

> Próxima sessão: implementar as 9 tasks da spec `cache-layer`.
> Branch: main (push direto, sem PR — MVP com 2 usuários).
> Spec completa em `.kiro/specs/cache-layer/`.

## Estado atual

- **314 unit tests passando** (frontend vitest)
- **87 C# tests passando** (arch + integration)
- **12 Go packages passando** (inclui busca, collector, scheduler)
- **8 E2E checks passando** (test:e2e-novos contra produção)
- **8 drift checks passando**
- BuscaContract implementado end-to-end (ADR-0030)
- Busca.Keyword removido, Marketplaces migrado para string[] (jsonb)
- Deploy em produção funcionando
- Mileny pode testar: Novos/Quedas para buscas tipo loja/mista

## O problema que a spec resolve

O frontend faz polling a cada 30s (dashboard realtime, ADR-0029). Os dados de produto mudam a cada 8h (coleta agendada). Cada poll atinge o Collector que chama APIs externas (Shopee). Resultado:
- Carga desnecessária no Collector (~120 requests/hora por usuário, mesmo sem dados novos)
- Latência evitável para o usuário (~500ms por request vs ~5ms com cache edge)
- Custo de API marketplace desproporcional ao valor (mesmos dados servidos 120x antes de mudar)

## Decisão arquitetural

Cache em duas camadas:

### L1 — Cloudflare Workers Cache (edge)
- Cache HTTP nativo na frente do Worker existente
- TTL 5 minutos, invalidação por tag (`busca:{busca_id}`)
- Absorve ~90% dos reads (30s poll / 300s TTL = 10 hits por window)
- Zero custo (free tier)
- Purge via Cloudflare API quando Collector detecta divergência

### L2 — Go Cache Sidecar (Cloud Run)
- Serviço gRPC (porta 50055) no mesmo pod
- In-memory LRU (256 MB, ~50k entries, ~850 tenants simultâneos)
- TTL 30 minutos (stale-while-revalidate)
- Protege Collector contra thundering herd (singleflight)
- Valida BuscaContract antes de armazenar
- C# API chama CacheService.Get em vez de Collector.Fetch

### Invalidação
- Collector.Collect → hash SHA-256 → compara com L2
- Se diverge: Invalidate L2 (gRPC local, <1ms) → Purge L1 (HTTP Cloudflare API, <100ms)
- Retry: 1x para L1 purge. Se falha 2x, skip (TTL expira naturalmente)

## Tasks (5 waves, 9 tasks)

| Wave | Tasks | Descrição |
|------|-------|-----------|
| 1 | 1 | Proto `cache.v1.CacheService` |
| 2 | 2, 6 | Go Sidecar + Cloudflare Worker caching |
| 3 | 3, 4, 5 | Tests + Collector divergence + C# integration |
| 4 | 7 | Dockerfile, deploy YAML, CI, registry |
| 5 | 8, 9 | Integration tests + E2E + ADR |

## Arquivos-chave

| Arquivo | O quê |
|---------|-------|
| `.kiro/specs/cache-layer/requirements.md` | 9 requirements, 36 AC |
| `.kiro/specs/cache-layer/design.md` | Design completo (12 correctness properties) |
| `.kiro/specs/cache-layer/tasks.md` | 9 tasks, 5 waves |
| `protos/cache/v1/cache.proto` | CRIAR: CacheService (Get, Invalidate, Healthz) |
| `services/cache/` | CRIAR: Go sidecar (server, lru, config, Dockerfile) |
| `services/collector/server.go` | MODIFICAR: divergence detection + invalidate |
| `src/Garimpei.Api/Endpoints/CuradoriaEndpoints.cs` | MODIFICAR: CacheService.Get + circuit breaker |
| `cloudflare-worker/worker.js` | MODIFICAR: Cache-Control + Cache-Tag + X-Cache-Status |
| `deploy/cloud-run-deploy-now.yaml` | MODIFICAR: adicionar cache-sidecar container |
| `contracts/registry.yaml` | MODIFICAR: adicionar cache-sidecar service + boundaries |

## Como verificar

```bash
# Spec
cat .kiro/specs/cache-layer/tasks.md

# Testes rápidos (durante implementação)
go test ./services/cache/... -v
go test ./services/collector/... -v
dotnet test src/Garimpei.Tests/Garimpei.Tests.csproj
cd web && bunx vitest run

# Drift checks
mise run checks

# E2E (confirma que cache não quebra pipeline)
mise run test:e2e-novos

# Verificar cache headers
curl -s -I https://garimpei.app.br/api/v2/curadoria/ranking?keyword=serum | grep -i "x-cache"
```

## Steering rules ativas

- `git.md` — nunca `--no-verify`, nunca push automático
- `ci.md` — nunca E2E real no CI, deploy conservador
- `dependencies.md` — sem deps novas (LRU é stdlib, singleflight transitiva)
