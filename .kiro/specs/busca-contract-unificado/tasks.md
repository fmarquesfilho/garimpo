# Implementation Plan

## Overview

Implementação do contrato canônico unificado BuscaContract, que elimina a fragmentação de identificação de buscas entre serviços. Inclui schema JSON, derivação cross-language, evolução de proto/BigQuery, fix do Frontend e enforcement no CI.

## Tasks

- [x] 1. Create `contracts/schemas/busca-contract.json` with JSON Schema Draft 2020-12 defining BuscaContract: id (string, minLength:1), tipo (enum: keyword, loja, loja-multi, categoria, mista), keywords (string[]), shop_ids (int64[]), shop_names (object), categorias (string[]), collection_keys (string[], minItems:1), cron (string|null), marketplaces (string[], minItems:1), owner_uid (string, minLength:1), comissao_min (number 0-1|null), vendas_min (integer >=0|null), fontes (string[], minItems:1). Add cross-field validation rules via if/then: tipo=keyword requires keywords minItems:1, tipo=loja requires shop_ids minItems:1, tipo=mista requires both non-empty.
- [x] 2. Add boundary entry in `contracts/registry.yaml` referencing `busca-contract.json` with all services (frontend, csharp-api, scheduler, collector, analyzer) as consumers. Validate with `mise run check:service-contracts`.
- [x] 3. Update `fixtures/buscas.json` to conform to new schema: add `collection_keys` field (derived), ensure `marketplaces` is array (not string), add `owner_uid` to all fixtures. Add edge-case fixtures: empty keywords with shop_ids, multiple keywords with mixed casing, shop_ids with keyword-like values.
- [x] 4. Add `inference` section to `rules/busca-rules.json` with declarative rules: shop_ids→marketplaces derivation, shop_ids non-empty→fontes includes "lojas", union logic for marketplaces, marketplace guard (cannot deactivate marketplace with active shops). Add `derivation` section specifying how collection_keys and tipo are computed. Update `rules/busca-rules.schema.json` and verify with `mise run check:rules-schema`.
- [x] 5. Implement `deriveCollectionKeys` in Go (`internal/busca/collection_keys.go`): shop_ids as strings + keywords lowercase/trimmed, sorted, deduped. Write test `internal/busca/collection_keys_test.go` loading `fixtures/buscas.json`.
- [x] 6. Implement `deriveCollectionKeys` in Python (`services/analyzer/collection_keys.py`): identical logic. Write test `services/analyzer/test_collection_keys.py` with pytest parametrize against same fixtures.
- [x] 7. Implement `deriveCollectionKeys` in C# (`src/Garimpei.Domain/CollectionKeys.cs`): identical logic. Write test `src/Garimpei.Api.Tests/CollectionKeysTests.cs` with xUnit Theory against same fixtures.
- [x] 8. Implement `deriveCollectionKeys` in TypeScript (`web/src/lib/collection-keys.js`): identical logic. Write test `web/src/lib/collection-keys.test.js` with vitest against same fixtures. Verify all 4 implementations produce identical output.
- [x] 9. Update BigQuery DDL/schema to add `busca_id STRING NOT NULL` column to `snapshots` table. Verify `mise run db:reset -- --prod --bq` handles the new column (drops and recreates). Update `services/analyzer/mock_data.py` to include `busca_id`.
- [x] 10. Add `string busca_id = 7` field to `CollectRequest` in `protos/collector/v1/collector.proto`. Regenerate proto stubs (`mise run proto:gen`). Update Collector Collect RPC handler to reject empty busca_id (InvalidArgument). Update Collector BigQuery writer to persist busca_id. Add tests for rejection of empty busca_id.
- [x] 11. Update `services/scheduler/jobs.go`: read `busca_id` from params in `executeShopCollection` and `executeKeywordSearch`, pass it to Collector.Collect via proto field. Add validation in `dispatchJob` — reject if params["busca_id"] is empty for collection jobs. Update Scheduler tests. Verify `mise run test:go`.
- [x] 12. Update `src/Garimpei.Api/Endpoints/SchedulerJobs.cs` `BuildRequest`: add `busca_id` (busca.Id) and `collection_keys` (derived via CollectionKeys.Derive, comma-joined) to req.Params. Support type="mixed" when both shop_ids and keywords present. Add unit test verifying BuildRequest always includes busca_id and collection_keys. Run `dotnet test`.
- [x] 13. Refactor `services/analyzer/routes/novidades.py`: replace `WHERE keyword LIKE @busca_id` with `WHERE busca_id = @busca_id` (parameterized exact match). Remove `f"%{busca_id}%"` wrapping. Apply same fix to `routes/quedas.py` and audit all other routes for LIKE usage on busca identification. Add Python tests verifying no LIKE in queries.
- [x] 14. Fix `web/src/lib/descobrir.js` `carregarOportunidades`: change `const buscaId = b.shop_ids?.[0]?.toString() || b.keywords?.[0] || b.id` to `const buscaId = b.id` (always use UUID). Verify all calls to `buscarNovidades` pass busca UUID. Add vitest test confirming buscaId resolution always returns busca.id. Run `npm run test` in `web/`.
- [x] 15. Create cross-language fixture test script (`.mise/tasks/check/fixtures-crosslang`) that runs Go, Python, C#, and JS derive functions against `fixtures/buscas.json` and diffs outputs. Add `mise run check:fixtures-contract` validation that fixtures conform to `busca-contract.json`. Add checks to `.github/workflows/ci.yml`. Run full CI locally.
- [x] 16. Run `mise run db:reset -- --prod --bq` to drop/recreate snapshots table with busca_id NOT NULL. Verify Scheduler cron passes busca_id correctly (manual trigger or logs). Verify Frontend "Novos" for busca tipo "loja" returns products. Verify "Quedas" for tipo "mista" returns variations. Confirm `grep -r "LIKE" services/analyzer/routes/` returns zero hits for busca identification.
- [x] 17. Audit and remove legacy endpoint patterns that violate the BuscaContract: (a) Remove `Busca.Keyword` (singular) field from entity — use only `Keywords[]` + `Id` for identity. Create EF Core migration. (b) Migrate `Busca.Marketplaces` from `string` (comma-separated) to `string[]` (jsonb). Create EF Core migration. (c) Refactor `POST /api/buscas` upsert logic from matching by `Keyword` string to matching by normalized identity fields (`buscaDuplicada.camposIdentidade`). (d) Refactor `POST /api/buscas?remover=true` to accept `id` (UUID) instead of keyword match. (e) Remove `keyword` field from `GET /api/lojas` response — use only `keywords[]`. (f) Unify `/api/lojas/novidades` and `/api/v2/curadoria/novos` into a single endpoint accepting `busca_id` UUID. Deprecate the duplicate. (g) Add optional `busca_id` param to `/api/v2/curadoria/quedas` for per-busca filtering. (h) Update `GET /api/buscas` response to stop including legacy `keyword` fallback logic in `shopNames` computation. (i) Run `mise run check:api-contract` to verify no frontend route references removed endpoints. (j) Run full test suite (`mise run test` + `mise run checks`).

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": [1] },
    { "wave": 2, "tasks": [2, 4, 9] },
    { "wave": 3, "tasks": [3, 10] },
    { "wave": 4, "tasks": [5, 6, 7, 8, 11] },
    { "wave": 5, "tasks": [12, 13, 14] },
    { "wave": 6, "tasks": [15] },
    { "wave": 7, "tasks": [16, 17] }
  ]
}
```

## Endpoints Legados Identificados (Non-Compliance)

Os seguintes endpoints/patterns NÃO aderem ao contrato unificado e devem ser eliminados ou migrados:

| Endpoint/Pattern | Problema | Ação |
|-----------------|----------|------|
| `GET /api/lojas` | Retorna `keyword` (campo legado singular) junto com `keywords[]`. Frontend não deveria usar `keyword`. | Remover campo `keyword` da response — usar apenas `keywords[]` |
| `GET /api/lojas/novidades?busca_id=X` | Aceita shop_id ou keyword como `busca_id` e faz proxy para Analyzer com LIKE. Frontend envia shop_id. | Migrar para aceitar UUID da busca. Proxy deve enviar UUID ao Analyzer. |
| `POST /api/buscas` | Faz upsert por `keyword` string (campo legado). Aceita `Marketplaces` como string comma-separated. | Migrar para upsert por UUID. `Marketplaces` deve ser `string[]` (array). |
| `POST /api/buscas?remover=true` | Busca para desativar por `Keyword == X`. | Migrar para desativar por `id` (UUID). |
| `Busca.Marketplaces` (entity) | É `string` comma-separated ("shopee,amazon"). | Migrar para `string[]` no PostgreSQL (jsonb). |
| `Busca.Keyword` (entity) | Campo legado singular. Usado como fallback/upsert key. | Deprecar — usar apenas `Keywords[]` e `Id` para identidade. |
| `GET /api/v2/curadoria/novos?busca_id=X` | Mesmo problema que `/api/lojas/novidades` — aceita qualquer string como busca_id, faz LIKE. | Remover ou unificar com `/api/lojas/novidades`. Migrar para UUID exact match. |
| `GET /api/candidatos?keyword=X&shop_ids=Y` | Usa `Fetch` (sem persistência). Frontend usa para curadoria em tempo real. Não relacionado a snapshots, mas é uma interface legada com params ad-hoc. | Manter (é busca live, não histórico). Mas adicionar `busca_id` opcional para correlação com traces. |
| `/api/v2/curadoria/quedas` | Não filtra por busca — retorna quedas globais do tenant. Inconsistente com `/novidades` que filtra por busca. | Unificar: aceitar `busca_id` opcional. Se fornecido, filtra por busca. Se não, retorna global. |

## Notes

- Nenhuma dependência npm/pip/go nova é necessária (conforme design)
- Task 16 (migration) é destrutiva (db:reset) — validar apenas localmente ou com confirmação explícita para prod
- Tasks 5-8 (derivation) são independentes entre si e podem ser paralelizadas
- O proto evolution (task 10) é aditivo — não quebra backward compat do wire format
- **`Busca.Keyword`** (singular, string) é o principal resquício legado — era a identidade antes de UUIDs. Deve ser removido do entity após migração.
- **`Busca.Marketplaces`** como string comma-separated viola o contrato (deveria ser `string[]`). Requer migration EF Core + atualização do entity.
