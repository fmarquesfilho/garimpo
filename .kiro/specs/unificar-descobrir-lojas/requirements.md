# Requirements Document

## Introduction

Unificação das páginas "Descobrir" (`/`) e "Lojas" (`/lojas`) em uma única página "Garimpar" (`/`). A separação atual causa confusão ao usuário sobre onde buscar produtos e onde gerenciar monitoramento de lojas. O conceito central de ambas é o mesmo — "encontrar produtos para publicar" — e devem coexistir em uma única interface coesa.

O backend (C#, Go, Python) permanece inalterado. Todos os endpoints existentes (`/api/lojas`, `/api/buscas`, `/api/candidatos`, `/api/lojas/novidades`) continuam funcionando. A mudança é exclusivamente frontend (SvelteKit 2, Svelte 5, Tailwind CSS v4).

## Glossary

- **Página_Garimpar**: A página unificada servida na rota `/`, substituindo as antigas páginas "Descobrir" e "Lojas"
- **Fonte**: Origem de dados de produtos no sistema de busca — curadoria, quedas, novos, favoritos, ou lojas
- **Toggle_Fonte**: Botão que ativa ou desativa uma fonte de dados na área de busca
- **Fonte_Lojas**: Nova fonte de dados "🏪 Lojas" que exibe produtos das lojas monitoradas diretamente nos resultados unificados
- **FilterBar**: Componente de filtros (busca por keyword, comissão mínima, vendas mínimas, categorias)
- **Área_Configuração**: Seção colapsável na parte inferior da página unificada contendo FormAdicionarLoja, GerenciarBuscas e PainelAlertas
- **FormAdicionarLoja**: Componente para adicionar novas lojas ao monitoramento via URL/ID
- **GerenciarBuscas**: Componente para criar e gerenciar buscas agendadas por keyword
- **PainelAlertas**: Componente para configurar alertas de preço via Telegram
- **NavDrawer**: Menu lateral de navegação do aplicativo
- **Loja_Monitorada**: Uma loja Shopee registrada para coleta periódica de produtos pelo Scheduler
- **ProductCard**: Componente visual que exibe um produto individual nos resultados
- **montarResultados**: Função pura de filtragem que combina dados de múltiplas fontes e aplica filtros client-side

## Requirements

### Requirement 1: Nova fonte "Lojas" no sistema de busca

**User Story:** As a affiliate marketer, I want to see monitored store products directly in the unified search results, so that I can find products to publish without navigating to a separate page.

#### Acceptance Criteria

1. WHEN the Página_Garimpar loads, THE Toggle_Fonte section SHALL display a "🏪 Lojas" toggle alongside the existing toggles (🔍 Busca, 📉 Quedas, 🆕 Novos, ⭐ Favoritos)
2. WHEN the Fonte_Lojas toggle is active, THE montarResultados function SHALL include products from all Loja_Monitorada instances in the unified results grid
3. WHEN the Fonte_Lojas toggle is active AND a specific Loja_Monitorada is selected as filter, THE montarResultados function SHALL include only products from that specific store
4. WHEN the Fonte_Lojas toggle is inactive, THE montarResultados function SHALL exclude store products from the results
5. THE montarResultados function SHALL tag each store product with `_fonte: 'loja'` to distinguish from other sources
6. WHEN keyword and category filters are active, THE montarResultados function SHALL apply them as AND conditions to store products (same behavior as other sources)

### Requirement 2: Integração dos componentes de gerenciamento na Área_Configuração

**User Story:** As a affiliate marketer, I want to manage stores, scheduled searches, and price alerts from the same page where I search for products, so that I have a single workflow for product discovery.

#### Acceptance Criteria

1. THE Página_Garimpar SHALL display an Área_Configuração section below the results grid
2. THE Área_Configuração SHALL contain FormAdicionarLoja, GerenciarBuscas, and PainelAlertas components in this order
3. THE Área_Configuração SHALL be collapsible via a toggle button, defaulting to collapsed state
4. WHEN the Área_Configuração is expanded, THE Página_Garimpar SHALL render the same FormAdicionarLoja component currently in `/lojas` with identical functionality (URL/ID input, origin selector, keywords, scheduling)
5. WHEN the Área_Configuração is expanded, THE Página_Garimpar SHALL render the same GerenciarBuscas component currently in `/lojas` with identical functionality (create/remove keyword searches with scheduling)
6. WHEN the Área_Configuração is expanded, THE Página_Garimpar SHALL render the same PainelAlertas component currently in `/lojas` with identical functionality (Telegram alert configuration)
7. WHEN a new Loja_Monitorada is added via FormAdicionarLoja, THE Página_Garimpar SHALL update the available store list for the Fonte_Lojas filter without requiring page reload

### Requirement 3: Seleção de loja como filtro nos resultados

**User Story:** As a affiliate marketer, I want to filter store products by a specific monitored store, so that I can focus on products from one store at a time.

#### Acceptance Criteria

1. WHEN the Fonte_Lojas toggle is active AND at least one Loja_Monitorada exists, THE Página_Garimpar SHALL display a store selector showing all monitored stores
2. WHEN a specific Loja_Monitorada is selected in the store selector, THE montarResultados function SHALL filter store-source products to only those from the selected store
3. WHEN "all stores" is selected (default), THE montarResultados function SHALL include products from all monitored stores
4. THE store selector SHALL display the store name and a badge with the product count from that store
5. WHEN no Loja_Monitorada exists AND the Fonte_Lojas toggle is active, THE Página_Garimpar SHALL display an empty state message guiding the user to add a store via the Área_Configuração

### Requirement 4: Remoção da rota `/lojas` com redirecionamento

**User Story:** As a user navigating to the old stores URL, I want to be redirected to the unified page, so that I don't encounter a 404 error or stale content.

#### Acceptance Criteria

1. WHEN a user navigates to `/lojas`, THE application SHALL redirect to `/` with HTTP status 308 (permanent redirect)
2. THE NavDrawer SHALL remove the "🏪 Lojas" link from the "Configurações" section
3. THE NavDrawer SHALL rename the "🔍 Descobrir" link to "🔍 Garimpar"
4. WHEN an internal reference to `/lojas` exists in the application (empty states, hints), THE Página_Garimpar SHALL update those references to point to the Área_Configuração section on `/`

### Requirement 5: Preservação da lógica existente de montarResultados

**User Story:** As a developer, I want the existing filtering logic and tests to remain functional, so that the unification does not introduce regressions.

#### Acceptance Criteria

1. THE montarResultados function SHALL accept an optional `dadosLojas` parameter (array) in addition to existing parameters
2. THE montarResultados function SHALL continue to produce identical output for existing parameters when `dadosLojas` is empty or absent
3. WHEN the `fontes.lojas` flag is true, THE montarResultados function SHALL include items from `dadosLojas` in the combined results before applying keyword, category, and numeric filters
4. THE existing 40+ unit tests in `descobrir.test.js` SHALL continue to pass without modification after the change
5. THE montarResultados function signature extension SHALL be backward-compatible (existing callers without `dadosLojas` produce same results)

### Requirement 6: Carregamento de produtos das lojas monitoradas

**User Story:** As a affiliate marketer, I want monitored store products to load automatically when the "Lojas" source is active, so that I see relevant products without extra steps.

#### Acceptance Criteria

1. WHEN the Fonte_Lojas toggle is activated, THE Página_Garimpar SHALL call the existing `buscarCandidatos` API with `fonte: 'shopee-shop'` for each monitored store
2. WHEN the Fonte_Lojas toggle is activated AND multiple stores are monitored, THE Página_Garimpar SHALL load products from all stores in parallel
3. IF a store product request fails, THEN THE Página_Garimpar SHALL display products from successful requests and show a non-blocking warning for the failed store
4. WHILE store products are loading, THE Página_Garimpar SHALL display a loading indicator within the results area
5. THE Página_Garimpar SHALL cache loaded store products for 2 minutes (same cache strategy as oportunidades) to avoid redundant API calls on rapid toggle changes

### Requirement 7: Título e identidade visual da página unificada

**User Story:** As a affiliate marketer, I want the unified page to have a clear identity that communicates its unified purpose, so that the mental model of the app is simpler.

#### Acceptance Criteria

1. THE Página_Garimpar SHALL display the title "O que publicar hoje?" as the main heading (preserving existing title from Descobrir)
2. THE Página_Garimpar SHALL set the browser tab title to "Garimpar — Garimpei"
3. THE Página_Garimpar SHALL display a subtitle that mentions both product search and store monitoring capabilities
4. WHEN the user is not authenticated, THE Página_Garimpar SHALL display only the FilterBar and source toggles with a login prompt instead of results
