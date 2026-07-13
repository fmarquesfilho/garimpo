# Requirements: Test Quality Pipeline

## Introdução

A suite de testes do Garimpei cresceu organicamente (419 Vitest + 94 xUnit) mas carece de:
1. Classificação formal (unit vs integration vs regression vs e2e)
2. Medição objetiva de qualidade (mutation score, não apenas coverage)
3. Tasks mise granulares para execução seletiva (local e CI)
4. Contrato de paridade normalização C#/JS (cross-language)

Esta spec define a pipeline de qualidade de testes que permite ao dev executar testes por camada, medir a eficácia com mutation testing (Stryker), e otimizar o fluxo tanto na página Descobrir como no app inteiro.

## Glossário

- **Unit Test**: Teste de função/módulo puro, sem I/O, sem DOM, sem banco. Executa em <10ms.
- **Integration Test**: Teste que cruza fronteiras (componente Svelte + DOM, query EF Core + InMemory DB). Mock de APIs externas mas exerce integração interna.
- **Regression Test**: Teste que reproduz um bug específico reportado e previne reincidência. Nomeado com referência ao bug.
- **E2E Test**: Teste via Playwright contra app compilado (local com mocks ou prod com API real).
- **Mutation Score**: Percentual de mutantes "mortos" pela suite. Mede se testes verificam comportamento (não só executam código).
- **Stryker**: Framework de mutation testing. StrykerJS (Vitest runner) para frontend; Stryker.NET para backend C#.
- **Coverage**: Percentual de linhas/branches executadas. Métrica necessária mas insuficiente.
- **mise task**: Script executável via `mise run <task>`. Composable, com depends e paralelismo.

## Requirements

### Requisito 1: Classificação Formal de Testes por Camada

**User Story:** Como desenvolvedor, eu quero testes separados por camada (unit/integration/regression/e2e) com tasks mise distintas, para executar rapidamente apenas o que preciso durante o desenvolvimento.

#### Critérios de Aceitação

1. WHEN eu rodo `mise run test:unit`, THE pipeline SHALL executar apenas testes unitários (funções puras, sem DOM, sem DB) e completar em <10s.
2. WHEN eu rodo `mise run test:integration`, THE pipeline SHALL executar testes de integração (componentes Svelte + DOM, queries EF Core + InMemory) e completar em <30s.
3. WHEN eu rodo `mise run test:regression`, THE pipeline SHALL executar testes de regressão (bugs reproduzidos) categorizados por prefixo/tag.
4. WHEN eu rodo `mise run test:e2e-local`, THE pipeline SHALL executar testes E2E Playwright contra o app local com mocks (sem backend real).
5. THE classificação SHALL ser declarativa: testes são categorizados por convención de diretório ou pattern de nome de arquivo (ex: `*.unit.test.js`, `*.integration.test.js`, `*.regression.test.js`).
6. FOR ALL testes existentes (419 Vitest + 94 xUnit), a categorização SHALL preservar retrocompatibilidade — testes existentes continuam passando sem rename obrigatório.

### Requisito 2: Mutation Testing com Stryker (Frontend + Backend)

**User Story:** Como desenvolvedor, eu quero medir o mutation score da minha suite de testes, para saber se os testes realmente verificam comportamento e não apenas executam código.

#### Critérios de Aceitação

1. WHEN eu rodo `mise run test:mutate`, THE pipeline SHALL executar StrykerJS contra os módulos core da página Descobrir (busca-engine, omnibox-intencao, busca-config, omnibox-parser, loja-registry) e gerar relatório HTML.
2. WHEN eu rodo `mise run test:mutate-backend`, THE pipeline SHALL executar Stryker.NET contra o projeto Garimpei.Domain (Loja.Normalizar, guards, entities) e gerar relatório.
3. THE mutation score target SHALL ser ≥70% para módulos core e ≥50% para módulos auxiliares (thresholds configuráveis).
4. IF o mutation score cai abaixo do threshold em CI, THEN THE pipeline SHALL falhar com mensagem indicando quais mutantes sobreviveram.
5. THE relatório de mutation testing SHALL ser gerado em `reports/mutation/` (frontend) e `reports/mutation-dotnet/` (backend), ambos ignorados pelo git.
6. THE pipeline SHALL suportar execução incremental (`--since HEAD~1`) para rodar apenas contra arquivos alterados (viabiliza uso frequente em desenvolvimento).

### Requisito 3: Coverage com Thresholds por Módulo

**User Story:** Como desenvolvedor, eu quero ver a cobertura de código por módulo com thresholds mínimos, para identificar áreas negligenciadas.

#### Critérios de Aceitação

1. WHEN eu rodo `mise run test:coverage`, THE pipeline SHALL executar vitest com `@vitest/coverage-v8` e gerar relatório por arquivo.
2. THE coverage thresholds SHALL ser definidos no `vitest.config.js` por glob pattern: módulos core (busca-engine, omnibox-*, busca-config) ≥80%, componentes ≥60%, utilitários ≥50%.
3. IF a cobertura cai abaixo do threshold em algum módulo, THEN `bun run test:unit` SHALL falhar com indicação do módulo e gap.
4. THE relatório SHALL incluir branches coverage (não apenas linhas) para detectar paths não exercitados.
5. WHEN combinado com mutation testing, THE pipeline SHALL exibir ambos: coverage (quantitativo) e mutation score (qualitativo) lado a lado.

### Requisito 4: Paridade Cross-Language (C# ↔ JS) da Normalização

**User Story:** Como desenvolvedor, eu quero garantia automática de que `Loja.Normalizar()` (C#) e `normalizarNome()` (JS) produzem output idêntico para o mesmo input, para que a busca local funcione consistentemente.

#### Critérios de Aceitação

1. THE pipeline SHALL manter um fixture compartilhado `fixtures/normalizacao-pares.json` com pares `{input, expected}` usados por AMBOS os testes C# e JS.
2. WHEN eu rodo `mise run test:parity`, THE pipeline SHALL executar ambos os testes (C# e Vitest) contra o mesmo fixture e falhar se qualquer divergência for detectada.
3. THE fixture SHALL cobrir edge cases: strings com emoji, diacríticos compostos (NFD), CJK, números puros, whitespace-only, null/empty.
4. IF um novo par for adicionado ao fixture, BOTH testes SHALL detectá-lo automaticamente sem edição de código.

### Requisito 5: Tasks mise Compostas e Granulares

**User Story:** Como desenvolvedor, eu quero tasks mise que combinem os níveis certos de teste para cada cenário (dev rápido, pre-push, CI full), para otimizar meu feedback loop.

#### Critérios de Aceitação

1. `mise run test:fast` SHALL executar apenas testes unitários puros (Vitest filtrado + xUnit Architecture) em <10s — feedback loop de desenvolvimento.
2. `mise run test:pre-push` SHALL executar unit + integration + regression + parity + svelte-check + build — tudo que o pre-push hook valida.
3. `mise run test:ci` SHALL executar unit + integration + regression + parity + coverage + mutation (incremental) — validação completa sem E2E.
4. `mise run test:full` SHALL executar ci + e2e-local — validação máxima local.
5. `mise run test:descobrir` SHALL executar apenas testes relacionados à página Descobrir (engine, omnibox, busca-config, store card) — feedback rápido para essa feature.
6. FOR ALL tasks, THE saída SHALL indicar claramente qual camada está rodando, tempo de execução, e resultado (pass/fail com contagens).

### Requisito 6: Integração com CI (GitHub Actions)

**User Story:** Como desenvolvedor, eu quero que o CI rode a pipeline de testes adequada sem incluir E2E reais ou mutation testing pesado em todo PR.

#### Critérios de Aceitação

1. THE CI workflow SHALL rodar `test:ci` em todo PR (unit + integration + regression + parity + coverage).
2. THE CI workflow SHALL rodar mutation testing APENAS no merge para main (ou via label `run-mutation`), por ser computacionalmente pesado.
3. THE CI SHALL cachear relatórios de mutation entre runs para baseline comparison.
4. THE CI SHALL publicar coverage e mutation score como comment no PR (quando disponível).
5. THE CI SHALL NOT rodar E2E reais (`test:e2e-prod`) — conforme regra existente no steering `ci.md`.
