# Requirements Document

## Introduction

Substituir a seção separada de scope cards (LojaCard/CategoriaCard abaixo do Omnibox) por chips inline renderizados **dentro** do campo de input do Omnibox. Cada chip representa uma loja ou categoria ativa no escopo da busca, com cores diferenciadas para tipo (loja vs. categoria) e um botão X para remoção sincronizada com o engine state. Isso unifica visualmente os controles de escopo com o campo de busca, reforçando que filtros, keywords e escopo são um único componente integrado.

## Glossary

- **Omnibox**: Componente de input unificado que concentra busca por keywords, lojas, categorias e marketplaces
- **Chip**: Pílula visual inline dentro do campo de input, representando uma loja ou categoria ativa no escopo
- **Engine**: BuscaEngine, controlador headless que detém todo o estado reativo e despacha eventos
- **Loja_Chip**: Chip de tipo loja, usando a variante de cor `ouro` da paleta quente
- **Categoria_Chip**: Chip de tipo categoria, usando a variante de cor `rosa/terracota` da paleta quente
- **Scope**: Conjunto de lojas e categorias que delimitam a busca atual (engine.ctx.shopIds + engine.ctx.categorias)
- **Badge**: Componente reutilizável de pílula com variantes semânticas de cor

## Requirements

### Requirement 1: Renderizar chips de loja dentro do input

**User Story:** As a user, I want to see scoped shops as colored chips inside the search input, so that I can immediately understand which shops are active in my search context.

#### Acceptance Criteria

1. WHEN a shop is added to the scope, THE Omnibox SHALL render a Loja_Chip inline within the input container representing that shop
2. THE Loja_Chip SHALL display the shop name as text content
3. THE Loja_Chip SHALL use a visually distinct color derived from the `ouro` palette tokens (background: `--ouro-fundo`, text: `--ouro-escuro`, border: `--ouro-claro`)
4. THE Loja_Chip SHALL include a prefix icon (🏪) to reinforce its type visually
5. WHEN multiple shops are in scope, THE Omnibox SHALL render one Loja_Chip per shop, in the order they were added

### Requirement 2: Renderizar chips de categoria dentro do input

**User Story:** As a user, I want to see active categories as colored chips inside the search input, so that I can distinguish them from shop chips and understand the full search context at a glance.

#### Acceptance Criteria

1. WHEN a category is added to the scope, THE Omnibox SHALL render a Categoria_Chip inline within the input container representing that category
2. THE Categoria_Chip SHALL display the category name as text content
3. THE Categoria_Chip SHALL use a visually distinct color derived from the `rosa/terracota` palette tokens (background: `--rosa-fundo`, text: `--rosa`, border: current border-border)
4. THE Categoria_Chip SHALL include a prefix icon (🏷️) to reinforce its type visually
5. WHEN multiple categories are in scope, THE Omnibox SHALL render one Categoria_Chip per category, in the order they were added

### Requirement 3: Remover chips via botão X sincronizado com o engine

**User Story:** As a user, I want to remove a shop or category from the search scope by clicking the X on its chip, so that I can refine my search without leaving the input area.

#### Acceptance Criteria

1. THE Loja_Chip SHALL display a clickable X button to remove the shop from scope
2. WHEN the user clicks the X on a Loja_Chip, THE Omnibox SHALL dispatch a REMOVER_LOJA event to the Engine with the corresponding shopId
3. THE Categoria_Chip SHALL display a clickable X button to remove the category from scope
4. WHEN the user clicks the X on a Categoria_Chip, THE Omnibox SHALL dispatch a REMOVER_CATEGORIA event to the Engine with the corresponding category name
5. WHEN the Engine processes the removal event, THE Omnibox SHALL immediately remove the corresponding chip from the input container

### Requirement 4: Eliminar a seção separada de scope cards

**User Story:** As a user, I want the search context to be fully integrated within the input field, so that I have a single unified control for managing my search scope.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL remove the separate "Escopo ativo" section that renders LojaCard and CategoriaCard components below the Omnibox
2. WHEN shops or categories are in scope, THE Omnibox SHALL be the sole visual indicator of active scope (via inline chips)
3. THE BuscaUnificada SHALL retain the LojaCard and CategoriaCard component files for potential reuse in other contexts

### Requirement 5: Layout do input com chips inline

**User Story:** As a user, I want the chips to appear inside the input field alongside my text, so that the input serves as both a search box and a scope indicator.

#### Acceptance Criteria

1. THE Omnibox SHALL render chips to the left of the text input cursor within the same container
2. THE Omnibox SHALL allow the text input area to shrink horizontally as chips occupy space
3. WHEN chips overflow the single visible line, THE Omnibox input container SHALL wrap to multiple lines
4. THE Omnibox SHALL maintain the search icon (🔍) as the leading element before all chips
5. WHILE chips are present, THE Omnibox text input SHALL remain focusable and functional for keyword entry

### Requirement 6: Acessibilidade dos chips

**User Story:** As a user relying on assistive technology, I want each chip to announce its type and state, so that I can understand and interact with the search scope without visual cues.

#### Acceptance Criteria

1. THE Loja_Chip SHALL expose an accessible label in the format "Loja: {nome} — ativa" to screen readers
2. THE Categoria_Chip SHALL expose an accessible label in the format "Categoria: {nome} — ativa" to screen readers
3. THE X button on each chip SHALL expose an aria-label in the format "Remover loja {nome}" or "Remover categoria {nome}"
4. WHEN a chip is removed, THE Omnibox SHALL announce the removal via an aria-live region (e.g., "Loja {nome} removida do escopo")
5. THE Omnibox chip container SHALL use role="group" with an aria-label "Filtros ativos" to identify the chip collection

### Requirement 7: Manter Omnibox como renderizador puro

**User Story:** As a developer, I want the Omnibox to remain a pure renderer with zero local state, so that the headless UI controller pattern is preserved.

#### Acceptance Criteria

1. THE Omnibox SHALL derive all chip data exclusively from the Engine state (engine.lojaCards, engine.categoriaCards)
2. THE Omnibox SHALL contain zero $state declarations for chip-related data
3. WHEN the user interacts with a chip, THE Omnibox SHALL dispatch events to the Engine without performing any state mutation locally
4. THE Engine SHALL expose chip rendering data as derived getters (existing lojaCards and categoriaCards) without requiring new state properties

### Requirement 8: Distinção visual perceptível entre tipos de chip

**User Story:** As a user, I want to instantly distinguish shop chips from category chips by color alone, so that I can parse the search context rapidly.

#### Acceptance Criteria

1. THE Loja_Chip and Categoria_Chip SHALL use colors with a minimum contrast ratio of 4.5:1 between text and background (WCAG AA)
2. THE Loja_Chip background color SHALL differ from the Categoria_Chip background color by at least 30° on the HSL hue wheel
3. THE Loja_Chip and Categoria_Chip SHALL each include a distinct icon prefix as a secondary visual differentiator beyond color
4. WHILE the system is in dark mode, THE chips SHALL use the corresponding dark-mode token values from the warm palette (--ouro-fundo dark, --rosa-fundo dark)

### Requirement 9: Garantia arquitetural — bloquear estado local em compilação

**User Story:** As a developer, I want the build/lint to fail if anyone adds $state to a pure-renderer component of the Descobrir page, so that the Headless UI Controller pattern cannot be violated by accident.

#### Acceptance Criteria

1. THE ESLint config SHALL include a `no-restricted-syntax` rule targeting `CallExpression[callee.name="$state"]` for the components: Omnibox.svelte, StoreCard.svelte, BuscaUnificada.svelte
2. WHEN a developer writes `$state(...)` in any of these components, THE lint SHALL fail with a message referencing ADR-0033 and explaining to use `engine.send()` instead
3. THE CI (`lint:js`) SHALL enforce this rule — pull requests with local state in pure-renderer components SHALL not pass
4. THE rule SHALL be scoped only to pure-renderer components of the Descobrir page (not to the engine class or other pages)
5. A mise task `check:pure-renderers` SHALL validate this constraint independently for fast feedback
