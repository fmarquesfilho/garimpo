# Requirements Document

## Introduction

O Garimpei possui atualmente três componentes separados que configuram variações de "busca": **FilterBar** (filtros real-time), **FormAdicionarLoja** (adicionar loja monitorada) e **GerenciarBuscas** (buscas agendadas por keyword). Esta feature unifica os três em um único componente **BuscaUnificada** que combina keywords, seleção de lojas (plural, multi-marketplace), filtros (comissão, vendas, categorias) e agendamento — tudo no mesmo painel posicionado no topo da página.

O componente executa a busca imediatamente quando filtros mudam (real-time) e oferece opção de salvar/agendar a configuração atual como busca persistente.

## Glossary

- **BuscaUnificada**: Componente Svelte unificado que substitui FilterBar, FormAdicionarLoja e GerenciarBuscas. Responsável por captura de keywords, seleção de lojas, filtros e agendamento.
- **Busca**: Entidade de domínio (C#) que persiste uma configuração de busca — keywords, shop IDs, cron, marketplaces.
- **Scheduler**: Serviço Go (gRPC) responsável por executar buscas agendadas via Cloud Tasks/Scheduler.
- **Collector**: Serviço Go (gRPC) responsável por coletar produtos de marketplaces (Shopee, Amazon, ML).
- **TagInput**: Sub-componente existente para entrada de tags (keywords, categorias, URLs).
- **AgendadorBusca**: Sub-componente existente para configuração visual de expressão cron.
- **ToggleGroup**: Sub-componente existente para seleção múltipla de opções (fontes, marketplaces).
- **SyncBuscaRequest**: Record do endpoint POST /api/buscas que aceita keywords, cron, sort_by e limit.
- **Marketplace**: Identificador de origem de produtos — "shopee", "amazon" ou "mercadolivre" (ADR-0016).
- **ShopIds**: Array de IDs numéricos (bigint[]) de lojas resolvidas pelo Collector.
- **CronExpression**: Expressão cron padrão (5 campos) que define frequência de execução periódica.
- **Filtros_De_Busca**: Conjunto de parâmetros opcionais aplicados à busca — comissão mínima, vendas mínimas e categorias.
- **Busca_Salva**: Uma configuração de busca persistida no banco de dados, com ou sem agendamento.

## Requirements

### Requirement 1: Entrada de Keywords

**User Story:** As a afiliado, I want to type keywords into the unified search component, so that I can search for products across marketplaces in real-time.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL display a single text input field for keyword entry at the top of the component.
2. WHEN the user types a keyword and pauses for 400ms, THE BuscaUnificada SHALL trigger a product search using the entered keywords.
3. WHEN the user presses Escape in the keyword input, THE BuscaUnificada SHALL clear the keyword field and remove focus from the input.
4. WHEN the user presses Enter in the keyword input, THE BuscaUnificada SHALL remove focus from the input and trigger the search immediately.
5. WHEN the keyword field contains text, THE BuscaUnificada SHALL display a clear button (✕) that resets the field to empty.

### Requirement 2: Seleção de Lojas (Plural, Multi-Marketplace)

**User Story:** As a afiliado, I want to select one or more shops from different marketplaces within the search component, so that I can scope my product search to specific sellers.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL provide a shop input field that accepts URLs or numeric IDs of shops.
2. WHEN the user adds a shop URL or ID, THE BuscaUnificada SHALL resolve the shop via the Collector gRPC service (ResolveShop) and display the resolved shop name as a tag.
3. THE BuscaUnificada SHALL allow adding multiple shops, including shops from different marketplaces (Shopee, Amazon, Mercado Livre).
4. WHEN no shops are selected, THE BuscaUnificada SHALL search across the entire marketplace catalog (keyword-only mode).
5. WHEN one or more shops are selected, THE BuscaUnificada SHALL restrict the search results to products from those shops only.
6. IF the Collector fails to resolve a shop URL, THEN THE BuscaUnificada SHALL display an inline error message with the failure reason and keep the input value for correction.
7. WHEN a shop tag is removed by the user, THE BuscaUnificada SHALL update the search scope immediately to exclude that shop.

### Requirement 3: Filtros de Busca (Comissão, Vendas, Categorias)

**User Story:** As a afiliado, I want to set minimum commission, minimum sales, and category filters within the search, so that I can narrow results to profitable products.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL provide filter controls for: minimum commission (select), minimum sales (numeric input), and category (text input with autocomplete).
2. THE BuscaUnificada SHALL display filter controls in a collapsible section toggled by a "⚙️ Filtros" button.
3. WHEN the filter section is collapsed and one or more non-default filters are active, THE BuscaUnificada SHALL display a badge on the toggle button showing the count of active filters.
4. WHEN any filter value changes, THE BuscaUnificada SHALL trigger a new search within 400ms (debounced).
5. THE BuscaUnificada SHALL load available categories from the API on mount and provide autocomplete suggestions as the user types.
6. WHEN the user saves a search, THE BuscaUnificada SHALL include all active filter values (comissão, vendas, categorias) as part of the saved configuration.

### Requirement 4: Execução Imediata (Real-Time)

**User Story:** As a afiliado, I want search results to update automatically when I change any parameter, so that I get instant feedback without clicking a "search" button.

#### Acceptance Criteria

1. WHEN any search parameter changes (keywords, shops, filters, sources), THE BuscaUnificada SHALL debounce for 400ms and then execute the search.
2. WHILE a search is in progress, THE BuscaUnificada SHALL display a loading indicator.
3. WHEN a new parameter change occurs while a previous search is pending, THE BuscaUnificada SHALL cancel the pending debounce and start a new 400ms wait.
4. THE BuscaUnificada SHALL emit search results to the parent page for rendering in the product grid.

### Requirement 5: Salvar Busca

**User Story:** As a afiliado, I want to save my current search configuration with a single click, so that I can reuse it later without re-entering all parameters.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL display a "💾 Salvar busca" button that persists the current configuration (keywords, shop IDs, filters, sources).
2. WHEN the user clicks "Salvar busca", THE BuscaUnificada SHALL send a POST request to /api/buscas with the complete search configuration including shop_ids, keywords, comissao_min, vendas_min, and categorias.
3. WHEN a search is saved successfully, THE BuscaUnificada SHALL display a success notification and add the saved search to the list of saved searches.
4. IF the save request fails, THEN THE BuscaUnificada SHALL display an error message without losing the current configuration state.
5. WHEN a saved search exists, THE BuscaUnificada SHALL allow the user to load it by clicking, restoring all parameters (keywords, shops, filters, sources) into the component.

### Requirement 6: Agendamento Opcional (Cron)

**User Story:** As a afiliado, I want to optionally schedule a saved search for periodic execution, so that I can receive fresh results automatically.

#### Acceptance Criteria

1. WHEN the user clicks "Salvar busca", THE BuscaUnificada SHALL display the AgendadorBusca sub-component allowing optional cron configuration before confirming the save.
2. WHEN a cron expression is provided, THE BuscaUnificada SHALL include it in the save request and the Scheduler SHALL register a periodic job for that search.
3. WHEN no cron expression is provided, THE BuscaUnificada SHALL save the search without scheduling (manual-only execution).
4. WHEN a saved search with a cron is loaded, THE BuscaUnificada SHALL display the active schedule indicator next to the search tag.
5. WHEN the user removes the cron from a saved search and re-saves, THE BuscaUnificada SHALL pause the corresponding Scheduler job.

### Requirement 7: Backend — Extensão do SyncBuscaRequest

**User Story:** As a developer, I want the POST /api/buscas endpoint to accept shop_ids and filter parameters, so that the unified frontend can persist complete search configurations.

#### Acceptance Criteria

1. THE SyncBuscaRequest SHALL accept an optional field `shop_ids` (long array) to persist shop associations with a saved search.
2. THE SyncBuscaRequest SHALL accept optional fields `comissao_min` (decimal), `vendas_min` (integer), and `categorias` (string array) to persist filter preferences.
3. WHEN `shop_ids` is provided in the request, THE BuscasEndpoints SHALL store the values in the Busca entity's ShopIds field.
4. WHEN filter fields are provided, THE BuscasEndpoints SHALL store them in corresponding fields on the Busca entity.
5. THE BuscasEndpoints SHALL maintain backward compatibility — existing requests without the new fields SHALL continue to work identically.

### Requirement 8: Backend — Extensão da Entidade Busca

**User Story:** As a developer, I want the Busca entity to store filter preferences, so that saved searches preserve the full user configuration.

#### Acceptance Criteria

1. THE Busca entity SHALL include nullable fields: `ComissaoMin` (decimal?), `VendasMin` (int?), and `Categorias` (string[]?).
2. WHEN a Busca is serialized for the GET /api/buscas response, THE BuscasEndpoints SHALL include the filter fields in the response payload.
3. THE new fields SHALL be nullable, and existing rows in the database SHALL default to null (no data migration required beyond schema addition).

### Requirement 9: Layout e Posicionamento

**User Story:** As a afiliado, I want the search component at the top of the page, always accessible, so that I can quickly adjust my search without scrolling.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL render at the top of the main page, above the product results grid.
2. THE BuscaUnificada SHALL be collapsible — the user can collapse it to maximize screen space for results.
3. WHILE the BuscaUnificada is collapsed, THE BuscaUnificada SHALL display a compact summary showing active keywords, shop count, and active filter count.
4. THE BuscaUnificada SHALL replace the existing FilterBar, FormAdicionarLoja, and GerenciarBuscas components in the page layout.

### Requirement 10: Buscas Salvas — Listagem Integrada

**User Story:** As a afiliado, I want to see my saved searches as quick-access tags below the search component, so that I can switch between searches with one click.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL display saved searches as clickable tags (chips) below the main input area.
2. WHEN the user clicks a saved search tag, THE BuscaUnificada SHALL load all parameters from that search into the active fields (keywords, shops, filters, sources, cron).
3. WHEN a saved search has an active cron, THE BuscaUnificada SHALL display a schedule indicator (⏱) next to the tag.
4. THE BuscaUnificada SHALL provide a way to delete a saved search from the tag list.
5. WHEN a saved search is deleted, THE BuscaUnificada SHALL call the appropriate API endpoint to deactivate it and pause its Scheduler job.

### Requirement 11: Compatibilidade com Lojas Existentes

**User Story:** As a afiliado, I want my existing monitored shops to appear in the unified search component, so that the migration does not lose my data.

#### Acceptance Criteria

1. WHEN the component mounts, THE BuscaUnificada SHALL load existing shops (buscas with ShopIds) from GET /api/lojas and display them as selectable entries.
2. THE BuscaUnificada SHALL support the existing "add new shop" flow (resolve via Collector, persist via POST /api/lojas) within the unified shop input.
3. WHEN adding a new shop that requires resolution, THE BuscaUnificada SHALL display a loading state on the shop tag until resolution completes.
4. THE BuscaUnificada SHALL preserve backward compatibility with the existing POST /api/lojas endpoint for shop resolution and creation.

### Requirement 12: Compatibilidade com Bits UI e shadcn-svelte

**User Story:** As a developer, I want the unified search component to follow the established Bits UI + shadcn-svelte patterns, so that the codebase remains consistent, accessible, and maintainable.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL use exclusively components from `$lib/components/ui/` (Button, Input, Select, Alert, Card, Badge, ToggleGroup, Collapsible) and Bits UI composites — zero inline `<button>`, `<select>`, or `<input>` elements outside of sub-components that already exist (TagInput, AgendadorBusca).
2. THE BuscaUnificada SHALL use Tailwind CSS utility classes for all styling, with zero CSS scoped blocks (`<style>` sections).
3. THE BuscaUnificada SHALL use the `cn()` utility from `$lib/utils.ts` for conditional class merging.
4. THE BuscaUnificada SHALL use design tokens from `tokens.css` exclusively — zero hardcoded hex colors or arbitrary color values.
5. THE BuscaUnificada SHALL follow the shadcn-svelte component API pattern: accept `variant`, `size`, `class`, and `...rest` props where applicable for internal sub-sections.
6. THE BuscaUnificada SHALL comply with WCAG accessibility requirements provided by Bits UI: keyboard navigation on all interactive elements, proper ARIA attributes on select/toggle/collapsible controls, and visible focus indicators.
7. THE BuscaUnificada SHALL pass `svelte-check` with zero errors, `stylelint color-no-hex` with zero violations, and `eslint --max-warnings=0` without new warnings.
8. THE BuscaUnificada script block SHALL respect the 180-line limit enforced by ESLint — logic exceeding this limit SHALL be extracted to a dedicated `.svelte.ts` or `.ts` module.

### Requirement 13: Seleção de Fontes de Dados

**User Story:** As a afiliado, I want to choose which data sources to include in my search (curadoria, quedas, novos, lojas, favoritos), so that I can focus on the type of opportunities I care about.

#### Acceptance Criteria

1. THE BuscaUnificada SHALL provide a ToggleGroup for selecting active data sources: curadoria, quedas de preço, produtos novos, lojas, favoritos.
2. WHEN any source toggle changes, THE BuscaUnificada SHALL trigger a new search including only the active sources.
3. WHEN the user saves a search, THE BuscaUnificada SHALL persist the selected sources as part of the saved configuration.
4. WHEN a saved search is loaded, THE BuscaUnificada SHALL restore the source selection to match the saved configuration.

