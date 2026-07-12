# ADR-0030: BuscaContract Unificado — Contrato Canônico Cross-Service

**Data:** 2026-07-11
**Status:** Aceita
**Decisores:** Fernando
**Spec:** `.kiro/specs/busca-contract-unificado/`

## Contexto

O tipo "Busca" estava fragmentado entre 5 serviços (Frontend, C# API, Scheduler Go, Collector Go, Analyzer Python). Cada um interpretava de forma ad-hoc o que identifica uma busca e como indexar/consultar snapshots no BigQuery.

Problemas concretos:

1. **Frontend enviava `shop_id || keyword || uuid`** como `busca_id` ao Analyzer — dependendo da estrutura da busca, o valor era diferente do que o Collector havia persistido.

2. **Analyzer usava `WHERE keyword LIKE '%X%'`** — full scan, falsos positivos, e falha total quando o Collector indexou com um identificador diferente do que o Frontend enviou.

3. **Resultado visível:** "Novos" e "Quedas" mostravam 0 produtos na página principal para buscas tipo "loja" e "mista".

4. **Sem regras de inferência:** adicionar uma loja Shopee não ativava automaticamente o marketplace Shopee na busca, causando inconsistências silenciosas.

5. **Campo `Busca.Keyword` (singular, string)** era usado como chave de identidade, fallback de nome, e separador de keywords por vírgula — 3 responsabilidades numa propriedade.

## Decisão

Implementar um **contrato canônico de Busca** (`BuscaContract`) definido uma vez em `contracts/schemas/busca-contract.json`, validado no CI, e consumido por todos os serviços.

### Princípios

1. **Schema como fonte de verdade** — JSON Schema Draft 2020-12 com validação cross-field via if/then
2. **Derivação determinística** — `deriveCollectionKeys` idêntico em 4 linguagens (Go, Python, C#, JS), validado por fixtures compartilhados
3. **busca_id UUID estável** — flui end-to-end (Frontend → C# → Scheduler → Collector → BigQuery → Analyzer) sem transformação
4. **Zero LIKE** — todas as queries usam exact match parameterizado
5. **Zero backward compat** — MVP com 2 usuários, dados antigos descartados via `db:reset`

### Componentes implementados

| Componente | Descrição |
|------------|-----------|
| `contracts/schemas/busca-contract.json` | Schema canônico com validação cross-field |
| `rules/busca-rules.json` (inference + derivation) | Regras declarativas de inferência e derivação |
| `internal/busca/collection_keys.go` | DeriveCollectionKeys em Go |
| `services/analyzer/collection_keys.py` | DeriveCollectionKeys em Python |
| `src/Garimpei.Domain/CollectionKeys.cs` | CollectionKeys.Derive em C# |
| `web/src/lib/collection-keys.js` | deriveCollectionKeys em JS |
| `fixtures/buscas.json` | 8 fixtures (5 tipos + 3 edge cases) |
| `.mise/tasks/check/fixtures-crosslang` | CI: diff cross-language |
| Proto `busca_id` (field 7) | Campo obrigatório no CollectRequest |
| BigQuery `busca_id NOT NULL` | Coluna explícita na tabela snapshots |

### Write path (coleta)

```
Frontend → POST /api/buscas
  → C# BuildRequest (busca_id + collection_keys + type=mixed)
    → Scheduler.SetSchedule(params)
      → dispatchJob (valida busca_id ≠ "")
        → Collector.Collect(target, busca_id)
          → BigQuery INSERT (busca_id, keyword=collection_key)
```

### Read path (consulta)

```
Frontend (buscaId = busca.id, SEMPRE)
  → GET /api/lojas/novidades?busca_id=UUID
    → Analyzer: WHERE busca_id = @busca_id (exact match)
      → BigQuery clustering O(1)
```

## Consequências

### Positivas

- **Bug corrigido:** "Novos" e "Quedas" retornam produtos para buscas tipo loja/mista
- **Performance:** `WHERE busca_id = @id` é O(1) via clustering vs `LIKE '%X%'` (full scan). Redução estimada de ~90% no custo BigQuery.
- **Determinismo cross-service:** mesma busca → mesmas collection_keys em qualquer linguagem, validado no CI
- **Dívida técnica reduzida:** `GET /api/lojas` não retorna mais campo `keyword` legado; `GET /api/buscas` sem fallback de shopNames via Keyword
- **Contrato validável:** `mise run check:service-contracts` valida existência do schema no registry
- **Upsert por ID:** `POST /api/buscas` e `?remover=true` aceitam UUID, não dependem mais de match por keyword string
- **Inference declarativo:** regras em JSON, não espalhadas em código ad-hoc

### Negativas / Riscos

- ~~**`Busca.Keyword` (singular) mantido no entity**~~ ✅ **Resolvido** — campo removido na migração. Entity agora usa `Keywords` (string[] array).
- ~~**`Busca.Marketplaces` continua string comma-separated**~~ ✅ **Resolvido** — migrado para `string[]` (jsonb). Entity usa `public string[] Marketplaces`.
- **`db:reset` destrutivo** — dados antigos descartados. Aceitável porque MVP com 2 usuários e cron repopula em 8h.
- **Sem protovalidate/CEL** — validação cross-field ficou no JSON Schema e no código (Go rejeita busca_id vazio). Protovalidate é evolução futura quando mais serviços consumirem o proto.

### Neutras

- Proto evolution é aditiva (field 7) — zero breaking change no wire format
- Fixtures cresceram de 5 para 8 (3 edge cases) — custo mínimo de manutenção
- CI ganha ~10s a mais no job `contracts` (cross-language diff) — desprezível

## Alternativas Consideradas

| Alternativa | Descartada porque |
|-------------|-------------------|
| Backfill migration (popular busca_id em dados antigos) | Complexidade desnecessária — MVP sem dados críticos, `db:reset` é mais simples |
| LIKE com índice (busca_id LIKE pattern otimizado) | Resolve performance mas não resolve ambiguidade de identidade |
| Proto como fonte de verdade (gRPC message como schema) | Proto não é consumido pelo Frontend; JSON Schema é mais acessível |
| GraphQL para unificar contratos | Overengineering para 2 usuários; REST + JSON Schema é suficiente |
| Remover Busca.Keyword agora | ~~Requer EF Core migration + atualizar todos os seeders/testes — separado para limitar blast radius~~ ✅ Já feito |

## Validação

| Suite | Resultado |
|-------|-----------|
| Frontend (vitest) | 314 tests ✅ |
| C# (xUnit) | 88 tests ✅ |
| Go (12 packages) | All pass ✅ |
| Python (ruff) | Clean ✅ |
| E2E Novos | 8/8 checks ✅ |
| Contracts check | Valid ✅ |
| Rules schema | Valid ✅ |
| Cross-language fixtures | Go = Python = JS ✅ |

Acceptance criteria do fix imediato:
1. ✅ "Novos" para busca tipo "loja" retorna produtos
2. ✅ "Quedas" para busca tipo "mista" retorna variações
3. ✅ Zero hits de `LIKE` em `services/analyzer/routes/` para identificação de busca

## Evolução Futura

- ~~**EF Core migration:** remover `Busca.Keyword` e migrar `Busca.Marketplaces` para `string[]` (jsonb)~~ ✅ Implementado (Busca.Keyword removido, Marketplaces migrado para string[] jsonb)
- **Protovalidate + CEL:** validação cross-field no proto quando mais serviços consumirem
- **Contract test no CI (C#):** adicionar `dotnet test` com fixtures quando jsonschema NuGet estiver disponível
- ~~**Inference engine no Frontend:** aplicar `rules/busca-rules.json` inference em tempo real (shop → marketplace automático)~~ ✅ Implementado (BuscaEngine FSM aplica inference via rules)
- **Cache Layer (ADR-0031):** CacheService.Get usa collection_keys derivadas para cache lookup, validando BuscaContract antes de armazenar

## Atualização 2026-07-12 — Omnibox (T-0055) compõe o mesmo contrato

A substituição das raias pelo **omnibox** (input unificado; ver ADR-0027 v4 e
`.kiro/specs/omnibox-input/`) é **puramente de frontend e NÃO altera o BuscaContract**.
O omnibox apenas oferece uma nova forma de *compor* o mesmo contexto de busca:

- `parsearInput` tokeniza o texto (`@loja`/`#categoria`/`!marketplace`/keyword) e
  `tokensParaContexto` resolve os tokens para `{ keyword, shopIds, categorias,
  marketplacesFiltro }` — exatamente os campos de identidade que alimentam
  `deriveCollectionKeys` e o `busca_id`.
- A resolução/persistência de busca continua via BuscaEngine + endpoints existentes
  (`ADICIONAR_LOJA`, `DIGITAR`, `POST /api/buscas`), então `busca_id`, collection_keys
  e o read/write path descritos acima permanecem idênticos. Zero mudança de schema,
  proto, BigQuery ou fixtures.

## Arquivos-Chave

| Arquivo | O quê |
|---------|-------|
| `contracts/schemas/busca-contract.json` | Schema canônico (fonte de verdade) |
| `contracts/registry.yaml` | Boundary `busca-contract` com 5 consumers |
| `rules/busca-rules.json` | Seções `inference` + `derivation` |
| `fixtures/buscas.json` | 8 fixtures cross-language |
| `internal/busca/collection_keys.go` | Go: DeriveCollectionKeys |
| `services/analyzer/collection_keys.py` | Python: derive_collection_keys |
| `src/Garimpei.Domain/CollectionKeys.cs` | C#: CollectionKeys.Derive |
| `web/src/lib/collection-keys.js` | JS: deriveCollectionKeys |
| `protos/collector/v1/collector.proto` | `string busca_id = 7` no CollectRequest |
| `deploy/bigquery_schema.sql` | `busca_id STRING NOT NULL` na tabela snapshots |
| `services/analyzer/routes/novidades.py` | `WHERE busca_id = @busca_id` (exact match) |
| `services/analyzer/routes/quedas.py` | `busca_id` opcional para filtragem per-busca |
| `web/src/lib/descobrir.js` | `const buscaId = b.id` (fix) |
| `services/scheduler/jobs.go` | Passa `BuscaId` ao Collector, rejeita vazio |
| `src/Garimpei.Api/Endpoints/SchedulerJobs.cs` | BuildRequest com busca_id + collection_keys + type=mixed |
| `src/Garimpei.Api/Endpoints/BuscasEndpoints.cs` | Upsert por ID + fallback keyword |
| `src/Garimpei.Api/Endpoints/LojasEndpoints.cs` | GET sem campo `keyword` legado |
| `.mise/tasks/check/fixtures-crosslang` | CI cross-language diff |
| `.github/workflows/ci.yml` | Job contracts com fixtures-crosslang |
