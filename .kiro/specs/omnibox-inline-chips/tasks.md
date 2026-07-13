# Tasks

## Dependency Graph

```
Wave 1 (foundation — no cross-deps):
  T1 Badge variants ──────────┐
  T2 ESLint rule ─────────────┤
  T3 mise check task ──────── │ ─── all independent
                               │
Wave 2 (feature — depends on Wave 1):
  T4 Rewrite Omnibox ─────── depends on T1
  T5 Remove scope cards ──── depends on T4
                               │
Wave 3 (validation — depends on Wave 2):
  T6 Component tests ──────── depends on T4, T5
  T7 E2E updates ─────────── depends on T4, T5
  T8 A11y audit ───────────── depends on T4
```

---

## Wave 1 — Foundation

### Task 1: Add `loja` and `categoria` Badge variants

**References:** Req 1.3, Req 2.3, Req 8.1, Req 8.2, Req 8.3

**File:** `web/src/lib/components/ui/Badge.svelte`

- [x] Add `loja` entry to VARIANTS map: `bg-[var(--ouro-fundo)] text-[var(--ouro-escuro)] border-[var(--ouro-claro)]`
- [x] Add `categoria` entry to VARIANTS map: `bg-[var(--rosa-fundo)] text-[var(--rosa)] border-border`
- [x] Update the `@prop variant` JSDoc to include the new values
- [x] Verify contrast ≥4.5:1 for both variants in light and dark mode (manual check with devtools)

**Acceptance:** Badge renders correctly with `variant="loja"` and `variant="categoria"` using the warm palette tokens.

---

### Task 2: Add ESLint `no-restricted-syntax` rule for pure-renderer components

**References:** Req 9.1, Req 9.2, Req 9.3, Req 9.4

**File:** `web/eslint.config.js`

- [x] Add a new config block scoped to `['**/Omnibox.svelte', '**/StoreCard.svelte', '**/BuscaUnificada.svelte']`
- [x] Rule: `'no-restricted-syntax': ['error', { selector: 'CallExpression[callee.name="$state"]', message: '❌ $state proibido em pure-renderer (ADR-0033). Use engine.send() para mutar estado.' }]`
- [x] Run `bun run lint:js` to confirm it passes on current code (no existing violations)
- [x] Add a temporary `$state()` in Omnibox.svelte, verify lint fails, then remove it

**Acceptance:** `bun run lint:js` enforces the rule. Adding `$state()` to any of the 3 files triggers an error referencing ADR-0033.

---

### Task 3: Create mise task `check:pure-renderers`

**References:** Req 9.5

**File:** `.mise/tasks/check/pure-renderers`

- [x] Create shell script following existing check task conventions (shebang, `#MISE description=`, set -uo pipefail)
- [x] Grep for `\$state(` in `Omnibox.svelte`, `StoreCard.svelte`, `BuscaUnificada.svelte`
- [x] Exit 0 if no matches, exit 1 with descriptive message if found
- [x] Include reference to ADR-0033 in the error output
- [x] Make executable: `chmod +x .mise/tasks/check/pure-renderers`
- [x] Verify: `mise run check:pure-renderers` exits 0

**Acceptance:** `mise run check:pure-renderers` passes cleanly. Artificially adding `$state()` causes it to fail with actionable message.

---

## Wave 2 — Feature Implementation

### Task 4: Rewrite Omnibox.svelte with inline chips

**References:** Req 1.1–1.5, Req 2.1–2.5, Req 3.1–3.5, Req 5.1–5.5, Req 6.1–6.5, Req 7.1–7.4

**File:** `web/src/lib/components/Omnibox.svelte`

- [x] Import `Badge` from `$lib/components/ui`
- [x] Add `$derived` for `engine.lojaCards` and `engine.categoriaCards`
- [x] Replace the outer `<div class="relative">` with a flex-wrap container: `role="group"`, `aria-label="Filtros ativos"`, focus-within styles
- [x] Render 🔍 as the leading flex item (shrink-0)
- [x] Render Loja_Chips: `{#each lojas as l (l.id)}` → `<Badge variant="loja">` with 🏪 prefix, nome text, and ✕ button dispatching `REMOVER_LOJA`
- [x] Render Categoria_Chips: `{#each categorias as c (c.nome)}` → `<Badge variant="categoria">` with 🏷️ prefix, nome text, and ✕ button dispatching `REMOVER_CATEGORIA`
- [x] Change `<input>` to `class="min-w-[120px] flex-1 border-none bg-transparent outline-none"` (unstyled, grows to fill)
- [x] Add `aria-label` to each chip: `"Loja: {nome} — ativa"` / `"Categoria: {nome} — ativa"`
- [x] Add `aria-label` to each ✕ button: `"Remover loja {nome}"` / `"Remover categoria {nome}"`
- [x] Add a second `<span aria-live="polite">` for chip removal announcements
- [x] Implement a `$derived` or `$effect` that sets the removal message text when a chip disappears (using engine state, no local $state)
- [x] Ensure zero `$state` declarations — verify with `bun run lint:js`
- [x] Keep the dropdown `<ul role="listbox">` positioned below the new container

**Acceptance:** Chips render inline, ✕ removes them via engine events, input remains functional, lint passes, screen reader announces removals.

---

### Task 5: Remove scope cards section from BuscaUnificada.svelte

**References:** Req 4.1, Req 4.2, Req 4.3

**File:** `web/src/lib/components/BuscaUnificada.svelte`

- [x] Remove the `{#if engine.lojaCards.length || engine.categoriaCards.length}` block that renders LojaCard/CategoriaCard
- [x] Remove the `LojaCard` and `CategoriaCard` imports (files are retained for reuse but no longer imported here)
- [x] Verify the remaining component still renders: omnibox, filters, marketplaces, buscas salvas, results
- [x] Run `bun run lint:js` — confirm no unused-import warnings or errors

**Acceptance:** BuscaUnificada no longer renders the "Escopo ativo" section. Omnibox is the sole scope indicator. `LojaCard.svelte` and `CategoriaCard.svelte` files remain in the codebase.

---

## Wave 3 — Validation

### Task 6: Component tests (vitest + @testing-library/svelte)

**References:** Req 1.1, Req 2.1, Req 3.2, Req 3.4, Req 5.3, Req 7.2

**Files:** `web/src/lib/components/__tests__/Omnibox.test.js` (new), `web/src/lib/components/__tests__/Badge.test.js` (extend)

- [x] Badge tests: render `variant="loja"` → assert ouro classes applied; render `variant="categoria"` → assert rosa classes
- [x] Omnibox test: mock engine with `lojaCards: [{ id: '1', nome: 'MegatronShop' }]` → assert chip renders with text "MegatronShop"
- [x] Omnibox test: click ✕ on loja chip → assert `engine.send` called with `{ type: 'REMOVER_LOJA', shopId: '1' }`
- [x] Omnibox test: mock engine with `categoriaCards: [{ nome: 'Eletrônicos', marketplaces: [] }]` → assert categoria chip renders
- [x] Omnibox test: click ✕ on categoria chip → assert `engine.send` called with `{ type: 'REMOVER_CATEGORIA', nome: 'Eletrônicos' }`
- [x] Omnibox test: verify no `$state` in rendered component (structural assertion — source grep)
- [x] Run: `bun run test -- --run src/lib/components/__tests__/Omnibox.test.js`

**Acceptance:** All tests pass. Coverage confirms chip rendering and removal dispatch.

---

### Task 7: E2E test updates (Playwright)

**References:** Req 3.5, Req 4.1, Req 5.1

**Files:** relevant Playwright spec files for the Descobrir page

- [x] Update existing selectors that target the old scope cards section (if any exist)
- [x] Add assertion: after adding a loja via omnibox, a chip with the loja name is visible inside the omnibox container
- [x] Add assertion: clicking ✕ on a chip removes it from the omnibox
- [x] Add assertion: the old "Escopo ativo" section no longer exists in the DOM
- [x] Verify tests pass locally: `bun run test:e2e` (not added to CI per project rules)

**Acceptance:** Playwright tests pass locally reflecting the new inline chip behavior.

---

### Task 8: Accessibility audit

**References:** Req 6.1–6.5, Req 8.1, Req 8.4

- [x] Manual audit: navigate Omnibox with screen reader (VoiceOver), verify chip announcements
- [x] Verify `role="group"` and `aria-label="Filtros ativos"` on chip container
- [x] Verify each chip has correct `aria-label` ("Loja: {nome} — ativa")
- [x] Verify ✕ button has `aria-label` ("Remover loja {nome}")
- [x] Verify `aria-live` announces "Loja {nome} removida do escopo" on removal
- [x] Check contrast ratio ≥4.5:1 for both chip variants in light and dark mode using browser devtools
- [x] Document any issues found and create follow-up tasks if needed

**Acceptance:** Chips are fully navigable and announced correctly by assistive technology. Contrast ratios meet WCAG AA.
