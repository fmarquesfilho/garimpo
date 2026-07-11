# Requirements Document

## Introduction

O sistema Garimpei precisa de um contrato canônico unificado para o tipo "Busca" — definido uma vez e validado em todos os serviços (Frontend Svelte/TS, C# API, Go Scheduler/Collector, Python Analyzer). O contrato elimina a fragmentação atual onde cada serviço interpreta ad-hoc o que identifica uma busca, resultando em queries `LIKE` falhas e snapshots não correlacionados. Além do schema estático, o contrato inclui regras de inferência declarativas (shop → marketplace, shop → fontes) que garantem consistência de estado derivado.

## Glossary

- **BuscaContract**: Tipo canônico que representa uma busca em todos os serviços do sistema. Definido em `contracts/schemas/busca-contract.json`.
- **CollectionKey**: Identificador string derivado deterministicamente dos campos da busca (shop_ids → string, keywords → lowercase trimmed). Usado para indexar snapshots no BigQuery.
- **InferenceEngine**: Módulo que aplica regras declarativas de inferência sobre o estado da busca, derivando campos como `marketplaces` e `fontes` a partir de campos primários (`shop_ids`, `keywords`, `categorias`).
- **DerivedState**: Campos do BuscaContract que são calculados automaticamente a partir de campos primários via regras de inferência (ex: `collection_keys`, `marketplaces`, `fontes`, `tipo`).
- **BuscaRules**: Arquivo JSON declarativo (`rules/busca-rules.json`) contendo intent table, guards, normalização e regras de inferência.
- **Protovalidate**: Framework de validação baseado em CEL expressions embutidas em arquivos `.proto`, que gera validadores em todas as linguagens-alvo.
- **ExactMatch**: Query BigQuery que usa `WHERE busca_id = @id` em vez de `LIKE '%X%'`.
- **Fixture**: Instância concreta de BuscaContract em `fixtures/buscas.json`, usada como caso de teste cross-language.
- **SchemaRegistry**: Arquivo `contracts/registry.yaml` que declara todas as fronteiras de comunicação entre serviços e seus schemas.
- **Marketplace**: Plataforma de e-commerce suportada (shopee, mercado_livre, amazon).
- **Fonte**: Origem de dados para produtos (curadoria, lojas, quedas, novos, favoritos).

## Requirements

### Requirement 1: Schema Canônico BuscaContract

**User Story:** As a developer, I want a single canonical type definition for "Busca" enforced across all services, so that any schema change breaks the build immediately in Go, Python, C#, and TypeScript.

#### Acceptance Criteria

1. THE SchemaRegistry SHALL declare `busca-contract.json` as a boundary schema referenced by all services that produce or consume Busca data
2. WHEN a field is added, removed, or modified in `busca-contract.json`, THEN THE CI Pipeline SHALL fail for any service that does not conform to the updated schema
3. THE BuscaContract SHALL define the following required fields: id (string), tipo (enum), keywords (string array), shop_ids (int64 array), shop_names (map), categorias (string array), collection_keys (string array), marketplaces (string array), owner_uid (string), fontes (string array)
4. THE BuscaContract SHALL define the following optional fields: cron (string or null), comissao_min (float or null), vendas_min (integer or null)
5. WHEN a BuscaContract instance is validated, THEN THE Validator SHALL enforce that `id` is a non-empty string
6. WHEN a BuscaContract instance is validated, THEN THE Validator SHALL enforce that `owner_uid` is a non-empty string
7. WHEN a BuscaContract instance is validated, THEN THE Validator SHALL enforce that `marketplaces` contains at least one supported marketplace value
8. WHEN a BuscaContract instance is validated, THEN THE Validator SHALL enforce that `fontes` contains at least one active source value

### Requirement 2: Regras de Inferência Declarativas

**User Story:** As a developer, I want inference rules defined declaratively in a single JSON file, so that adding a shop from marketplace X automatically activates marketplace X in the busca state, without scattered ad-hoc logic.

#### Acceptance Criteria

1. WHEN a shop_id belonging to marketplace X is added to the busca, THEN THE InferenceEngine SHALL include marketplace X in the `marketplaces` field
2. WHEN `shop_ids` contains at least one item, THEN THE InferenceEngine SHALL include "lojas" in the `fontes` field
3. THE InferenceEngine SHALL compute `marketplaces` as the union of explicitly-selected marketplaces and marketplaces inferred from shop_ids
4. WHEN a user attempts to deactivate a marketplace that has active shops in scope, THEN THE InferenceEngine SHALL prevent the deactivation and keep the marketplace active
5. THE BuscaRules file SHALL contain a declarative `inference` section specifying all derivation rules in a machine-readable format
6. WHEN inference rules are modified in `busca-rules.json`, THEN THE CI Pipeline SHALL validate that all services consuming inference rules still pass their test suites

### Requirement 3: Derivação de Estado (Derived Fields)

**User Story:** As a developer, I want derived fields (collection_keys, tipo, marketplaces, fontes) computed deterministically from primary fields, so that the same busca always produces the same derived state regardless of which service computes it.

#### Acceptance Criteria

1. THE InferenceEngine SHALL derive `collection_keys` as the sorted union of shop_ids (as strings) and keywords (lowercase, trimmed)
2. WHEN `collection_keys` is derived, THEN THE InferenceEngine SHALL produce a non-empty array for any valid BuscaContract
3. THE InferenceEngine SHALL derive `tipo` from the combination of populated primary fields: keywords-only yields "keyword", shop_ids-only yields "loja" or "loja-multi", categorias-only yields "categoria", keywords+shop_ids yields "mista"
4. FOR ALL valid BuscaContract instances, the derivation of `collection_keys` SHALL be deterministic: the same input always produces the same output
5. FOR ALL valid BuscaContract instances, the derivation of `collection_keys` SHALL produce a sorted array with no duplicates
6. WHEN the same BuscaContract is processed by any service (Go, Python, C#, TypeScript), THEN THE derivation functions SHALL produce identical outputs for identical inputs

### Requirement 4: Cross-Service Validation (Protovalidate + CEL)

**User Story:** As a developer, I want cross-field validation rules embedded in the proto schema using protovalidate and CEL expressions, so that invalid states are rejected at the boundary in all languages without manual validation code.

#### Acceptance Criteria

1. WHEN `tipo` equals "keyword", THEN THE Validator SHALL enforce that `keywords` contains at least one non-empty string
2. WHEN `tipo` equals "loja", THEN THE Validator SHALL enforce that `shop_ids` contains at least one item
3. WHEN `tipo` equals "mista", THEN THE Validator SHALL enforce that both `keywords` and `shop_ids` contain at least one item
4. WHEN `collection_keys` is empty or null, THEN THE Validator SHALL reject the BuscaContract as invalid
5. WHEN `comissao_min` is provided, THEN THE Validator SHALL enforce it is between 0 and 1 inclusive
6. WHEN `vendas_min` is provided, THEN THE Validator SHALL enforce it is a non-negative integer
7. THE Proto definition SHALL embed validation rules using protovalidate CEL expressions that generate validators for Go, Python, C#, and TypeScript

### Requirement 5: Identificação por busca_id (Eliminar LIKE)

**User Story:** As a developer, I want all snapshot queries to use exact match on `busca_id` instead of LIKE patterns, so that queries are O(1) via clustering and never produce false matches.

#### Acceptance Criteria

1. THE Collector SHALL persist every snapshot row with a non-empty `busca_id` field corresponding to the originating BuscaContract's id
2. WHEN the Analyzer queries snapshots, THEN THE Analyzer SHALL use `WHERE busca_id = @busca_id` with parameterized exact match
3. THE Analyzer SHALL NOT use LIKE operators in any query involving busca identification
4. WHEN the Frontend requests novidades or quedas, THEN THE Frontend SHALL send the busca's UUID (`busca.id`) as the identifier, never a shop_id or keyword
5. WHEN a query returns no results for a given `busca_id`, THEN THE Analyzer SHALL return an empty response without fallback to LIKE or alternative matching strategies

### Requirement 6: Correlação End-to-End (Write Path)

**User Story:** As a developer, I want the busca_id to flow through the entire write path (Frontend → C# API → Scheduler → Collector → BigQuery), so that every snapshot can be traced back to its originating busca.

#### Acceptance Criteria

1. WHEN the C# API creates a schedule for a busca, THEN THE C# API SHALL include `busca_id` and `collection_keys` in the SetSchedule request parameters
2. WHEN the Scheduler dispatches a collection job, THEN THE Scheduler SHALL pass `busca_id` to the Collector in the CollectRequest
3. WHEN the Collector persists a snapshot to BigQuery, THEN THE Collector SHALL write the `busca_id` and the `collection_key` to every row
4. IF the Scheduler receives a job without `busca_id` in its parameters, THEN THE Scheduler SHALL reject the job with an InvalidArgument error
5. IF the Collector receives a CollectRequest without `busca_id`, THEN THE Collector SHALL reject the request with an InvalidArgument error

### Requirement 7: Migração Sem Backward Compatibility

**User Story:** As a developer, I want to reset all existing data and start fresh with the new schema, so that there is zero legacy code path and no LIKE fallback logic.

#### Acceptance Criteria

1. THE BigQuery snapshots table SHALL require `busca_id` as a NOT NULL column after migration
2. WHEN the migration runs, THEN THE System SHALL drop all existing snapshot data via `db:reset`
3. THE Analyzer SHALL NOT contain any fallback logic that uses LIKE when `busca_id` is not found
4. THE Scheduler cron cycle SHALL repopulate snapshot data with correct `busca_id` values within one cron period (8 hours) after migration

### Requirement 8: Busca Identity e Deduplicação

**User Story:** As a developer, I want stable busca identity (UUID) used for deduplication in BigQuery and across services, so that duplicate buscas are detected consistently.

#### Acceptance Criteria

1. THE BuscaContract SHALL use a UUID as its stable identity (`id` field) for all cross-service correlation
2. WHEN checking for duplicate buscas, THEN THE System SHALL compare identity fields as defined in `busca-rules.json` `buscaDuplicada.camposIdentidade` (keyword, shopIds, categorias, marketplacesFiltro)
3. WHEN normalizing identity fields for comparison, THEN THE System SHALL apply the normalization rules specified in `busca-rules.json` (sort for shopIds, sort_lowercase for categorias, sort for marketplacesFiltro)

### Requirement 9: Fixtures Cross-Language

**User Story:** As a developer, I want shared test fixtures in `fixtures/buscas.json` that every service uses for validation, so that behavioral drift between implementations is detected immediately.

#### Acceptance Criteria

1. THE Fixtures file SHALL contain at least one example for each `tipo` value (keyword, loja, loja-multi, categoria, mista)
2. WHEN CI runs, THEN THE CI Pipeline SHALL execute `deriveCollectionKeys` in all four languages (Go, Python, C#, TypeScript) against the shared fixtures and assert identical outputs
3. WHEN a fixture is added or modified, THEN THE CI Pipeline SHALL verify that all language implementations produce matching results
4. THE Fixtures file SHALL include edge cases: empty keywords with shop_ids, multiple keywords with mixed casing, shop_ids with keywords that look like numbers

### Requirement 10: Proto Evolution (CollectRequest)

**User Story:** As a developer, I want the CollectRequest proto to include `busca_id` as a required field, so that the Collector always knows which busca originated the collection.

#### Acceptance Criteria

1. THE CollectRequest proto message SHALL include a `string busca_id` field
2. WHEN a CollectRequest is received with an empty `busca_id`, THEN THE Collector SHALL reject it with an InvalidArgument status
3. THE CollectResponse SHALL continue returning `persisted: true/false` indicating BigQuery write status
