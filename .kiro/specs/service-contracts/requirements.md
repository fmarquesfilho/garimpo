# Requirements Document

## Introduction

Service Contracts define the formal agreements between the Garimpo system components — the C# API (Cloud Run), Go microservices (publisher, collector, alerter, scheduler), Python analyzer, and the Svelte frontend (Cloudflare Pages). After the migration from Go monolith to C# + Go microservices, several integration failures occurred due to missing contracts: UUID vs chat_id confusion, silent endpoint failures, JSON casing mismatches, and data format inconsistencies. This feature establishes machine-verifiable contracts at each integration boundary, with CI checks that block deployment when contracts are violated.

## Glossary

- **Contract_Registry**: A centralized YAML/JSON file that declares each service's provided and consumed interfaces (endpoints, gRPC methods, event payloads, expected schemas)
- **Contract_Validator**: A CI script that reads the Contract_Registry and verifies all declared interfaces exist and match across service boundaries
- **Integration_Test_Suite**: A set of tests that verify cross-service behavior by invoking real gRPC calls or HTTP endpoints between services (using Docker Compose or testcontainers)
- **Schema_Contract**: A machine-readable definition of the JSON payload shape (field names, types, required fields) exchanged between two services
- **Boundary**: A point of communication between two services (HTTP endpoint, gRPC call, shared database table, message payload)
- **C#_API**: The ASP.NET Core 10 web application serving the REST API from Cloud Run
- **Publisher_Service**: The Go sidecar responsible for sending messages to Telegram/WhatsApp via gRPC
- **Collector_Service**: The Go sidecar responsible for fetching product data from Shopee via gRPC
- **Analyzer_Service**: The Python service that queries BigQuery for price drops, new products, and statistics
- **Frontend**: The Svelte SPA deployed on Cloudflare Pages that consumes the C# API via HTTP/JSON
- **GroupId_Resolution**: The process by which the C# API resolves a PostgreSQL UUID (destino_id) into a chat_id string that the Publisher_Service understands
- **Proto_Source_Of_Truth**: The .proto files in protos/ that define the gRPC contracts between C# and Go services

## Requirements

### Requirement 1: Contract Registry Declaration

**User Story:** As a developer, I want a centralized registry of all service boundaries and their contracts, so that I can see at a glance what each service expects and provides.

#### Acceptance Criteria

1. THE Contract_Registry SHALL declare every Boundary between services as a named entry containing: source service, target service, protocol (http/grpc/db), endpoint or method name, and reference to the Schema_Contract
2. WHEN a new service Boundary is introduced, THE Contract_Registry SHALL require an entry before the CI pipeline passes
3. THE Contract_Registry SHALL be stored in a machine-readable format (YAML) at `contracts/registry.yaml` in the repository root
4. THE Contract_Registry SHALL include entries for all existing boundaries: C#_API → Publisher_Service (gRPC Publish), C#_API → Collector_Service (gRPC Search), C#_API → Analyzer_Service (HTTP), Frontend → C#_API (REST/JSON)

### Requirement 2: JSON Schema Contracts for HTTP Boundaries

**User Story:** As a developer, I want machine-verifiable schemas for all JSON payloads exchanged over HTTP, so that field naming (snake_case), types, and required fields are enforced automatically.

#### Acceptance Criteria

1. THE Schema_Contract SHALL define JSON Schema files for every HTTP request and response payload in the Contract_Registry
2. THE Schema_Contract SHALL enforce snake_case field naming for all Frontend → C#_API payloads
3. WHEN the C#_API returns a response to the Frontend, THE Schema_Contract SHALL validate that all field names use snake_case (not camelCase or PascalCase)
4. THE Schema_Contract SHALL define the exact type for identifier fields: destino_id as UUID string, chat_id as numeric string, produto_id as string
5. WHEN a Schema_Contract defines a field as an array type, THE C#_API SHALL serialize that field as a JSON array (not as a comma-separated string or single value)

### Requirement 3: gRPC Contract Enforcement

**User Story:** As a developer, I want the proto definitions to be the single source of truth for inter-service gRPC communication, so that type mismatches are caught at compile time.

#### Acceptance Criteria

1. THE Proto_Source_Of_Truth SHALL be the only definition for gRPC message formats between the C#_API and Go microservices
2. WHEN the Proto_Source_Of_Truth is modified, THE Contract_Validator SHALL verify that generated stubs in gen/go/ and src/Garimpei.Protos/Generated/ are regenerated and in sync
3. THE Proto_Source_Of_Truth SHALL include field-level documentation comments specifying the semantic meaning and valid value range for ambiguous fields (group_id: "Telegram/WhatsApp chat identifier, numeric string, NOT a PostgreSQL UUID")
4. WHEN the C#_API calls Publisher_Service.Publish with a group_id value, THE C#_API SHALL pass only resolved chat_id values (not PostgreSQL UUIDs) as enforced by GroupId_Resolution

### Requirement 4: Integration Tests for Cross-Service Orchestration

**User Story:** As a developer, I want integration tests that verify the complete flow across services, so that silent failures (endpoint saves to DB but does not publish) are caught before deploy.

#### Acceptance Criteria

1. THE Integration_Test_Suite SHALL include a test verifying that POST /api/publicar triggers a gRPC Publish call to the Publisher_Service (not just a database write)
2. THE Integration_Test_Suite SHALL include a test verifying that the GroupId_Resolution correctly resolves a PostgreSQL UUID destino_id to a chat_id string before calling Publisher_Service
3. THE Integration_Test_Suite SHALL include a test verifying that POST /api/publicacoes with immediate scheduling (no AgendadaEm) triggers a gRPC Publish call
4. THE Integration_Test_Suite SHALL run against real service containers (via Docker Compose or testcontainers) to verify actual gRPC connectivity
5. IF a gRPC service is unreachable during an Integration_Test_Suite run, THEN THE Integration_Test_Suite SHALL report a clear connectivity failure (not a silent pass)

### Requirement 5: CI Contract Validation Pipeline

**User Story:** As a developer, I want CI to block deploys when any service contract is violated, so that integration failures are caught before reaching production.

#### Acceptance Criteria

1. THE Contract_Validator SHALL run as a required CI job that blocks merge when any contract violation is detected
2. WHEN a developer modifies an HTTP endpoint signature in the C#_API, THE Contract_Validator SHALL verify the change is reflected in the corresponding Schema_Contract
3. WHEN a developer modifies the Frontend API client (api.js), THE Contract_Validator SHALL verify all referenced endpoints exist in the C#_API with matching parameter names
4. THE Contract_Validator SHALL verify that all gRPC proto files pass buf lint and that generated stubs are in sync (extending existing proto CI job)
5. THE Contract_Validator SHALL verify that the Contract_Registry entries match the actual service implementations (no stale or missing entries)

### Requirement 6: Data Format Consistency Enforcement

**User Story:** As a developer, I want automated checks that enforce consistent data formats across boundaries, so that issues like keywords stored as string vs array are caught immediately.

#### Acceptance Criteria

1. THE Schema_Contract SHALL define the canonical format for multi-value fields (keywords as JSON array of strings, never as comma-separated single string)
2. WHEN the Frontend sends a multi-value field, THE C#_API SHALL accept it only as a JSON array and reject comma-separated strings with a 400 error
3. WHEN the C#_API stores a multi-value field in PostgreSQL, THE C#_API SHALL persist it in the same array format defined by the Schema_Contract
4. THE Contract_Validator SHALL include a check that all database column types for array fields use PostgreSQL array or jsonb types (not text with delimiters)

### Requirement 7: Orchestration Documentation

**User Story:** As a developer, I want up-to-date documentation of service orchestration flows, so that new developers understand the publish flow, collection flow, and alerting flow.

#### Acceptance Criteria

1. THE Contract_Registry SHALL include a `flows` section documenting each end-to-end orchestration: publish flow (Frontend → C#_API → Publisher_Service → Telegram/WhatsApp), collection flow (Scheduler → Collector_Service → BigQuery), alerting flow (Analyzer_Service → Alerter_Service → Telegram)
2. WHEN a developer modifies a service's role in an orchestration flow, THE Contract_Validator SHALL verify that the flows documentation references only valid Boundary entries from the Contract_Registry
3. THE Contract_Registry flows documentation SHALL be auto-verifiable: each step in a flow must reference an existing Boundary entry

### Requirement 8: Backward Compatibility Checks

**User Story:** As a developer, I want to be warned when a contract change is backward-incompatible, so that I can plan migrations instead of breaking running services.

#### Acceptance Criteria

1. WHEN a Schema_Contract removes a required field or changes a field type, THE Contract_Validator SHALL flag the change as a breaking change in the CI output
2. WHEN a proto file removes or renames a field, THE Contract_Validator SHALL flag the change as a breaking change using buf breaking (comparing against the main branch)
3. IF a breaking change is detected, THEN THE Contract_Validator SHALL require explicit acknowledgment via a `BREAKING:` prefix in the commit message to allow the merge
