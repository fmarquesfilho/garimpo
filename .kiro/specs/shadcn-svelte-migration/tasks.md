# Tasks

## Phase 1: Foundation — Tailwind CSS + shadcn-svelte Setup

- [ ] 1.1 Install Tailwind CSS v4 and `@tailwindcss/vite` plugin as dev dependencies
- [ ] 1.2 Update `vite.config.js` to add the Tailwind CSS Vite plugin before the SvelteKit plugin
- [ ] 1.3 Update `app.css` to add `@import "tailwindcss"` at the top (before token import)
- [ ] 1.4 Add `@theme` block in `app.css` mapping existing CSS variable tokens to Tailwind color/font/radius utilities
- [ ] 1.5 Configure dark mode variant to use `[data-theme="dark"]` selector instead of `.dark` class
- [ ] 1.6 Install `tailwind-merge` and `clsx` dependencies for the `cn()` utility
- [ ] 1.7 Run `npx shadcn-svelte@latest init` to create `components.json` and utils file
- [ ] 1.8 Update `.stylelintrc.json` to allow Tailwind at-rules (`@theme`, `@variant`, `@utility`, `@source`, `@import "tailwindcss"`)
- [ ] 1.9 Update `knip.json` to include `.ts` files in entry/project globs and recognize `$lib/utils.ts`
- [ ] 1.10 Verify foundation: run `npm run check`, `npm run lint:css`, `npm run lint:js`, `npm run lint:dead`, `npm run build`, `npm run test:unit` — all must pass

## Phase 2: Primitive Component Migration

- [ ] 2.1 Add shadcn-svelte Button component via CLI, customize variants to match existing API (primary, secondary, danger, ghost) and sizes (sm, md, lg)
- [ ] 2.2 Update all Button consumers if any import path changes; verify `npm run test:unit` passes
- [ ] 2.3 Add shadcn-svelte Alert component via CLI, customize variants (error → destructive, success, warning) preserving `inline` prop
- [ ] 2.4 Verify Alert migration: `npm run check && npm run test:unit && npm run lint:css`
- [ ] 2.5 Add shadcn-svelte Badge component via CLI, map existing variant props to Tailwind classes
- [ ] 2.6 Verify Badge migration: `npm run check && npm run test:unit && npm run lint:css`
- [ ] 2.7 Add shadcn-svelte Card component via CLI, adapt to existing simple `<Card>` usage pattern with children slot
- [ ] 2.8 Verify Card migration: `npm run check && npm run test:unit && npm run lint:css`
- [ ] 2.9 Add shadcn-svelte Input component via CLI, preserve label/placeholder/disabled/error props interface
- [ ] 2.10 Verify Input migration: `npm run check && npm run test:unit && npm run lint:css`

## Phase 3: Composite Component Migration

- [ ] 3.1 Add shadcn-svelte Select component via CLI, create wrapper preserving existing prop interface (value bindable, label, options, placeholder, size, disabled)
- [ ] 3.2 Remove all `:global(.select-*)` CSS from Select.svelte; verify tests pass
- [ ] 3.3 Add shadcn-svelte Tabs component via CLI, preserve keyboard navigation and existing consumer API
- [ ] 3.4 Verify Tabs migration: `npm run check && npm run test:unit`
- [ ] 3.5 Add shadcn-svelte Dialog component via CLI, preserve focus trap, escape-to-close, and overlay click behavior
- [ ] 3.6 Verify Dialog migration: `npm run check && npm run test:unit`
- [ ] 3.7 Add shadcn-svelte DropdownMenu component via CLI, preserve ARIA menu roles and keyboard nav
- [ ] 3.8 Verify DropdownMenu migration: `npm run check && npm run test:unit`
- [ ] 3.9 Add shadcn-svelte Tooltip component via CLI, verify Provider in layout still works
- [ ] 3.10 Verify Tooltip migration: `npm run check && npm run test:unit`

## Phase 4: Cleanup

- [ ] 4.1 Remove replaced global utility classes from `app.css` (`.btn`, `.input`, `.card`, `.badge`, `.flex`, `.gap-*`, `.text-*`, `.font-*`, `.truncate`, `.line-clamp-2`) that are now covered by Tailwind
- [ ] 4.2 Update `components/ui/index.js` barrel exports to match new file structure (shadcn-svelte uses subdirectories)
- [ ] 4.3 Run `npm run lint:dead` (knip) and remove any dead exports or unused imports
- [ ] 4.4 Evaluate `tokens.css` — keep `:root` color definitions (referenced by `@theme`), remove redundant comments if desired
- [ ] 4.5 Run full CI check: `npm run check && npm run lint:css && npm run lint:js && npm run lint:dead && npm run build && npm run test:unit`
- [ ] 4.6 Run Playwright E2E tests: `npm run test` — all 53 tests must pass

## Phase 5: Documentation

- [ ] 5.1 Update `docs/componentes.md` with new shadcn-svelte component usage patterns and examples
- [ ] 5.2 Document the theme configuration: how `tokens.css` maps through `@theme` to Tailwind utilities and shadcn-svelte semantic colors
- [ ] 5.3 Document how to add new shadcn-svelte components via CLI (`npx shadcn-svelte@latest add <component>`)
- [ ] 5.4 Create a migration reference mapping old component APIs to new equivalents (for developers still adapting)
