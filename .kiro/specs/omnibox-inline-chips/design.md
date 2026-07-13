# Design Document: Omnibox Inline Chips

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  BuscaEngine (headless controller)                                  │
│                                                                     │
│  ctx.shopIds ──► get lojaCards ─────┐                               │
│  ctx.categorias ──► get categoriaCards ─┐                           │
│                                         │                           │
│  send({ REMOVER_LOJA })  ◄──────────── │ ─── dispatch from chips    │
│  send({ REMOVER_CATEGORIA })  ◄──────  │                           │
└──────────────────────────────────────── │ ──────────────────────────┘
                                          │
                                          ▼
┌─────────────────────────────────────────────────────────────────────┐
│  Omnibox.svelte (pure renderer — zero $state)                       │
│                                                                     │
│  ┌─ flex-wrap container (role="group", aria-label="Filtros ativos")─┐
│  │  🔍  [Loja_Chip] [Loja_Chip] [Cat_Chip]  |  <input ...>        │
│  └──────────────────────────────────────────────────────────────────┘
│                                                                     │
│  Derives:                                                           │
│    let lojas = $derived(engine.lojaCards)                            │
│    let categorias = $derived(engine.categoriaCards)                  │
│                                                                     │
│  Events dispatched:                                                  │
│    click X on Loja_Chip → engine.send({ type: 'REMOVER_LOJA' })     │
│    click X on Cat_Chip  → engine.send({ type: 'REMOVER_CATEGORIA' })│
└─────────────────────────────────────────────────────────────────────┘
```

**Key constraint:** No new engine state. The existing `engine.lojaCards` and `engine.categoriaCards` getters already expose all data needed. Omnibox reads them via `$derived` and renders chips. Events `REMOVER_LOJA` / `REMOVER_CATEGORIA` already exist.

---

## Component Structure

### Omnibox.svelte — Updated Layout

The current `<input>` with a `relative` div becomes a **flex-wrap container** that holds chips + input:

```svelte
<div
  class="flex flex-wrap items-center gap-1.5 rounded-sm border border-input
         bg-background px-2 py-1.5 focus-within:border-ring
         focus-within:ring-2 focus-within:ring-ring/20"
  role="group"
  aria-label="Filtros ativos"
>
  <!-- Leading search icon -->
  <span class="pointer-events-none shrink-0 opacity-50">🔍</span>

  <!-- Loja chips -->
  {#each lojas as l (l.id)}
    <Badge variant="loja" aria-label="Loja: {l.nome} — ativa">
      🏪 {l.nome}
      <button
        type="button"
        aria-label="Remover loja {l.nome}"
        onclick={() => engine.send({ type: 'REMOVER_LOJA', shopId: l.id })}
        class="..."
      >✕</button>
    </Badge>
  {/each}

  <!-- Categoria chips -->
  {#each categorias as c (c.nome)}
    <Badge variant="categoria" aria-label="Categoria: {c.nome} — ativa">
      🏷️ {c.nome}
      <button
        type="button"
        aria-label="Remover categoria {c.nome}"
        onclick={() => engine.send({ type: 'REMOVER_CATEGORIA', nome: c.nome })}
        class="..."
      >✕</button>
    </Badge>
  {/each}

  <!-- Text input (grows to fill remaining space) -->
  <input
    class="min-w-[120px] flex-1 border-none bg-transparent outline-none ..."
    ...combobox attrs...
  />
</div>
```

The `<input>` shrinks as chips fill the line, and the flex container wraps to multiple lines when needed (Req 5).

### What stays the same

- Dropdown listbox (`<ul role="listbox">`) remains unchanged below the container.
- All keyboard/combobox logic remains driven by `engine.omnibox`.
- `aria-live` region for option count stays.
- A second `aria-live` region is added for chip removal announcements.

---

## New Badge Variants

Two new entries in `Badge.svelte`'s VARIANTS map:

| Variant      | Background             | Text               | Border             | Icon |
|-------------|------------------------|--------------------|--------------------|------|
| `loja`      | `bg-[var(--ouro-fundo)]` | `text-[var(--ouro-escuro)]` | `border-[var(--ouro-claro)]` | 🏪   |
| `categoria` | `bg-[var(--rosa-fundo)]` | `text-[var(--rosa)]`        | `border-border`     | 🏷️   |

### CSS Token Mapping

```
Variant loja:
  --ouro-fundo   → background (warm gold wash)
  --ouro-escuro  → text color (dark gold / readable)
  --ouro-claro   → border accent

Variant categoria:
  --rosa-fundo   → background (soft rose/terracotta wash)
  --rosa         → text color (mid-tone rosa, ≥4.5:1 contrast)
  border-border  → neutral border (existing token)
```

The hue difference between `--ouro-fundo` (~40° warm gold) and `--rosa-fundo` (~350° rose/terracotta) is ~50° on the HSL wheel, exceeding the 30° minimum (Req 8.2).

Dark mode: tokens are defined in `app.css` as CSS custom properties with dark-mode overrides via `prefers-color-scheme` or `.dark` class. No additional work needed — chips inherit the correct values automatically.

---

## ESLint Rule Configuration

New block added to `web/eslint.config.js`:

```js
{
  files: [
    '**/Omnibox.svelte',
    '**/StoreCard.svelte',
    '**/BuscaUnificada.svelte'
  ],
  rules: {
    'no-restricted-syntax': ['error', {
      selector: 'CallExpression[callee.name="$state"]',
      message: '❌ $state proibido em pure-renderer (ADR-0033). Use engine.send() para mutar estado.'
    }]
  }
}
```

This fires at lint time (`bun run lint:js`) and CI. The rule is scoped only to pure-renderer components of the Descobrir page (Req 9.4).

### Mise task: `check:pure-renderers`

A shell script at `.mise/tasks/check/pure-renderers` that greps for `$state(` in the target files and exits non-zero if found. Provides instant feedback without needing the full ESLint pass.

---

## What Gets Removed from BuscaUnificada.svelte

The "Escopo ativo" section (currently lines ~128-145):

```svelte
<!-- REMOVE THIS BLOCK -->
{#if engine.lojaCards.length || engine.categoriaCards.length}
  <div class="flex flex-wrap gap-2.5">
    {#each engine.lojaCards as l (l.id)}
      <LojaCard ... onremover={() => engine.send({ type: 'REMOVER_LOJA', shopId: l.id })} />
    {/each}
    {#each engine.categoriaCards as c (c.nome)}
      <CategoriaCard ... onremover={() => engine.send({ type: 'REMOVER_CATEGORIA', nome: c.nome })} />
    {/each}
  </div>
{/if}
```

The `LojaCard` and `CategoriaCard` **imports stay** (they may be reused elsewhere), but they are no longer rendered in BuscaUnificada. The import can be removed if the linter warns about unused imports, or left with a comment noting reuse potential (Req 4.3).

---

## Accessibility Approach

| Concern | Implementation |
|---------|---------------|
| Chip container grouping | `role="group"` + `aria-label="Filtros ativos"` on the flex wrapper |
| Individual chip identity | Each `<Badge>` gets `aria-label="Loja: {nome} — ativa"` or `"Categoria: {nome} — ativa"` |
| Remove button | `aria-label="Remover loja {nome}"` / `"Remover categoria {nome}"` on the ✕ button |
| Removal announcement | `<span aria-live="polite">` updated with `"Loja {nome} removida do escopo"` on removal |
| Combobox semantics | `role="combobox"`, `aria-expanded`, `aria-controls`, `aria-activedescendant` stay on the `<input>` |
| Focus management | Focus returns to `<input>` after chip removal |

The existing `aria-live` for option count is separate from the new chip-removal announcer — two distinct live regions to avoid message collisions.

---

## State Flow Summary

```
User adds loja via Omnibox → engine.send(ADICIONAR_LOJA)
  → engine updates ctx.shopIds
    → engine.lojaCards getter recomputes
      → Omnibox $derived(engine.lojaCards) reactively updates
        → new Loja_Chip rendered

User clicks ✕ on Loja_Chip → engine.send(REMOVER_LOJA, shopId)
  → engine updates ctx.shopIds
    → engine.lojaCards getter recomputes
      → Omnibox $derived reactively updates
        → chip disappears
          → aria-live announces removal
```

No new `$state` in Omnibox. No intermediate stores. Pure derived rendering.
