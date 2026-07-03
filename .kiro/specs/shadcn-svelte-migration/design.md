# Design Document: shadcn-svelte Migration

## Overview

This design describes how to migrate the Garimpei frontend from hand-written CSS components to shadcn-svelte, an incremental upgrade since the project already uses Bits UI (which shadcn-svelte builds on top of). The migration installs Tailwind CSS v4, initializes shadcn-svelte, maps existing design tokens to the shadcn-svelte theme, and replaces components one-by-one while keeping CI green.

## Architecture

### Current State

```
app.css → @import tokens.css (CSS variables)
         → global utility classes (.btn, .card, .badge, .input, .msg-erro, etc.)

components/ui/ → 5 Primitives (Button, Alert, Badge, Card, Input) with <style> blocks using tokens
              → 6 Composites (Select, Tabs, Dialog, Tooltip, DropdownMenu, ThemeToggle) wrapping Bits UI with :global() CSS
              → 9 Application components (DashPanel, MetricCard, etc.)
```

### Target State

```
app.css → @import "tailwindcss"
       → @theme { ... token mappings ... }
       → base styles (body, headings, links)
       → remaining utility classes not covered by Tailwind

components/ui/ → shadcn-svelte components using Tailwind classes + cn() utility
              → Same public APIs (variant/size props), Bits UI internals preserved
              → 9 Application components (unchanged or minimally updated)
```

### Key Decisions

1. **Tailwind CSS v4** — Uses the new CSS-first configuration (`@theme` directive in CSS) rather than `tailwind.config.js`. This aligns with Svelte 5 ecosystem tooling and simplifies configuration.

2. **Token preservation strategy** — Design tokens remain as CSS custom properties but are mapped into the `@theme` block so Tailwind generates corresponding utility classes. The shadcn-svelte CSS variables (`--primary`, `--background`, etc.) reference the existing token values.

3. **Dark mode via CSS selector** — shadcn-svelte's dark mode will use the existing `[data-theme="dark"]` selector (configurable in Tailwind v4 via `@variant dark (&:where([data-theme="dark"], [data-theme="dark"] *))`) rather than the default `.dark` class.

4. **Component output path** — shadcn-svelte components go into `$lib/components/ui/` following the project's existing convention. The `components.json` will be configured accordingly.

5. **Incremental migration order** — Foundation first (Tailwind + theme), then primitives (Button → Alert → Badge → Card → Input), then composites (Select → Tabs → Dialog → DropdownMenu → Tooltip), then cleanup.

## Detailed Design

### Phase 1: Foundation (Tailwind + shadcn-svelte init)

#### 1.1 Install Tailwind CSS v4

```bash
npm install -D tailwindcss @tailwindcss/vite
```

Update `vite.config.js`:
```js
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()]
});
```

Update `app.css` to import Tailwind at the top:
```css
@import "tailwindcss";
```

#### 1.2 Configure the Theme

In `app.css`, after the Tailwind import, define the theme mapping:

```css
@theme {
  /* Colors — mapped from tokens */
  --color-porcelana: var(--porcelana);
  --color-nevoa: var(--nevoa);
  --color-branco: var(--branco);
  --color-linha: var(--linha);
  --color-tinta-suave: var(--tinta-suave);
  --color-tinta: var(--tinta);
  --color-ouro: var(--ouro);
  --color-ouro-hover: var(--ouro-hover);
  --color-ouro-claro: var(--ouro-claro);
  --color-ouro-fundo: var(--ouro-fundo);
  --color-ouro-escuro: var(--ouro-escuro);
  --color-rosa: var(--rosa);
  --color-rosa-hover: var(--rosa-hover);
  --color-rosa-fundo: var(--rosa-fundo);
  --color-sucesso-texto: var(--sucesso-texto);
  --color-sucesso-fundo: var(--sucesso-fundo);
  --color-sucesso-borda: var(--sucesso-borda);
  --color-erro-texto: var(--erro-texto);
  --color-erro-fundo: var(--erro-fundo);
  --color-erro-borda: var(--erro-borda);
  --color-aviso-texto: var(--aviso-texto);
  --color-aviso-fundo: var(--aviso-fundo);
  --color-aviso-borda: var(--aviso-borda);

  /* shadcn-svelte semantic tokens */
  --color-background: var(--porcelana);
  --color-foreground: var(--tinta);
  --color-card: var(--nevoa);
  --color-card-foreground: var(--tinta);
  --color-primary: var(--ouro);
  --color-primary-foreground: var(--branco);
  --color-secondary: var(--porcelana);
  --color-secondary-foreground: var(--tinta);
  --color-muted: var(--porcelana);
  --color-muted-foreground: var(--tinta-suave);
  --color-accent: var(--ouro-fundo);
  --color-accent-foreground: var(--ouro-escuro);
  --color-destructive: var(--rosa);
  --color-destructive-foreground: var(--branco);
  --color-border: var(--linha);
  --color-input: var(--linha);
  --color-ring: var(--ouro);

  /* Font families */
  --font-display: 'Fraunces', Georgia, serif;
  --font-sans: 'Archivo', system-ui, sans-serif;
  --font-mono: 'Space Mono', ui-monospace, monospace;

  /* Border radius */
  --radius-sm: var(--raio-sm);
  --radius-md: var(--raio);
  --radius-lg: var(--raio-lg);
  --radius-full: var(--raio-full);
}
```

#### 1.3 Dark Mode Variant

Configure dark mode to use the existing `data-theme` attribute:

```css
@variant dark (&:where([data-theme="dark"], [data-theme="dark"] *));
```

#### 1.4 Initialize shadcn-svelte

```bash
npx shadcn-svelte@latest init
```

Configuration choices:
- Style: Default
- Base color: Neutral (we override with our tokens)
- Components path: `$lib/components/ui`
- Utils path: `$lib/utils.ts`

This creates `components.json` and the `cn()` utility function.

#### 1.5 Update CI Tooling

**Stylelint** — Add Tailwind at-rule exceptions:
```json
{
  "rules": {
    "at-rule-no-unknown": [true, {
      "ignoreAtRules": ["tailwind", "apply", "layer", "theme", "variant", "utility", "source"]
    }]
  }
}
```

Or install `stylelint-config-tailwindcss` for comprehensive support.

**Knip** — Add `components.json` to recognized config and `$lib/utils.ts` to entry points:
```json
{
  "entry": ["src/routes/**/+*.{js,ts,svelte}", "src/lib/**/*.{js,ts,svelte}"]
}
```

### Phase 2: Primitive Component Migration

Each primitive is replaced with a shadcn-svelte equivalent. The migration pattern:

1. Add the shadcn-svelte component via CLI: `npx shadcn-svelte@latest add button`
2. Customize the generated component to match existing prop API (variant names, sizes)
3. Update `index.js` barrel export (if path changes)
4. Run tests: `npm run check && npm run test:unit && npm run lint:css`
5. Remove old `<style>` block

#### 2.1 Button

Add via CLI, then customize variants to match existing API:
- `primary` → maps to shadcn `default` variant with `bg-ouro text-branco hover:bg-ouro-hover`
- `secondary` → maps to shadcn `outline` variant with brand colors
- `danger` → maps to shadcn `destructive` variant
- `ghost` → maps to shadcn `ghost` variant

Sizes: `sm`, `md` (default), `lg` mapped to shadcn size variants.

#### 2.2 Alert

Add via CLI: `npx shadcn-svelte@latest add alert`
- Map `error` → `destructive` variant
- Map `success` → custom variant with success tokens
- Map `warning` → custom variant with warning tokens
- Preserve `inline` prop as a size/layout modifier

#### 2.3 Badge

Add via CLI: `npx shadcn-svelte@latest add badge`
- Preserve existing variant names
- Map to brand colors via Tailwind classes

#### 2.4 Card

Add via CLI: `npx shadcn-svelte@latest add card`
- Uses Card, CardHeader, CardContent, CardFooter sub-components
- The simple `<Card>` usage maps to just `<Card><CardContent>...</CardContent></Card>`
- Surface styling comes from theme tokens automatically

#### 2.5 Input

Add via CLI: `npx shadcn-svelte@latest add input`
- Preserve `label`, `placeholder`, `disabled`, error state props
- Add wrapper component that adds label + error message pattern on top of shadcn Input primitive

### Phase 3: Composite Component Migration

These already use Bits UI internally. The migration replaces hand-written `:global()` CSS with Tailwind classes from shadcn-svelte patterns.

#### 3.1 Select

Add via CLI: `npx shadcn-svelte@latest add select`
- Preserve prop interface: `value` (bindable), `label`, `options`, `placeholder`, `size`, `disabled`
- The wrapper remains a single-file component that composes shadcn-svelte Select parts
- Remove all `:global(.select-*)` CSS rules

#### 3.2 Tabs

Add via CLI: `npx shadcn-svelte@latest add tabs`
- Preserve keyboard navigation (comes free from Bits UI)
- Style with Tailwind via shadcn-svelte classes

#### 3.3 Dialog

Add via CLI: `npx shadcn-svelte@latest add dialog`
- Preserve focus trap, escape-to-close
- Preserve `Tooltip.Provider` at layout level

#### 3.4 DropdownMenu

Add via CLI: `npx shadcn-svelte@latest add dropdown-menu`
- Preserve ARIA menu roles
- Keyboard navigation from Bits UI

#### 3.5 Tooltip

Add via CLI: `npx shadcn-svelte@latest add tooltip`
- `Tooltip.Provider` already in `+layout.svelte`
- Keyboard/focus activation preserved from Bits UI

### Phase 4: Cleanup and Documentation

#### 4.1 CSS Cleanup

- Remove hand-written utility classes from `app.css` that are now covered by Tailwind (`.flex`, `.gap-*`, `.text-*`, `.font-*`, `.truncate`, `.line-clamp-2`)
- Evaluate if `tokens.css` can be reduced — the `:root` block with color definitions stays (needed for the `@theme` references), but documentation/comments can be simplified
- Remove `.btn`, `.input`, `.card`, `.badge` global classes from `app.css` once all consumers use components
- Run `knip` to detect any dead exports

#### 4.2 Application Components Update

Application components (DashPanel, MetricCard, PageHeader, etc.) that use `<style>` blocks with token variables can remain as-is initially. They can be incrementally migrated to Tailwind classes in a future pass, but this is not required for the initial migration.

#### 4.3 Documentation Update

- Update `docs/componentes.md` with new component usage patterns
- Document how to add new shadcn-svelte components
- Document theme customization workflow

## File Changes

| File | Change |
|------|--------|
| `package.json` | Add `tailwindcss`, `@tailwindcss/vite`, `tailwind-merge`, `clsx`, `tailwind-variants` (or `class-variance-authority`) |
| `vite.config.js` | Add Tailwind CSS vite plugin |
| `app.css` | Add `@import "tailwindcss"`, `@theme` block, `@variant dark`, remove replaced utility classes |
| `components.json` | New — shadcn-svelte config |
| `$lib/utils.ts` | New — `cn()` utility function |
| `.stylelintrc.json` | Allow Tailwind at-rules |
| `knip.json` | Add `.ts` to entry globs, recognize `utils.ts` |
| `components/ui/Button.svelte` | Replace with shadcn-svelte Button |
| `components/ui/Alert.svelte` | Replace with shadcn-svelte Alert |
| `components/ui/Badge.svelte` | Replace with shadcn-svelte Badge |
| `components/ui/Card.svelte` | Replace with shadcn-svelte Card |
| `components/ui/Input.svelte` | Replace with shadcn-svelte Input |
| `components/ui/Select.svelte` | Restyle with shadcn-svelte patterns |
| `components/ui/Tabs.svelte` | Restyle with shadcn-svelte patterns |
| `components/ui/Dialog.svelte` | Restyle with shadcn-svelte patterns |
| `components/ui/DropdownMenu.svelte` | Restyle with shadcn-svelte patterns |
| `components/ui/Tooltip.svelte` | Restyle with shadcn-svelte patterns |
| `components/ui/index.js` | Update exports if paths change (shadcn uses subdirs) |
| `docs/componentes.md` | Update with new patterns |

## Testing Strategy

### Per-Component Verification

After each component migration:
1. `npm run check` — svelte-check passes
2. `npm run lint:css` — stylelint passes (no hex colors, Tailwind directives allowed)
3. `npm run lint:js` — ESLint passes (zero warnings)
4. `npm run lint:dead` — knip passes (no dead code)
5. `npm run test:unit` — all 141 unit tests pass
6. `npm run build` — static build succeeds

### Final Verification

After all migrations complete:
1. Full CI pipeline passes (all checks above + E2E)
2. `npm run test` — all 53 Playwright E2E tests pass
3. Visual inspection of all pages in both light and dark mode
4. axe-core accessibility audit reports zero violations
5. `npm run audit:ui` — coverage report shows improvement

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Tailwind class conflicts with existing utility classes in `app.css` | Migrate one utility class group at a time; Tailwind's `@layer` prevents specificity issues |
| shadcn-svelte Bits UI version mismatch | Lock Bits UI version; shadcn-svelte v2 targets Bits UI v2 |
| `color-no-hex` stylelint rule breaks on Tailwind-generated CSS | Stylelint only runs on source files (`.svelte`, `.css`), not generated output |
| Component API differences break consumers | Maintain same prop interface; add adapter layer if needed |
| Bundle size increase from Tailwind | Tailwind v4 tree-shakes aggressively; monitor with `npm run build` output |
| Dark mode regression | Test both themes after each migration; E2E tests cover critical flows |

## Correctness Properties

### CP-1: Token Mapping Preservation
For all design tokens defined in `tokens.css`, the corresponding Tailwind utility class SHALL produce the same computed CSS value as using the CSS variable directly.

### CP-2: Component API Backward Compatibility
For all components migrated from custom CSS to shadcn-svelte, the public prop interface (prop names, types, default values) SHALL remain identical so that consuming components require zero template changes.

### CP-3: Accessibility Invariant
For all migrated components, axe-core audits SHALL report zero WCAG AA violations, preserving the accessibility guarantees provided by Bits UI.

### CP-4: CI Pipeline Green Invariant
After each individual component migration commit, the full CI pipeline (svelte-check, stylelint, ESLint, knip, vitest, build) SHALL pass with zero errors.

### CP-5: Visual Parity in Dark Mode
For all migrated components, both `data-theme="light"` (default) and `data-theme="dark"` states SHALL render with correct contrast ratios (≥4.5:1 for text, ≥3:1 for UI elements) using the mapped token values.
