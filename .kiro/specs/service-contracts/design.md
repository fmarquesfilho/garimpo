# Technical Design: Service Contracts

## Overview

This design implements machine-verifiable contracts between all Garimpo services, using a layered approach:

1. **Static contracts** (registry + JSON schemas) — caught at CI time
2. **Compile-time contracts** (proto + generated stubs) — caught at build time
3. **Runtime contracts** (integration tests) — caught at test time

The design leverages existing infrastructure (CI scripts, proto checks, pre-push hook) and adds minimal new tooling.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    contracts/                                     │
│  ├── registry.yaml          ← All boundaries + flows            │
│  ├── schemas/               ← JSON Schema per endpoint          │
│  │   ├── publicacoes.request.json                               │
│  │   ├── publicacoes.response.json                              │
│  │   ├── buscas.request.json                                    │
│  │   ├── buscas.response.json                                   │
│  │   ├── destinos.request.json                                  │
│  │   └── ...                                                    │
│  └── README.md              ← How to add new contracts          │
├── scripts/                                                       │
│  └── check-service-contracts.sh  ← CI validator                 │
├── protos/                   ← gRPC source of truth (existing)   │
│  └── publisher/v1/publisher.proto (enhanced with docs)          │
└── src/Garimpei.Tests/                                           │
   └── Integration/                                                │
       └── PublishFlowTests.cs ← Cross-service integration tests  │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Contract Registry (`contracts/registry.yaml`)

Central declaration of every service boundary. Machine-readable, human-reviewable.

```yaml
version: "1.0"
services:
  - id: csharp-api
    name: "Garimpei API (C#)"
    runtime: cloud-run
    port: 8080
  - id: publisher
    name: "Publisher (Go)"
    runtime: cloud-run-sidecar
    port: 50052
    proto: protos/publisher/v1/publisher.proto
  - id: collector
    name: "Collector (Go)"
    runtime: cloud-run-sidecar
    port: 50051
    proto: protos/collector/v1/collector.proto
  - id: alerter
    name: "Alerter (Go)"
    runtime: cloud-run-sidecar
    port: 50053
    proto: protos/alerter/v1/alerter.proto
  - id: scheduler
    name: "Scheduler (Go)"
    runtime: cloud-run-sidecar
    port: 50054
    proto: protos/scheduler/v1/scheduler.proto
  - id: analyzer
    name: "Analyzer (Python)"
    runtime: cloud-run-sidecar
    port: 8060
  - id: frontend
    name: "Frontend (Svelte)"
    runtime: cloudflare-pages

boundaries:
  # ── Frontend → C# API ──────────────────────────────────
  - id: fe-publicacoes
    source: frontend
    target: csharp-api
    protocol: http
    method: POST
    path: /api/publicacoes
    request_schema: contracts/schemas/publicacoes.request.json
    response_schema: contracts/schemas/publicacoes.response.json
    notes: "Envio imediato (sem agendada_em) DEVE triggerar gRPC Publish"

  - id: fe-buscas-get
    source: frontend
    target: csharp-api
    protocol: http
    method: GET
    path: /api/buscas
    response_schema: contracts/schemas/buscas.response.json
    notes: "keywords DEVE ser array, nunca string única"

  - id: fe-buscas-post
    source: frontend
    target: csharp-api
    protocol: http
    method: POST
    path: /api/buscas
    request_schema: contracts/schemas/buscas.request.json
    notes: "keywords aceita array de strings"

  - id: fe-destinos
    source: frontend
    target: csharp-api
    protocol: http
    method: GET
    path: /api/destinos
    response_schema: contracts/schemas/destinos.response.json

  - id: fe-publicar
    source: frontend
    target: csharp-api
    protocol: http
    method: POST
    path: /api/publicar
    request_schema: contracts/schemas/publicar.request.json

  # ── C# API → Go Sidecars (gRPC) ────────────────────────
  - id: api-publisher
    source: csharp-api
    target: publisher
    protocol: grpc
    service: publisher.v1.PublisherService
    method: Publish
    notes: |
      group_id DEVE ser o chat_id resolvido (ex: @mileseleciona, -1001234).
      NUNCA um UUID do PostgreSQL. A resolução acontece via GroupId_Resolution
      no C# antes de chamar o gRPC.

  - id: api-collector
    source: csharp-api
    target: collector
    protocol: grpc
    service: collector.v1.CollectorService
    method: Fetch

  # ── C# API → Python Analyzer (HTTP) ────────────────────
  - id: api-analyzer
    source: csharp-api
    target: analyzer
    protocol: http
    method: GET
    path: /candidatos
    notes: "Curadoria/scoring de produtos"

  # ── Scheduler → Collector (gRPC) ───────────────────────
  - id: scheduler-collector
    source: scheduler
    target: collector
    protocol: grpc
    service: collector.v1.CollectorService
    method: Fetch

  # ── Scheduler → Analyzer (HTTP) ────────────────────────
  - id: scheduler-analyzer
    source: scheduler
    target: analyzer
    protocol: http
    method: POST
    path: /detect-coupons

flows:
  publish:
    name: "Publicação de Oferta"
    steps:
      - boundary: fe-publicacoes
        action: "Frontend envia produto + destino_id + legenda"
      - boundary: api-publisher
        action: "API resolve destino_id → chat_id, chama Publisher.Publish"
      - action: "Publisher envia para Telegram/WhatsApp via Bot API"
    invariants:
      - "destino_id (UUID) é resolvido para chat_id ANTES do gRPC"
      - "Se envio imediato, DEVE chamar gRPC (não apenas salvar no banco)"
      - "Se agendamento, salva com status 'agendada' sem chamar gRPC"

  collection:
    name: "Coleta de Produtos"
    steps:
      - boundary: scheduler-collector
        action: "Scheduler dispara Fetch por keyword/loja"
      - action: "Collector busca na Shopee API e grava no BigQuery"
      - boundary: scheduler-analyzer
        action: "Scheduler chama Analyzer para detectar novidades"
    invariants:
      - "Coleta falha gracefully por marketplace (não aborta)"

  alerting:
    name: "Alertas de Preço"
    steps:
      - action: "Analyzer detecta queda de preço"
      - action: "Alerter envia notificação via Telegram"
    invariants:
      - "Alerter usa token próprio (ALERTAS_TELEGRAM_TOKEN), não o do Publisher"
```

### 2. JSON Schema Contracts (`contracts/schemas/`)

Each schema validates the exact payload shape. Example for publicações:

```json
// contracts/schemas/publicacoes.request.json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "POST /api/publicacoes request",
  "type": "object",
  "properties": {
    "produto_id": { "type": "string" },
    "nome": { "type": "string" },
    "categoria": { "type": ["string", "null"] },
    "preco": { "type": "number", "minimum": 0 },
    "comissao": { "type": "number" },
    "link": { "type": ["string", "null"] },
    "imagem": { "type": ["string", "null"] },
    "estrategia": { "type": ["string", "null"] },
    "destino_id": {
      "type": "string",
      "description": "UUID do destino no PostgreSQL. API resolve para chat_id antes de publicar.",
      "format": "uuid"
    },
    "template_id": { "type": ["string", "null"] },
    "agendada_em": { "type": ["string", "null"], "format": "date-time" },
    "legenda_custom": { "type": ["string", "null"] }
  },
  "required": ["destino_id"],
  "additionalProperties": false,
  "x-naming": "snake_case"
}
```

```json
// contracts/schemas/buscas.response.json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "GET /api/buscas response",
  "type": "object",
  "properties": {
    "buscas": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string", "format": "uuid" },
          "keywords": {
            "type": "array",
            "items": { "type": "string" },
            "minItems": 1,
            "description": "SEMPRE array. Nunca string única ou comma-separated."
          },
          "ativo": { "type": "boolean" },
          "criado_em": { "type": "string", "format": "date-time" },
          "sort_by": { "type": "string" },
          "limit": { "type": "integer" }
        },
        "required": ["id", "keywords"]
      }
    },
    "total": { "type": "integer" }
  },
  "required": ["buscas", "total"]
}
```

### 3. Contract Validator Script (`scripts/check-service-contracts.sh`)

A bash script (following the pattern of existing check-*.sh scripts) that:

1. **Parses `contracts/registry.yaml`** and validates structure
2. **Verifies boundaries exist in code**:
   - HTTP paths → grep in C# Endpoints/*.cs and frontend api.js
   - gRPC methods → verify proto files contain the declared service/method
3. **Validates JSON schemas** against snake_case naming (no camelCase keys)
4. **Validates flow references** → each boundary_id in flows must exist in boundaries
5. **Checks for orphan endpoints** → endpoints in code not declared in registry (warning)

```bash
#!/usr/bin/env bash
# check-service-contracts.sh — Validates contracts/registry.yaml against code
# Requires: yq (YAML parser), jq (JSON parser)

# 1. Parse registry, extract all HTTP boundaries
# 2. For each HTTP boundary, verify path exists in C# endpoints
# 3. For each gRPC boundary, verify service+method in .proto
# 4. For each schema file referenced, verify it exists and passes JSON Schema meta-validation
# 5. For each flow step, verify boundary_id exists
# 6. Check all schema field names are snake_case
```

### 4. Integration Tests (C#)

New test class in `src/Garimpei.Tests/Integration/PublishFlowTests.cs` that uses a mock gRPC server (Grpc.Core.Testing) to verify:

```csharp
[Fact]
public async Task PublicacoesImediato_DeveChamarPublisher()
{
    // Arrange: POST /api/publicacoes with destino_id and no agendada_em
    // Assert: gRPC PublishAsync was called with resolved chat_id (not UUID)
}

[Fact]
public async Task PublicacoesAgendado_NaoDeveChamarPublisher()
{
    // Arrange: POST /api/publicacoes with agendada_em set
    // Assert: gRPC PublishAsync was NOT called
    // Assert: DB record has status "agendada"
}

[Fact]
public async Task GroupIdResolution_ResolveUuidParaChatId()
{
    // Arrange: Destino in DB with Config = "@mileseleciona"
    // Act: POST /api/publicar with destino_id = UUID
    // Assert: gRPC called with group_id = "@mileseleciona"
}
```

These use `WebApplicationFactory<Program>` with a mocked `PublisherServiceClient` injected via DI, no Docker needed for unit-level verification. Full Docker Compose integration tests are a separate optional CI stage.

### 5. Proto Documentation Enhancement

Add semantic comments to proto fields that have been sources of confusion:

```protobuf
message PublishRequest {
  string owner_uid = 1;
  string channel = 2; // "telegram" | "whatsapp"

  // RESOLVED chat identifier for the target channel.
  // For Telegram: "@channel_name" or numeric chat_id (e.g. "-1001234567890").
  // For WhatsApp: phone number in E.164 format.
  // ⚠️ NEVER pass a PostgreSQL UUID here. The C# API must resolve
  // Destino.Config before calling this method (see GroupId_Resolution).
  string group_id = 3;

  PublishContent content = 4;
}
```

### 6. CI Integration

Add to `.github/workflows/ci.yml`:

```yaml
  # ── Service Contracts ───────────────────────────────────────────
  contracts:
    name: Service contracts check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - name: Install tools
        run: |
          sudo wget -qO /usr/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
          sudo chmod +x /usr/bin/yq
      - name: Validate contracts
        run: ./scripts/check-service-contracts.sh
      - name: Proto breaking changes
        run: |
          cd protos && buf breaking --against '../.git#branch=main'
```

Also added to the pre-push hook's `block_checks` function.

### 7. Backward Compatibility (buf breaking)

The existing `proto` CI job already runs `buf lint`. Add `buf breaking` comparison against main branch. If breaking change detected:
- CI outputs clear message: "⚠️ BREAKING: field X removed in publisher.proto"
- Merge blocked unless commit message contains `BREAKING:` prefix

For JSON schemas: a simple diff check — if a required field is removed or type changed, flag as breaking.

## Data Models

### Contract Registry Schema (registry.yaml)

```yaml
version: string           # semver format
services: Service[]       # list of all services
boundaries: Boundary[]    # all integration points
flows: Flow[]             # end-to-end orchestrations

Service:
  id: string              # unique identifier
  name: string            # human name
  runtime: enum           # cloud-run | cloud-run-sidecar | cloudflare-pages
  port: int?              # service port
  proto: string?          # path to proto file (gRPC services only)

Boundary:
  id: string              # unique, used in flow references
  source: string          # service id
  target: string          # service id
  protocol: enum          # http | grpc | db
  method: string?         # HTTP method or gRPC method
  path: string?           # HTTP path (for http protocol)
  service: string?        # gRPC service name (for grpc protocol)
  request_schema: string? # path to JSON Schema file
  response_schema: string? # path to JSON Schema file
  notes: string?          # semantic documentation

Flow:
  name: string            # human readable
  steps: FlowStep[]       # ordered steps
  invariants: string[]    # rules that must always hold

FlowStep:
  boundary: string?       # reference to boundary.id
  action: string          # description of what happens
```

### JSON Schema Contract Format

Each schema file follows JSON Schema 2020-12 with custom extensions:

- `x-naming: "snake_case"` — enforced by validator
- `format: "uuid"` — for PostgreSQL identifiers
- `description` — documents semantic meaning and valid values
- `additionalProperties: false` — strict schema (no undeclared fields)

## Data Flow: Publish (with contracts)

```
Frontend                    C# API                         Publisher (Go)
   │                          │                                │
   │ POST /api/publicacoes    │                                │
   │ {destino_id: UUID,       │                                │
   │  nome, preco, legenda}   │                                │
   │─────────────────────────>│                                │
   │                          │                                │
   │  ┌─── Schema validation (publicacoes.request.json)        │
   │  │    ✓ destino_id is UUID format                         │
   │  │    ✓ all fields snake_case                             │
   │  └───                                                     │
   │                          │                                │
   │  ┌─── GroupId_Resolution                                  │
   │  │    SELECT config FROM destinos WHERE id = UUID         │
   │  │    resolved = "@mileseleciona"                         │
   │  └───                                                     │
   │                          │                                │
   │                          │ gRPC Publish(group_id=          │
   │                          │   "@mileseleciona")            │
   │                          │───────────────────────────────>│
   │                          │                                │
   │                          │      PublishResponse           │
   │                          │<───────────────────────────────│
   │                          │                                │
   │ {publicacao: {status:    │                                │
   │   "enviada", detalhe}}   │                                │
   │<─────────────────────────│                                │
```

## File Structure (new files)

```
contracts/
├── registry.yaml                    ← Service boundaries + flows
├── schemas/
│   ├── publicacoes.request.json
│   ├── publicacoes.response.json
│   ├── publicar.request.json
│   ├── buscas.request.json
│   ├── buscas.response.json
│   ├── destinos.response.json
│   ├── favoritos.response.json
│   └── categorias.response.json
└── README.md                        ← How to add contracts

scripts/
└── check-service-contracts.sh       ← CI validator (new)

src/Garimpei.Tests/Integration/
└── PublishFlowTests.cs              ← Integration tests (new)

protos/publisher/v1/publisher.proto  ← Enhanced with semantic docs (edit)
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Registry YAML is malformed | Validator exits with parse error + line number |
| Boundary references non-existent service | Validator reports "unknown service 'X'" |
| Schema file missing | Validator reports "schema not found: path" |
| Schema has camelCase field | Validator reports "naming violation in field 'fieldName'" |
| Flow references non-existent boundary | Validator reports "flow step references unknown boundary 'X'" |
| Proto breaking change without BREAKING: prefix | CI blocks merge with explanation |
| gRPC unreachable in integration test | Test fails with clear "connection refused to :50052" |

## Testing Strategy

### Unit Tests (existing, enhanced)
- Dispatcher tests verify fallback behavior when destino not in store
- Schema validation tested via JSON Schema meta-validation

### Integration Tests (new: `PublishFlowTests.cs`)
- Use `WebApplicationFactory<Program>` with mocked `PublisherServiceClient`
- Verify publish flow triggers gRPC call with correct parameters
- Verify GroupId_Resolution resolves UUID → chat_id
- Verify agendamento does NOT trigger gRPC

### Contract Validation (CI)
- `check-service-contracts.sh` runs on every push/PR
- `buf breaking` catches proto incompatibilities
- Pre-push hook includes contract check in the `block_checks` parallel group

### E2E (existing Playwright)
- Existing E2E tests verify frontend → backend flows
- No changes needed — contracts catch issues earlier in pipeline

## Correctness Properties

### Property 1: GroupId Invariant
Any value passed as `group_id` in a gRPC PublishRequest MUST be a resolved chat_id (starts with `@`, `-100`, `+`, or is numeric). Never a UUID format.

**Validates: Requirements 3.4**

### Property 2: Array Serialization Invariant
Any field declared as `type: array` in a Schema_Contract MUST be serialized as a JSON array in both request and response. Never as comma-separated string.

**Validates: Requirements 2.5, 6.1**

### Property 3: Boundary Completeness
Every HTTP endpoint in the C# API that is called by the frontend MUST have a corresponding boundary entry in the registry.

**Validates: Requirements 1.1, 5.3**

### Property 4: Flow Referential Integrity
Every `boundary` reference in a flow step MUST point to an existing boundary entry by id.

**Validates: Requirements 7.2, 7.3**

### Property 5: Proto Sync
Generated stubs MUST always match the current proto files (existing check, reinforced).

**Validates: Requirements 3.1, 3.2**

## Trade-offs & Decisions

| Decision | Rationale |
|----------|-----------|
| YAML registry (not code-gen) | Simple, readable, no new tooling dependency. Scripts validate against code. |
| JSON Schema (not OpenAPI) | Granular per-endpoint validation. OpenAPI is heavier and we already have openapi.yaml for external docs. |
| Bash validator (not custom tool) | Follows existing pattern (check-*.sh). No new language/runtime needed. |
| Mock gRPC in C# tests (not Docker) | Fast feedback. Docker Compose integration tests are optional/nightly. |
| `buf breaking` for proto compat | Industry standard, already using buf for lint. |
| snake_case enforced globally | One JSON naming policy (already configured). Schemas validate field names. |

## Migration Notes

- Existing `check-api-contract.sh` continues to run (checks route existence)
- New `check-service-contracts.sh` is additive (checks contracts, schemas, flows)
- No database schema changes needed
- Proto changes are documentation-only (no field renaming)
