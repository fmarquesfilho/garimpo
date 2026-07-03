# Implementation Plan: Service Contracts

## Overview

Implementação de contratos verificáveis entre todos os serviços do Garimpo, adicionando um registry central, JSON schemas, script validador no CI, testes de integração e detecção de breaking changes.

## Tasks

- [x] 1. Create contract registry file (`contracts/registry.yaml`) with all services, boundaries, and orchestration flows as defined in the design. Create `contracts/README.md` explaining how to add new contracts. Validate YAML is parseable with `yq`. **Requirements: 1.1, 1.3, 1.4, 7.1**
- [x] 2. Create JSON Schema contracts for critical endpoints: `publicacoes.request.json`, `publicacoes.response.json`, `publicar.request.json`, `buscas.request.json`, `buscas.response.json`, `destinos.response.json`. Ensure all use snake_case, `keywords` as array type, `destino_id` as UUID format, and `additionalProperties: false`. **Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 6.1**
- [x] 3. Enhance proto documentation: add semantic constraints to `protos/publisher/v1/publisher.proto` for `group_id` (must be resolved chat_id, never UUID) and `description` field. Run `buf lint` and regenerate stubs. **Requirements: 3.1, 3.3**
- [x] 4. Create contract validator script (`scripts/check-service-contracts.sh`): validate registry structure, verify HTTP boundaries exist in C# code, verify gRPC boundaries exist in protos, check schema files exist and use snake_case, validate flow references, warn on orphan endpoints. **Requirements: 5.1, 5.2, 5.3, 5.5, 7.2, 7.3**
- [x] 5. Create integration tests for publish flow (`src/Garimpei.Tests/Integration/PublishFlowTests.cs`): test immediate publish triggers gRPC, scheduled publish does NOT trigger gRPC, GroupId_Resolution resolves UUID to chat_id, unreachable publisher returns clear error. **Requirements: 4.1, 4.2, 4.3, 4.5**
- [x] 6. Add contract validation to CI pipeline: new `contracts` job in ci.yml running `check-service-contracts.sh`, add `buf breaking` to proto job, add to pre-push hook parallel checks. **Requirements: 5.1, 5.4, 8.2**
- [x] 7. Add backward compatibility detection: configure `buf breaking` against main, detect `BREAKING:` commit prefix, add JSON schema diff warning for removed required fields, document convention in README. **Requirements: 8.1, 8.2, 8.3**

## Task Dependency Graph

```json
{
  "waves": [
    [1, 2, 3, 5],
    [4],
    [6],
    [7]
  ]
}
```

Wave 1: Tasks 1, 2, 3, 5 são independentes (registry, schemas, proto docs, integration tests).
Wave 2: Task 4 (validator) depende de 1 e 2.
Wave 3: Task 6 (CI) depende de 3 e 4.
Wave 4: Task 7 (breaking changes) depende de 6.

## Notes

- Tasks 1-2 podem ser feitas em paralelo (são arquivos independentes)
- Task 4 depende de 1 e 2 (precisa do registry e schemas para validar)
- Task 5 é independente e pode ser feita em paralelo com as demais
- Task 6 depende de 3 e 4 (precisa do script e proto docs para integrar no CI)
- Task 7 depende de 6 (estende o job de CI)
- Nenhuma mudança no schema do banco de dados é necessária
- As alterações no proto são apenas de documentação (sem renomear campos)
