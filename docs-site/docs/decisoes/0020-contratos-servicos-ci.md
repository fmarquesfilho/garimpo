# ADR-0020: Contratos de serviço e validação CI

## Status

Aceito (2026-07-02)

## Contexto

Após a migração Go → C# + Go microserviços (ADR-0012), tivemos uma série de
falhas de integração silenciosas em produção:

1. **UUID vs chat_id**: C# API passava UUID do PostgreSQL como `group_id` ao
   publisher Go, que esperava um chat_id Telegram (`@canal` ou numérico)
2. **Endpoint sem efeito**: `POST /api/publicacoes` salvava no banco mas não
   chamava o publisher gRPC — a publicação nunca era enviada ao Telegram
3. **Array vs string**: `GET /api/buscas` retornava `keywords` como string única
   quando o frontend esperava array
4. **JSON naming**: Frontend enviava `destino_id` (snake_case) mas o C# não
   mapeava corretamente sem policy configurada

Nenhum dos checks existentes (testes unitários, lint, build) pegava esses
problemas porque eram **falhas de contrato entre serviços**, não bugs dentro de
um serviço isolado.

## Decisão

Implementar um sistema de **contratos de serviço verificáveis por máquina**:

### 1. Registry central (`contracts/registry.yaml`)

Declara todos os serviços, fronteiras de integração e fluxos de orquestração em
YAML. Cada fronteira documenta protocolo, método, path/service, e referencia
schemas quando aplicável.

### 2. JSON Schemas (`contracts/schemas/*.json`)

Schemas JSON Schema 2020-12 para payloads HTTP críticos, enforçando:
- Nomenclatura snake_case
- Tipos corretos (UUID para `destino_id`, array para `keywords`)
- Campos obrigatórios documentados

### 3. Validador CI (`.mise/tasks/check/service-contracts`)

Script que valida automaticamente:
- Fronteiras HTTP existem no código C#
- Métodos gRPC existem nos protos
- Schemas referenciados são JSON válido e usam snake_case
- Referências de flows apontam para boundaries existentes

### 4. Proto com documentação semântica

Campos ambíguos nos .proto recebem comentários documentando o formato esperado
e restrições (ex: `group_id` = NUNCA UUID, SEMPRE chat_id resolvido).

### 5. Testes de integração para invariantes

Testes C# que verificam os contratos de orquestração:
- Publicação imediata DEVE chamar gRPC
- GroupId_Resolution resolve UUID → chat_id
- Agendamento NÃO chama gRPC

### 6. `buf breaking` para compatibilidade proto

Compara proto changes contra main; breaking changes bloqueiam merge sem
prefixo `BREAKING:` no commit.

## Alternativas consideradas

| Alternativa | Motivo da rejeição |
|-------------|-------------------|
| OpenAPI validation | Pesado demais; já temos openapi.yaml para docs externas. JSON Schema é mais granular. |
| Contract testing (Pact) | Over-engineering para o tamanho do time. Shell scripts seguem o padrão existente. |
| Nenhum (só code review) | Falhou — os bugs passaram pelo review humano. |
| gRPC reflection at startup | Resolve só parte do problema (gRPC) e adiciona complexidade runtime. |

## Evolução futura: Taskfile

Os scripts `check-*.sh` na pasta `/scripts` estão se acumulando (8 atualmente).
Para o próximo ciclo, consideraremos migrar para [Taskfile](https://taskfile.dev)
ou [mise](https://mise.jdx.dev):

- **Taskfile**: YAML declarativo, dependências entre tasks, variáveis, cross-platform
- **mise**: além de tasks, gerencia versões de runtime (Go, Node, Python, .NET)
- Ambos substituiriam Makefile + scripts bash com algo mais legível e manutenível
- Decisão será registrada em ADR separada quando avaliada (T-0036)

## Consequências

### Positivas
- Falhas de integração são detectadas antes do deploy (CI + pre-push)
- Novos desenvolvedores entendem os fluxos pelo registry + schemas
- Breaking changes são coordenadas (não surpresas em produção)
- Proto é documentação executável (lint + breaking check)

### Negativas
- Overhead de manter registry.yaml atualizado a cada novo endpoint
- Dependência de `yq` no CI (leve — binário único)
- Schemas podem ficar desatualizados se não forem mantidos junto com o código

### Mitigações
- Validador CI falha se schema está faltando → força atualização
- README documenta o processo de adicionar contratos
- Pre-push roda localmente → feedback antes do push
