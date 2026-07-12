# Implementation Plan: Store Workflow Refactor

## Overview

Refatoração do subsistema de lojas da página Descobrir. Cria entidade Loja independente no backend,
implementa normalização robusta de nomes, estados de resolução explícitos, e separação clara entre
lojas monitoradas e escopadas. 11 tasks ordenadas por dependência (backend-first, frontend depois).

## Tasks

- [x] 1. Criar entidade Loja no domain e configurar persistência EF Core
  - Criar `src/Garimpei.Domain/Entities/Loja.cs` com campos: Id (Guid), OwnerUid, ShopId (long), Nome (max 200), NomeNormalizado (max 200), Marketplace (max 50), CronExpression, SourceUrl, OrigemPadrao, CreatedAt, UpdatedAt
  - Implementar `Loja.Normalizar(string nome)`: NFD → remove combining marks → keep [a-z0-9] → lowercase
  - Adicionar `DbSet<Loja> Lojas` ao AppDbContext com entity config (PK, índice único ShopId+Marketplace+OwnerUid, query filter tenant)
  - Gerar migration EF Core `AddLojaEntity`
  - Testes xUnit para `Loja.Normalizar()` com pares parametrizados
  - **Requirements: 1.1, 1.2, 3.1, 5.1**

- [x] 2. Implementar endpoints GET /api/lojas/registro e POST /api/lojas/resolver
  - Criar `ResolverLojaRequest` record (Input required, Marketplace optional, Origem optional)
  - GET /api/lojas/registro: lista Lojas do tenant → {id, nome, nome_normalizado, marketplace, monitorada, cron, origem}
  - POST /api/lojas/resolver: resolve via Collector gRPC, upsert na tabela Lojas, retorna loja registrada
  - Tratar erros gRPC (NotFound, InvalidArgument, Unimplemented) → 400 com mensagem descritiva
  - Validar marketplace contra lista de suportados
  - Testes de integração xUnit com mock do CollectorService
  - **Requirements: 1.3, 1.5, 5.1, 5.2, 5.4, 5.6**

- [x] 3. Criar módulo frontend loja-registry.js com normalização e matching
  - Criar `web/src/lib/loja-registry.js` com `normalizarNome(nome)` espelhando C# e `matchLojas(inputNormalizado, lojas, max)`
  - Criar fixture `fixtures/normalizacao-pares.json` consumível por Vitest e xUnit
  - Testes Vitest: normalizarNome com mesmos pares do backend, matchLojas com edge cases (@glory, @gloryofseoul, input < 2 chars)
  - **Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6**

- [x] 4. Refatorar estado da engine — substituir lojaResolvendo/lojaErro por resolucaoLoja
  - Em `busca-engine-state.js`: remover `lojaResolvendo` e `lojaErro`, adicionar `resolucaoLoja: { status: 'idle' }`
  - Atualizar guard `lojaInputValida` e adicionar guard `resolucaoPermitida`
  - Atualizar todos os testes que referenciam os campos removidos
  - Verificar `bun run check` passa
  - **Requirements: 4.7, 6.3**

- [x] 5. Refatorar effects — novo carregarRegistroLojas, remover listarLojasMonitoradas
  - Adicionar effect `carregarRegistroLojas()` → GET /api/lojas/registro
  - Alterar effect `resolverLoja(input)` → POST /api/lojas/resolver
  - Remover effect `listarLojasMonitoradas()`
  - Atualizar `api.js`: adicionar `listarRegistroLojas()`, renomear `adicionarLoja` → `resolverLoja`
  - Atualizar mocks nos testes existentes
  - **Requirements: 1.3, 1.4**

- [x] 6. Refatorar #adicionarLoja na BuscaEngine — separar caminhos sync/async
  - Renomear `#adicionarLojaMonitorada` → `#adicionarLojaConhecida` com campo `tipo` no shopMeta
  - Criar `#resolverLojaRemota(input)`: guard concorrência, timeout 10s, validação marketplace, populate contexto
  - Refatorar `#adicionarLoja(event)` como router: loja.id → sync; value → checar registro → sync ou async; inválido → erro
  - Após resolução: append loja a `ctx.lojasDisponiveis`
  - Testes: caminho sync, async sucesso, timeout, guard concorrência, marketplace inválido, payload inválido, retry
  - **Requirements: 4.1–4.10, 5.2–5.6, 6.1–6.5**

- [x] 7. Refatorar inicialização — carregar registro de lojas em paralelo
  - Adicionar `carregarRegistroLojas()` ao Promise.all em `#inicializar()`
  - Atribuir a `ctx.lojasDisponiveis`
  - Remover chamada legada a `listarLojasMonitoradas`
  - Graceful degradation se falha (lojasDisponiveis = [])
  - Atualizar testes de inicialização
  - **Requirements: 1.3, 1.4**

- [x] 8. Atualizar omnibox-sugestoes.js para usar normalização robusta
  - Importar `normalizarNome`, `matchLojas` de `loja-registry.js`
  - Refatorar `matchLojas()` local para usar normalização + delegation
  - Atualizar testes: @gloryofseoul encontra "Glory of Seoul", @glory idem, @le encontra "Le Botanic"
  - Garantir que `sugestoesCtx.lojasMonitoradas` usa formato com `nome_normalizado`
  - **Requirements: 3.2, 3.3, 3.4, 3.5, 3.6**

- [x] 9. Implementar distinção visual monitorada/escopada nos loja cards
  - Atualizar getter `lojaCards` para incluir `tipo` no output
  - Em `BuscaUnificada.svelte`: badge ⏱ com title cron para monitoradas
  - Acessibilidade: title descritivo no badge, aria-label no botão remover
  - **Requirements: 2.1, 2.2, 2.3, 2.4**

- [x] 10. Limpar código legado e atualizar rules/contracts
  - Remover `listarLojasMonitoradas` e `encontrarLojaPorNome` de `descobrir-logic.js`
  - Remover testes associados de `descobrir.test.js`
  - Atualizar `rules/busca-rules.json` com bloco `lojaRegistro`
  - Atualizar `contracts/schemas/lojas.request.json` e `lojas.response.json`
  - Executar `bun run check && bunx vitest run && bun run build` + `dotnet test` — tudo verde
  - **Requirements: 1.1, 5.1, 5.5**

- [x] 11. Validação end-to-end e deploy
  - Rodar suites completas: Vitest, dotnet test, check, build
  - Aplicar migration: `dotnet ef database update`
  - Playtest local com DEV_API_PROXY: autocomplete, resolver nova loja, badge monitorada, @gloryofseoul match
  - Verificar persistência: loja resolvida aparece após re-inicializar sem re-resolver
  - Commitar: `feat(lojas): refactor store workflow — registro independente, normalização, estados explícitos`
  - **Requirements: 1.3, 1.5, 3.6, 4.2, 4.7, 5.2, 6.4**

## Task Dependency Graph

```json
{
  "waves": [
    {"tasks": [1, 3, 4]},
    {"tasks": [2, 5]},
    {"tasks": [6]},
    {"tasks": [7]},
    {"tasks": [8, 9]},
    {"tasks": [10]},
    {"tasks": [11]}
  ]
}
```

Tasks 1, 3, 4 podem ser feitas em paralelo (sem dependência entre si).
Task 2 depende de 1. Task 5 depende de 4. Task 6 depende de 5 e 2.
Tasks 7-9 são sequenciais. Task 10 depende de todas. Task 11 é validação final.

## Notes

- A fixture `fixtures/normalizacao-pares.json` (Task 3) garante cross-validation entre C# e JS.
- Os endpoints legados (POST/GET/DELETE /api/lojas) são mantidos na Task 2 para não quebrar o Scheduler. Remoção futura.
- A Task 6 é a maior e mais complexa — contém a lógica central da refatoração. Pode ser subdividida se necessário.
- Req 2.5 e 2.6 (promoção/demoção de cron) são implementados implicitamente: ao re-carregar o registro após mudança de cron, o `tipo` derivado muda automaticamente.
