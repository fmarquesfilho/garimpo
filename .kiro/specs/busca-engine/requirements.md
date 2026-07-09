# Requirements Document

## Introduction

A página Garimpar (BuscaUnificada) sofre de múltiplos bugs causados pela falta de um modelo de estado coerente. O estado é gerenciado de forma ad hoc — variáveis soltas em `criarEstado()`, sem invariantes definidas, sem validação de transições, e sem forma de garantir que uma mudança propaga corretamente para todos os componentes dependentes.

Esta feature implementa uma **State Machine Engine** leve (~200 linhas) que modela toda a lógica da página como uma máquina de estados finita (FSM). A engine substitui o padrão atual (`BuscaUnificada.svelte.js` com getters/setters reativos) por um modelo declarativo com estados explícitos, transições validadas por guards, e efeitos colaterais injetáveis.

**Bugs resolvidos pela engine:**
1. Adicionar loja não restringe resultados à loja selecionada
2. Comissão mínima exibe float cru ("7.0000000")
3. Categorias perderam autocomplete (categorias Shopee de primeiro nível)
4. Chip de busca salva mostra "(sem keywords)" ao invés do label correto
5. Clicar no chip limpa keyword e não mostra resultados
6. Toggle 🆕 Novos não recarrega após coleta
7. Impossível validar execução de busca agendada

**Princípio:** A engine é 100% testável sem DOM, sem Svelte, sem API real — efeitos são mocks injetáveis.

## Glossary

- **Engine**: Módulo JavaScript puro (`busca-engine.js`) que implementa a máquina de estados — recebe config, retorna objeto com `send(event)` e `state` reativo.
- **Machine_Config**: Objeto declarativo que define estados, transições, guards e effects da máquina.
- **Machine_State**: Estado atual da máquina — um par `(stateName, context)` onde stateName é um dos estados possíveis e context contém os dados associados.
- **Machine_Context**: Dados mutáveis associados ao estado da máquina — keyword, shopIds, filtros, resultados, buscas salvas, erros.
- **Machine_Event**: Ação despachada para a máquina que pode causar uma transição — ex: DIGITAR, ADICIONAR_LOJA, MUDAR_FILTRO.
- **Guard**: Função pura que recebe (context, event) e retorna boolean — pré-condição para permitir uma transição.
- **Effect**: Função assíncrona associada a uma transição — executa efeitos colaterais (chamadas API) e retorna eventos de resposta.
- **Transition**: Regra declarativa `(estadoAtual, evento) → (novoEstado, updates ao context, effects a executar)`.
- **BuscaUnificada**: Componente Svelte que renderiza o Machine_State e despacha Machine_Events — view pura sem lógica de negócio.
- **Effects_Module**: Módulo separado (`busca-engine-effects.js`) que implementa os efeitos concretos (chamadas API) — injetável para testes.
- **Busca_Salva**: Configuração de busca persistida no servidor — inclui keyword, shopIds, filtros, fontes e cron.
- **Fonte**: Uma das fontes de dados: curadoria, quedas, novos, lojas, favoritos.
- **Debounce_Timer**: Timer interno que atrasa execução de busca após mudança de parâmetros (400ms).
- **Shopee_Categories**: Lista de categorias de primeiro nível da API Shopee usadas para autocomplete (ADR-0019).

## Requirements

### Requirement 1: Engine Core — createMachine

**User Story:** As a developer, I want a lightweight state machine factory function, so that I can model page behavior declaratively and test all transitions without DOM or API dependencies.

#### Acceptance Criteria

1. THE Engine SHALL export a `createMachine(config, effects)` function that returns a machine instance with `send(event)` method and reactive `state` property.
2. THE Engine SHALL accept a Machine_Config object declaring: `initialState`, `initialContext`, `states` (map of state names to transition rules), and `guards` (map of guard names to predicate functions).
3. WHEN `send(event)` is called with an event that has a valid transition from the current state, THE Engine SHALL update the Machine_State to the target state and apply context updates defined in the transition.
4. WHEN `send(event)` is called with an event that has no valid transition from the current state, THE Engine SHALL ignore the event and leave Machine_State unchanged.
5. WHEN a transition defines a guard, THE Engine SHALL evaluate the guard function with (context, event) and only execute the transition if the guard returns true.
6. WHEN a transition defines an effect, THE Engine SHALL invoke the effect function asynchronously after the state transition completes and dispatch the returned event(s) back into the machine.
7. THE Engine SHALL have zero external dependencies — no XState, no Svelte, no DOM APIs.
8. THE Engine SHALL expose the `state` property as a Svelte 5 compatible reactive value (using `$state` rune internally or a getter pattern that integrates with Svelte reactivity).

### Requirement 2: Machine States Definition

**User Story:** As a developer, I want explicitly defined states for the search page, so that every UI condition maps to exactly one machine state.

#### Acceptance Criteria

1. THE Machine_Config SHALL define the following states: `idle`, `searching`, `results`, `saving`, `loading_saved`, `error`.
2. WHILE in `idle` state, THE Engine SHALL accept events: DIGITAR, ADICIONAR_LOJA, MUDAR_FILTRO, MUDAR_FONTES, CARREGAR_SALVA, INICIALIZAR.
3. WHILE in `searching` state, THE Engine SHALL accept events: BUSCA_SUCESSO, BUSCA_ERRO, CANCELAR.
4. WHILE in `results` state, THE Engine SHALL accept events: DIGITAR, ADICIONAR_LOJA, REMOVER_LOJA, MUDAR_FILTRO, MUDAR_FONTES, SALVAR, CARREGAR_SALVA, LIMPAR.
5. WHILE in `saving` state, THE Engine SHALL accept events: SALVAR_SUCESSO, SALVAR_ERRO.
6. WHILE in `error` state, THE Engine SHALL accept events: DIGITAR, RETRY, LIMPAR.
7. WHEN the machine enters `searching` state, THE Engine SHALL set `context.loading` to true.
8. WHEN the machine exits `searching` state, THE Engine SHALL set `context.loading` to false.

### Requirement 3: Machine Context — Dados do Estado

**User Story:** As a developer, I want a single context object that holds all search-related data, so that state is never fragmented across multiple variables.

#### Acceptance Criteria

1. THE Machine_Context SHALL contain the following fields: `keyword` (string), `shopIds` (string array), `shopNomes` (object map id→name), `comissaoMin` (number), `vendasMin` (number), `categorias` (string array), `fontes` (object with boolean flags), `cron` (string or null), `resultados` (array), `buscasSalvas` (array), `loading` (boolean), `error` (string or null).
2. THE Engine SHALL initialize Machine_Context with default values: keyword empty, shopIds empty, comissaoMin 0.07, vendasMin 0, fontes with curadoria+quedas+novos active, loading false, error null.
3. WHEN a transition specifies context updates, THE Engine SHALL apply updates immutably — creating a new context object rather than mutating the existing one.
4. THE Engine SHALL never allow direct mutation of Machine_Context from outside — context changes happen exclusively through transitions triggered by events.

### Requirement 4: Guards — Invariantes de Negócio

**User Story:** As a developer, I want guards that enforce business invariants on every transition, so that invalid states are impossible to reach.

#### Acceptance Criteria

1. THE Engine SHALL define a guard `temContextoBusca` that returns true only when `keyword.trim().length > 0` OR `shopIds.length > 0` — preventing empty searches without context.
2. THE Engine SHALL define a guard `comissaoValida` that returns true only when `comissaoMin` is a decimal between 0 and 1 — preventing raw percentage values like 7.0 from entering context.
3. THE Engine SHALL define a guard `salvarCompleto` that returns true only when the current context has at least one keyword or one shopId — preventing saves of empty configurations.
4. WHEN a MUDAR_FILTRO event sets comissaoMin, THE Engine SHALL normalize the value: if the value is greater than 1, THE Engine SHALL divide by 100 before storing in context.
5. WHEN the guard `temContextoBusca` fails on a DIGITAR event (keyword cleared and no shops), THE Engine SHALL transition to `idle` instead of `searching`.

### Requirement 5: Busca com Loja — Keyword + ShopIds

**User Story:** As a afiliado, I want adding a shop to automatically search within that shop using my current keyword, so that I see relevant products from the specific seller.

#### Acceptance Criteria

1. WHEN an ADICIONAR_LOJA event succeeds (loja resolvida), THE Engine SHALL add the shopId to `context.shopIds` and the shop name to `context.shopNomes`.
2. WHEN `context.keyword` is non-empty AND a new shopId is added, THE Engine SHALL automatically trigger a search combining keyword + shopIds in the API request.
3. WHEN `context.shopIds` contains one or more IDs, THE Effects_Module SHALL include `shop_ids` parameter in the API call to `/api/candidatos`.
4. WHEN a shop is removed (REMOVER_LOJA event), THE Engine SHALL remove the shopId from context and trigger a new search with the remaining parameters.
5. IF the shop resolution fails (API error), THEN THE Engine SHALL set `context.lojaErro` with the error message and remain in the current state without adding any shopId.

### Requirement 6: Formatação de Filtros Numéricos

**User Story:** As a afiliado, I want commission values always displayed correctly (e.g., "7%" not "7.0000000"), so that I can trust the filter UI.

#### Acceptance Criteria

1. THE Engine SHALL store `comissaoMin` as a decimal fraction (0.07 for 7%) — never as a raw percentage integer or float with excessive precision.
2. WHEN a MUDAR_FILTRO event carries a comissaoMin value greater than 1, THE Engine SHALL divide the value by 100 before storing it in context.
3. WHEN a MUDAR_FILTRO event carries a comissaoMin value between 0 and 1, THE Engine SHALL store it directly without transformation.
4. THE Engine SHALL round `comissaoMin` to at most 4 decimal places when storing in context.
5. WHEN results are filtered client-side by commission, THE Engine SHALL compare using the decimal fraction value from context without any further conversion.

### Requirement 7: Salvar e Carregar Busca Salva

**User Story:** As a afiliado, I want saving a search to capture ALL current state and loading it to restore EVERYTHING, so that I never lose parameters between save and load.

#### Acceptance Criteria

1. WHEN a SALVAR event is dispatched, THE Engine SHALL capture the complete context snapshot: keyword, shopIds, shopNomes, comissaoMin, vendasMin, categorias, fontes, cron.
2. WHEN a SALVAR event is dispatched, THE Effects_Module SHALL send a POST to /api/buscas with ALL fields from the context snapshot — never a partial payload.
3. WHEN a CARREGAR_SALVA event is dispatched with a saved search config, THE Engine SHALL replace the ENTIRE active context with the saved values — including restoring keyword to the input field.
4. WHEN a CARREGAR_SALVA event restores context, THE Engine SHALL automatically trigger a search execution with the restored parameters.
5. THE Engine SHALL generate the chip label using `gerarLabelBusca(config)` which includes the actual keyword text — never showing "(sem keywords)" when a keyword exists in the saved config.

### Requirement 8: Categorias Autocomplete

**User Story:** As a afiliado, I want category autocomplete suggesting Shopee first-level categories, so that I can filter by category without memorizing names.

#### Acceptance Criteria

1. WHEN the machine initializes (INICIALIZAR event), THE Effects_Module SHALL fetch the list of Shopee first-level categories and store them in `context.categoriasDisponiveis`.
2. WHEN the user types in the category input, THE BuscaUnificada SHALL filter `context.categoriasDisponiveis` by prefix match and display matching suggestions.
3. WHEN a category is selected from autocomplete, THE Engine SHALL add it to `context.categorias` array via a MUDAR_FILTRO event.
4. WHEN `context.categoriasDisponiveis` is empty (fetch failed), THE BuscaUnificada SHALL allow free-text category entry without autocomplete.

### Requirement 9: Fontes Assíncronas — Novos e Quedas

**User Story:** As a afiliado, I want the "Novos" and "Quedas" toggles to show data from scheduled searches, so that I can see fresh results after collection runs.

#### Acceptance Criteria

1. WHEN fontes `novos` or `quedas` are active, THE Effects_Module SHALL fetch novelty data from all monitored shops via /api/lojas/novidades.
2. WHEN the BUSCA_SUCESSO event arrives with results, THE Engine SHALL merge results from all active sources (curadoria, quedas, novos, lojas, favoritos) into `context.resultados`.
3. WHEN a fonte toggle changes (MUDAR_FONTES event), THE Engine SHALL re-fetch data for newly activated sources and filter out deactivated sources from results.
4. THE Engine SHALL never show results from a fonte that is toggled off — even if cached data exists from a previous fetch.
5. WHEN the user activates fonte `novos` and no cached data exists, THE Effects_Module SHALL trigger a fresh fetch rather than showing empty results.

### Requirement 10: Effects Module — Injeção de Dependências

**User Story:** As a developer, I want all API calls isolated in an injectable effects module, so that I can test the engine with mocks and swap implementations easily.

#### Acceptance Criteria

1. THE Effects_Module SHALL export an object with named effect functions: `executarBusca`, `salvarBusca`, `carregarBuscasSalvas`, `resolverLoja`, `carregarCategorias`, `carregarNovidades`.
2. THE `createMachine` function SHALL accept the effects object as second parameter — enabling injection of mock effects for testing.
3. WHEN an effect function is invoked, THE Engine SHALL pass `(context, event)` as arguments so effects have access to current state.
4. WHEN an effect function resolves, THE Engine SHALL dispatch the returned event (success or error) back into the machine.
5. IF an effect function rejects (throws), THEN THE Engine SHALL dispatch an error event with the error message and transition to `error` state.
6. THE Effects_Module SHALL never import Svelte modules or DOM APIs — effects operate on plain data and return plain data.

### Requirement 11: Integração Svelte 5 — View Pura

**User Story:** As a developer, I want the BuscaUnificada component to be a pure view that only reads machine state and dispatches events, so that all logic lives in the testable engine.

#### Acceptance Criteria

1. THE BuscaUnificada component SHALL create the machine instance using `createMachine(buscaConfig, effects)` and expose `machine.state` for template rendering.
2. THE BuscaUnificada component SHALL dispatch events to the machine via `machine.send(event)` — never modifying state directly.
3. THE BuscaUnificada component SHALL derive all UI values (resumo, labels, badge counts) from `machine.state.context` using `$derived` runes.
4. THE BuscaUnificada component script block SHALL contain zero business logic — only event dispatch bindings and derived UI computations.
5. THE Engine `state` property SHALL be compatible with Svelte 5 reactivity — changes to state SHALL trigger component re-renders automatically.

### Requirement 12: Testes Unitários da Engine

**User Story:** As a developer, I want comprehensive unit tests for every transition and guard, so that I can verify all bug fixes without running the full app.

#### Acceptance Criteria

1. THE test suite SHALL verify each state transition by sending events and asserting the resulting state name and context values.
2. THE test suite SHALL verify that guards block invalid transitions — asserting that state remains unchanged when a guard fails.
3. THE test suite SHALL verify the complete scenario: DIGITAR "serum" → ADICIONAR_LOJA "Le Botanic" → assert search executes with keyword+shopId combined.
4. THE test suite SHALL verify that MUDAR_FILTRO with comissaoMin=7 normalizes to 0.07 in context.
5. THE test suite SHALL verify that SALVAR captures full context and CARREGAR_SALVA restores it completely including keyword.
6. THE test suite SHALL use mock effects (never real API calls) and run via Vitest.
7. THE test suite SHALL include a round-trip property test: for any valid context, `salvar(context)` then `carregar(saved)` SHALL produce an equivalent context.

### Requirement 13: Debounce e Cancelamento

**User Story:** As a afiliado, I want typing to trigger search only after I pause, and new typing to cancel pending searches, so that I don't flood the API with requests.

#### Acceptance Criteria

1. WHEN a DIGITAR event is received, THE Engine SHALL start a 400ms debounce timer before transitioning to `searching` state.
2. WHEN a new DIGITAR event arrives before the debounce timer expires, THE Engine SHALL reset the timer to 400ms from the new event.
3. WHEN the debounce timer expires and the guard `temContextoBusca` passes, THE Engine SHALL transition to `searching` and invoke the `executarBusca` effect.
4. WHEN the debounce timer expires and the guard `temContextoBusca` fails, THE Engine SHALL transition to `idle` and clear results.
5. WHEN a non-debounced parameter change occurs (ADICIONAR_LOJA, REMOVER_LOJA), THE Engine SHALL trigger search immediately without debounce delay.

### Requirement 14: Bypass de Autenticação para Testes

**User Story:** As a developer, I want a test bypass mechanism for authentication, so that I can run Playwright E2E tests without Firebase Emulator dependency.

#### Acceptance Criteria

1. WHEN the environment variable `SKIP_AUTH` is set to "true" OR the query parameter `?test=1` is present, THE application SHALL inject a fixed test user without requiring Firebase authentication.
2. THE test bypass SHALL only be available in non-production environments (development and test modes).
3. WHEN the test bypass is active, THE application SHALL use a hardcoded user object with a known tenant_id, allowing API calls to proceed normally.
4. THE Playwright test fixture `authedPage` SHALL use this bypass mechanism as the primary authentication strategy, falling back to Firebase Emulator only when the bypass is unavailable.

### Requirement 15: Migração Incremental — Compatibilidade

**User Story:** As a developer, I want the engine to coexist with existing code during migration, so that I can migrate incrementally without breaking the 174+ existing tests.

#### Acceptance Criteria

1. THE Engine module (`busca-engine.js`) SHALL be a new file — not modifying the existing `BuscaUnificada.svelte.js` until migration is complete.
2. THE Engine SHALL reuse existing pure functions from `busca-unificada-logic.js` (configToPayload, payloadToConfig, gerarResumo, gerarLabelBusca) — not duplicating them.
3. THE Effects_Module SHALL delegate to existing API functions from `api.js` and `descobrir.js` — preserving current endpoint contracts.
4. WHEN the engine is integrated into BuscaUnificada.svelte, THE existing tests that depend on the component behavior SHALL continue to pass.
5. THE Engine SHALL respect ESLint rules: max-lines-per-function 80 for .js files, using multiple small functions composed together.
