# Requirements Document

## Introduction

Serviço de regras como sidecar Go usando gorules/zen-go para avaliar Decision Models JSON (JDM) versionados no repositório. O serviço recebe contexto via gRPC, avalia decision tables e expression nodes, e retorna decisões estruturadas (intent, guards, normalização, fontes necessárias). Integra-se ao Cloud Run multi-container existente como 6º container, com proxy REST via API C# e consumo pelo frontend (BuscaEngine).

## Requirements

### Requirement 1: Serviço gRPC Rules

**User Story:** As a developer, I want a rules evaluation service accessible via gRPC, so that the frontend can offload complex decision logic to the backend.

#### Acceptance Criteria

1. THE rules-service SHALL expose a gRPC server on port 50055 implementing `RulesService.EvaluateRules`.
2. WHEN `EvaluateRules` is called with a context map and decision_id, THE service SHALL evaluate the JDM decision model and return intent, guards, normalized values, required sources, and validation.
3. WHEN decision_id is empty, THE service SHALL use "busca-rules" as default.
4. WHEN the engine is not loaded, THE service SHALL return gRPC UNAVAILABLE.
5. THE service SHALL respond in under 1ms for typical evaluations (target: 50-200μs).
6. THE service SHALL support concurrent evaluations from multiple goroutines without locking.

### Requirement 2: JDM Versionado no Repositório

**User Story:** As a developer, I want decision models stored as JSON files in the git repository, so that rule changes are reviewable via PR and tracked in version history.

#### Acceptance Criteria

1. THE JDM files SHALL reside in the `rules/` directory at the repository root.
2. THE service SHALL load all `.json` files from the rules directory on boot.
3. THE JDM SHALL define decision tables for: intent resolution, guards evaluation, and normalization.
4. WHEN a JDM file is invalid, THE service SHALL fail to start with a clear error message indicating which file and what parsing error.
5. THE JDM format SHALL follow the GoRules JSON Decision Model specification.

### Requirement 3: Hot-Reload sem Downtime

**User Story:** As a operator, I want to reload rules without restarting the service, so that rule updates can take effect during a running deployment.

#### Acceptance Criteria

1. WHEN the service receives SIGHUP, THE service SHALL re-read all JDM files from disk and swap the engine atomically.
2. WHEN a reload fails (invalid JDM), THE service SHALL keep the previous engine active and log the error.
3. DURING reload, THE service SHALL continue serving evaluations using the previous engine without interruption.
4. THE service SHALL expose a `ReloadRules` RPC as programmatic alternative to SIGHUP.

### Requirement 4: Proto Contract

**User Story:** As a developer, I want a typed gRPC contract for rules evaluation, so that both C# and Go can communicate with compile-time safety.

#### Acceptance Criteria

1. THE proto SHALL be defined at `protos/rules/v1/rules.proto` with package `rules.v1`.
2. THE proto SHALL define `EvaluateRulesRequest` with `map<string, string> context` and `string decision_id`.
3. THE proto SHALL define `EvaluateRulesResponse` with fields: `intent`, `GuardsResult`, `NormalizedValues`, `repeated string required_sources`, `ValidationResult`, `int64 evaluation_time_us`.
4. THE proto SHALL define `ReloadRulesRequest`/`ReloadRulesResponse` for programmatic reload.
5. THE proto SHALL generate Go stubs (gen/go/rules/v1) and C# stubs (Garimpei.Protos).

### Requirement 5: Proxy REST no C# API

**User Story:** As a frontend developer, I want a REST endpoint to evaluate rules, so that I can call it from the browser without gRPC-web complexity.

#### Acceptance Criteria

1. THE C# API SHALL expose `POST /api/rules/evaluate` that accepts JSON body with `context` (dict) and optional `decisionId`.
2. THE C# API SHALL proxy the request to the rules-service via gRPC (localhost:50055).
3. THE C# API SHALL return the response as JSON with camelCase field names.
4. WHEN the rules-service is unavailable, THE C# API SHALL return HTTP 503 with graceful error message.
5. THE endpoint SHALL require authentication (Firebase JWT) via the existing middleware.

### Requirement 6: Decision Tables — Intent

**User Story:** As a business analyst, I want the intent of a search to be determined by a decision table, so that I can modify search behavior by editing a JSON table.

#### Acceptance Criteria

1. THE JDM SHALL contain a decision table that resolves intent from `hasKeyword` × `hasShop` inputs.
2. THE decision table SHALL produce exactly one of: `keyword_na_loja`, `keyword_global`, `loja_completa`, `nenhum`.
3. THE decision table SHALL use hit policy `first` (first matching row wins).
4. WHEN both keyword and shop are present, THE intent SHALL be `keyword_na_loja`.

### Requirement 7: Decision Tables — Guards

**User Story:** As a developer, I want guards evaluated declaratively via decision table, so that adding new guard conditions doesn't require code changes.

#### Acceptance Criteria

1. THE JDM SHALL contain a decision table that evaluates guards: `temContextoBusca`, `podeSalvar`.
2. `temContextoBusca` SHALL be true when keyword is non-empty OR shopIds is non-empty.
3. `podeSalvar` SHALL be true only when `temContextoBusca` is also true.
4. THE guards table SHALL be extensible — new guards can be added as columns without code changes.

### Requirement 8: Expression Nodes — Normalização

**User Story:** As a developer, I want value normalization defined as expressions in the JDM, so that normalization rules are declarative and version-controlled.

#### Acceptance Criteria

1. THE JDM SHALL contain expression nodes that normalize `comissaoMin` (divide by 100 if > 1, max 4 decimals).
2. THE JDM SHALL contain expression nodes that normalize `vendasMin` (floor, min 0).
3. THE normalized values SHALL be returned in the `NormalizedValues` field of the response.
4. WHEN normalization is applied, THE response SHALL always contain valid decimal values (never NaN or Infinity).

### Requirement 9: Integração Frontend (BuscaEngine)

**User Story:** As a frontend developer, I want the BuscaEngine to use the rules service for complex decisions, so that business logic is centralized and frontend stays thin.

#### Acceptance Criteria

1. THE BuscaEngine effects module SHALL include an `evaluateRules(ctx)` function that calls `POST /api/rules/evaluate`.
2. WHEN adding a shop or saving a search, THE BuscaEngine SHALL call `evaluateRules` to get the intent and validation result.
3. THE BuscaEngine SHALL cache rules evaluation results for 30 seconds to avoid redundant calls on rapid interactions.
4. WHEN the rules service is unavailable, THE BuscaEngine SHALL fall back to local guards evaluation (existing JavaScript logic as fallback).
5. Simple guards (temContextoBusca for debounce) SHALL remain local for zero-latency UX.

### Requirement 10: Deployment

**User Story:** As a DevOps engineer, I want the rules service deployed as part of the existing Cloud Run multi-container, so that it shares the same lifecycle and networking.

#### Acceptance Criteria

1. THE Dockerfile SHALL produce a minimal Alpine-based image with the Go binary and rules/ directory.
2. THE Cloud Run service definition SHALL include the rules container with: CPU 0.25, RAM 128Mi, startup probe on gRPC health.
3. THE CI pipeline SHALL build the rules-service image alongside the other 5 containers.
4. THE deploy step SHALL copy `rules/` directory into the container image at build time.

### Requirement 11: Testes

**User Story:** As a developer, I want comprehensive tests for the rules service, so that JDM changes are validated before deployment.

#### Acceptance Criteria

1. THE test suite SHALL verify each intent resolution case (4 combinations keyword × shop).
2. THE test suite SHALL verify guard consistency (podeSalvar implies temContextoBusca).
3. THE test suite SHALL verify normalization idempotency.
4. THE test suite SHALL include property-based tests for determinism (same input → same output).
5. THE test suite SHALL verify concurrent evaluation + reload does not produce errors.
6. THE tests SHALL run via `go test ./services/rules/...` in CI.
