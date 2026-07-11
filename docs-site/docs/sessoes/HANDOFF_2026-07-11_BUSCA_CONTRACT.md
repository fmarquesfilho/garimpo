# Handoff — Busca Contract Unificado (2026-07-11)

> Próxima sessão: implementar as 17 tasks da spec `busca-contract-unificado`.
> Branch: main (push direto, sem PR — MVP com 2 usuários).
> Spec completa em `.kiro/specs/busca-contract-unificado/`.

## Estado atual

- **298 unit tests passando** (frontend vitest)
- **72 C# tests passando** (arch + integration)
- **75 E2E checks passando** (cross-service, 0 warnings)
- **8 drift checks passando**
- OTel em produção, traces correlacionados
- ADR-0029 publicada no docs-site
- Manual do desenvolvedor criado (`docs/10-manual-do-desenvolvedor.md`)
- Branches limpas (apenas `main`)
- `mise run test:e2e-novos` — novo teste que valida pipeline Novos em ~15s

## O problema que a spec resolve

O tipo "Busca" está fragmentado entre serviços. Cada um interpreta de forma ad-hoc:
- **Frontend** envia `shop_id || keyword || uuid` como `busca_id` para o Analyzer
- **Analyzer** faz `WHERE keyword LIKE '%X%'` — falha se Collector indexou diferente
- **Resultado**: "Novos" e "Quedas" mostram 0 na página principal para buscas com loja

Além disso, regras de inferência não existem:
- Adicionar loja Shopee NÃO ativa automaticamente o marketplace Shopee
- Isso causa buscas "duplicadas" porque a engine não reconhece a equivalência

## Decisão arquitetural

- **BuscaContract**: tipo canônico em `contracts/schemas/busca-contract.json`
- **Inference rules**: declarativas em `rules/busca-rules.json` (shop → marketplace, shop → fontes)
- **busca_id**: UUID estável no BigQuery (exact match, zero LIKE)
- **Protovalidate + CEL**: validação cross-field no proto
- **Cross-language fixtures**: mesmos 5 exemplos testados em Go/Python/C#/TS
- **Zero backward compat**: `db:reset`, dados antigos descartados

## Tasks (7 waves, 17 tasks)

| Wave | Tasks | Descrição |
|------|-------|-----------|
| 1 | 1 | JSON Schema `busca-contract.json` |
| 2 | 2, 4, 9 | Registry boundary, inference rules, BigQuery DDL |
| 3 | 3, 10 | Fixtures, proto `busca_id` no CollectRequest |
| 4 | 5-8, 11 | `deriveCollectionKeys` em 4 linguagens + Scheduler |
| 5 | 12-14 | C# BuildRequest, Analyzer exact match, Frontend fix |
| 6 | 15 | CI cross-language fixture diff |
| 7 | 16, 17 | Migration + E2E + remoção de endpoints legados |

## Endpoints legados a eliminar (task 17)

| Endpoint | Problema |
|----------|----------|
| `Busca.Keyword` (entity, singular) | Campo legado — upsert key e fallback |
| `Busca.Marketplaces` (string comma-separated) | Deveria ser `string[]` jsonb |
| `POST /api/buscas` upsert por keyword | Migrar para UUID/identity fields |
| `POST /api/buscas?remover=true` por keyword | Migrar para `id` |
| `GET /api/lojas` campo `keyword` na response | Remover — usar apenas `keywords[]` |
| `GET /api/lojas/novidades` LIKE | Migrar para UUID exact match |
| `GET /api/v2/curadoria/novos` duplicado | Unificar com `/api/lojas/novidades` |
| `GET /api/v2/curadoria/quedas` sem busca_id | Adicionar param opcional |

## Arquivos-chave

| Arquivo | O quê |
|---------|-------|
| `.kiro/specs/busca-contract-unificado/design.md` | Design completo |
| `.kiro/specs/busca-contract-unificado/requirements.md` | 10 requirements, 35 AC |
| `.kiro/specs/busca-contract-unificado/tasks.md` | 17 tasks, 7 waves |
| `rules/busca-rules.json` | Rules declarativas (adicionar `inference`) |
| `contracts/registry.yaml` | Service contracts (adicionar boundary) |
| `protos/collector/v1/collector.proto` | Adicionar `busca_id` ao CollectRequest |
| `protos/scheduler/v1/scheduler.proto` | Já tem TriggerJob, SetSchedule |
| `services/scheduler/jobs.go` | Passar busca_id ao Collector |
| `services/analyzer/routes/novidades.py` | LIKE → exact match |
| `services/analyzer/routes/quedas.py` | LIKE → exact match |
| `src/Garimpei.Api/Endpoints/SchedulerJobs.cs` | BuildRequest com busca_id |
| `src/Garimpei.Api/Endpoints/BuscasEndpoints.cs` | Upsert legado por keyword |
| `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` | /novidades proxy com LIKE |
| `src/Garimpei.Domain/Entities/Busca.cs` | Entity com campos legados |
| `web/src/lib/descobrir.js` | Frontend fix (sempre enviar busca.id) |
| `fixtures/buscas.json` | Atualizar com collection_keys + busca_id |

## Como verificar

```bash
# Spec
cat .kiro/specs/busca-contract-unificado/tasks.md

# Testes rápidos (durante implementação)
cd web && bunx vitest run
dotnet test src/Garimpei.sln
go test ./...
mise run lint:python

# Drift checks
mise run checks

# E2E Novos (confirma o fix)
mise run test:e2e-novos

# E2E completo
mise run test:e2e-full-pipeline
```

## Steering rules ativas

- `git.md` — nunca `--no-verify`, nunca push automático
- `ci.md` — nunca E2E real no CI, deploy conservador
- `dependencies.md` — sem deps com >3 meses sem release
