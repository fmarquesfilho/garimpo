# Contratos de Serviço — Garimpei

## O que é isso?

Este diretório contém a definição formal dos contratos entre todos os serviços do Garimpo. Cada ponto de comunicação (fronteira) entre serviços é declarado em `registry.yaml` e, quando aplicável, tem um JSON Schema correspondente em `schemas/`.

## Por que existe?

Após a migração Go → C# + microserviços, tivemos falhas de integração que passaram despercebidas:
- UUID passado como chat_id para o publisher (formatos incompatíveis)
- Endpoint que salvava no banco mas não publicava no Telegram
- Keywords retornadas como string quando o frontend esperava array
- JSON em camelCase quando o frontend esperava snake_case

Os contratos previnem essas falhas sendo verificados automaticamente pelo CI.

## Como funciona?

1. **`registry.yaml`** — Declara serviços, fronteiras e fluxos
2. **`schemas/*.json`** — JSON Schemas para payloads HTTP
3. **`.mise/tasks/check/service-contracts`** — Valida tudo no CI e pre-push

## Como adicionar um novo contrato

### Nova fronteira HTTP (Frontend → API)

1. Adicione a entrada em `registry.yaml` na seção `boundaries`:
   ```yaml
   - id: fe-novo-endpoint
     source: frontend
     target: csharp-api
     protocol: http
     method: POST
     path: /api/novo-endpoint
     request_schema: contracts/schemas/novo-endpoint.request.json
   ```

2. Crie o JSON Schema em `contracts/schemas/novo-endpoint.request.json`:
   ```json
   {
     "$schema": "https://json-schema.org/draft/2020-12/schema",
     "title": "POST /api/novo-endpoint request",
     "type": "object",
     "properties": {
       "campo_snake_case": { "type": "string" }
     },
     "required": ["campo_snake_case"],
     "additionalProperties": false,
     "x-naming": "snake_case"
   }
   ```

3. Se o endpoint faz parte de um fluxo, adicione ao `flows` correspondente.

### Nova fronteira gRPC (API → Sidecar)

1. Defina o serviço/método no `.proto` correspondente (em `protos/`)
2. Adicione a entrada em `registry.yaml`:
   ```yaml
   - id: api-servico-metodo
     source: csharp-api
     target: servico
     protocol: grpc
     service: servico.v1.ServicoService
     method: NomeDoMetodo
     notes: "Documentação semântica dos parâmetros"
   ```

3. Regenere os stubs: `cd protos && buf generate`

### Regras obrigatórias

- **Campos HTTP**: sempre `snake_case` (nunca camelCase/PascalCase)
- **Arrays**: sempre `type: array` no schema (nunca comma-separated string)
- **IDs**: `destino_id` = UUID do PostgreSQL; `chat_id`/`group_id` = identificador do canal (resolvido)
- **Breaking changes**: prefixe o commit com `BREAKING:` se remover campos ou mudar tipos

### Breaking changes

Quando uma mudança é incompatível com versões anteriores (remover campo obrigatório, mudar tipo, renomear campo no proto), o CI bloqueia o merge. Para prosseguir:

1. Certifique-se que a mudança é intencional e necessária
2. Coordene com os consumidores do contrato (ex: frontend precisa ser atualizado junto)
3. Use o prefixo `BREAKING:` na mensagem de commit:
   ```
   BREAKING: remover campo legacy_id do schema de publicacoes
   ```

O CI detecta breaking changes via:
- **Proto**: `buf breaking --against main` (renames, removals, type changes)
- **JSON Schema**: diff de campos `required` removidos entre branches

## Validação local

```bash
# Rodar o validador de contratos
mise run check:service-contracts

# Rodar todos os checks
mise run checks

# Ou o CI completo localmente
mise run ci
```

## Referência

- `protos/` — Fonte de verdade para gRPC (buf lint + buf breaking)
- `api/openapi.yaml` — Documentação externa da API (Scalar)
- `mise run check:api-contract` — Verifica rotas frontend ↔ backend (complementar)
