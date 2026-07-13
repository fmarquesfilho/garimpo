# Implementation Plan: Test Quality Pipeline

## Overview

Pipeline de qualidade de testes: classificação por camada, mutation testing (Stryker), coverage granular, paridade cross-language, e tasks mise compostas. 12 tasks em 4 waves.

## Tasks

- [ ] 1. Instalar dependências e configurar coverage
  - Adicionar `@vitest/coverage-v8` como devDep (pinned) em web/package.json
  - Configurar coverage em vitest.config.js: provider 'v8', thresholds por módulo, reporter html+text+lcov
  - Adicionar `reports/` ao .gitignore
  - Rodar `bunx vitest run --coverage` e verificar que relatório é gerado
  - **Requirements: 3.1, 3.2, 3.3, 3.4**

- [x] 2. Configurar StrykerJS (mutation testing frontend)
  - Instalar `@stryker-mutator/core` e `@stryker-mutator/vitest-runner` como devDeps (pinned)
  - Criar `web/stryker.config.json` com mutate targeting módulos core (busca-engine, omnibox-intencao, busca-config, omnibox-parser, loja-registry)
  - Configurar thresholds: high=80, low=70, break=60
  - Configurar reporters: html (reports/mutation/), clear-text, progress
  - Habilitar incremental mode
  - Rodar `bunx stryker run` e verificar relatório gerado + mutation score
  - **Requirements: 2.1, 2.3, 2.5, 2.6**

- [x] 3. Configurar Stryker.NET (mutation testing backend)
  - Instalar dotnet tool: `dotnet tool install dotnet-stryker`
  - Criar `src/stryker-config.json` targeting Garimpei.Domain com Garimpei.Tests
  - Configurar thresholds: high=80, low=70, break=60
  - Configurar reporter html em reports/mutation-dotnet/
  - Rodar `dotnet stryker` e verificar mutation score do Domain (Loja.Normalizar, etc.)
  - **Requirements: 2.2, 2.3, 2.5**

- [x] 4. Expandir fixture de normalização + teste de paridade JS
  - Expandir `fixtures/normalizacao-pares.json` com edge cases: emoji+texto, CJK, whitespace-only, diacríticos compostos (NFD), números puros, strings longas, null/empty
  - Criar `web/src/tests/normalizacao-parity.test.js` que importa `normalizarNome` de loja-registry.js e valida contra o mesmo fixture
  - Verificar que ambos os testes (C# LojaTests + JS parity) leem o fixture e passam
  - Se divergências encontradas, corrigir a implementação (C# é autoritativa)
  - **Requirements: 4.1, 4.2, 4.3, 4.4**

- [ ] 5. Classificar testes frontend via vitest workspace
  - Adicionar configuração `test.workspace` em vitest.config.js com projects: unit, integration, regression
  - Definir globs para cada project conforme mapeamento no design.md
  - Verificar que `bunx vitest run --project unit` roda apenas testes unitários
  - Verificar que `bunx vitest run --project integration` roda apenas integração
  - Verificar que `bunx vitest run` (sem filtro) roda todos
  - **Requirements: 1.1, 1.2, 1.3, 1.5, 1.6**

- [x] 6. Criar testes faltantes da Smart Search (edge cases)
  - StoreCard.svelte: testes de componente (render com/sem imagem, monitorada, campos condicionais, click emite evento)
  - Engine edge cases: OMNIBOX_INPUT null/undefined, OMNIBOX_SELECIONAR fora de bounds, OMNIBOX_KEYDOWN com opcoes vazio, MONITORAR_LOJA com id null, BUSCAR_LOJAS whitespace
  - omnibox-intencao: categoria sem marketplaces relevantes (filtrada), intencao desabilitada, categoria como string pura
  - Executar intencao `resolver_link` e `categoria` end-to-end via engine
  - `#executarSugestaoLegado` (todos os 4 tipos: loja, categoria, marketplace, busca_salva)
  - **Requirements: todos (cobertura)**

- [x] 7. Criar task mise `test:fast`
  - Cria `.mise/tasks/test/fast`: vitest com --project unit + dotnet test filtrado por Architecture|Domain|CollectionKeys|Scheduler|Tenancy|Services
  - Target: <10s
  - **Requirements: 5.1**

- [x] 8. Criar tasks mise `test:integration`, `test:regression`, `test:parity`, `test:descobrir`
  - `test:integration`: vitest --project integration + dotnet test --filter Integration|Persistence
  - `test:regression`: vitest --project regression
  - `test:parity`: vitest normalizacao-parity.test.js + dotnet test --filter LojaTests
  - `test:descobrir`: vitest run filtrado para busca-engine-*, omnibox-*, busca-config, busca-duplicata, busca-unificada
  - **Requirements: 5.5, 1.2, 1.3, 4.2**

- [x] 9. Criar tasks mise `test:coverage`, `test:mutate`, `test:mutate-backend`
  - `test:coverage`: vitest run --coverage (verifica thresholds, falha se abaixo)
  - `test:mutate`: StrykerJS com incremental mode (módulos core)
  - `test:mutate-backend`: Stryker.NET contra Domain
  - **Requirements: 2.1, 2.2, 3.1, 5.2**

- [x] 10. Criar tasks mise compostas `test:pre-push`, `test:ci`, `test:full`
  - `test:pre-push`: depends=[test:unit, test:integration, test:regression, test:parity] + bun run check + bun run build
  - `test:ci`: depends=[test:pre-push, test:coverage]
  - `test:full`: depends=[test:ci, test:e2e-local]
  - Atualizar `test:all` para depender de `test:full`
  - Atualizar `test:unit` existente para usar nova classificação
  - **Requirements: 5.2, 5.3, 5.4**

- [ ] 11. Integrar com CI (GitHub Actions)
  - Atualizar `.github/workflows/ci.yml`: step `mise run test:ci` no job de testes
  - Adicionar job condicional `mutation` que roda `mise run test:mutate` + `test:mutate-backend` apenas em merge para main ou label `run-mutation`
  - Cachear `reports/mutation/stryker-incremental.json` entre runs
  - **Requirements: 6.1, 6.2, 6.3, 6.5**

- [ ] 12. Validação final e documentação
  - Rodar cada task mise individualmente e verificar output correto
  - Medir tempos: test:fast <10s, test:unit <15s, test:integration <30s
  - Verificar mutation score baseline (anotar no design.md)
  - Atualizar README.md com seção de testes (tasks disponíveis)
  - Commitar: `feat(test): pipeline de qualidade — mutation testing, coverage, tasks mise`
  - **Requirements: todos**

## Task Dependency Graph

```json
{
  "waves": [
    {"tasks": [1, 2, 3, 4]},
    {"tasks": [5, 6]},
    {"tasks": [7, 8, 9]},
    {"tasks": [10, 11, 12]}
  ]
}
```

Wave 1: Setup de ferramentas (coverage, Stryker, fixture, deps). Paralelas.
Wave 2: Classificação de testes + testes faltantes. Depende das ferramentas.
Wave 3: Tasks mise granulares. Depende da classificação.
Wave 4: Tasks compostas + CI + validação. Depende das granulares.

## Notes

- StrykerJS suporta Svelte nativamente desde v3.30 e tem runner nativo para Vitest.
- Stryker.NET agora suporta Microsoft Testing Platform runner (mais rápido que VSTest).
- A classificação por glob preserva 100% de retrocompatibilidade — nenhum teste existente precisa ser renomeado.
- O mutation testing completo pode levar 2-5 minutos nos módulos core; o modo incremental (default em dev) muta apenas arquivos alterados.
- O pre-push hook existente (`.husky/pre-push`) pode ser atualizado para rodar `mise run test:pre-push` em vez do script atual.
- Coverage + mutation juntos dão a visão completa: coverage diz "isso foi executado", mutation diz "isso foi verificado".
