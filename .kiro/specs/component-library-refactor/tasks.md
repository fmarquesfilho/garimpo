# Implementation Plan

## Overview

Progressive refactor of the Garimpo frontend to adopt Bits UI as a headless component library with formalized design tokens. Organized in 5 phases: Foundation (tokens + deps + ADR), Primitives (Button, Input, Badge, Alert, Card), Composites (Select, Tabs, Dialog, Tooltip, DropdownMenu), Testing Infrastructure, and Bundle Monitoring + Migration Cleanup.

## Tasks

- [x] 1. Create `web/src/lib/components/ui/tokens.css` with all design tokens extracted from `app.css`, organized by category (color neutrals, ouro accents, rosa accents, feedback, spacing, typography families/scale/weights, surfaces). Preserve exact property names and values. Add category comment headers.
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.6
  - Design: Token Schema Structure

- [x] 2. Install Bits UI as a dependency in `web/` — run `npm install bits-ui` and verify it resolves a version compatible with Svelte 5 (^5.56.4). Confirm the package is added to `package.json` dependencies.
  - Requirements: 2.1, 2.2
  - Design: Architecture > Headless Primitive Layer

- [x] 3. Install test dependencies — `npm install -D vitest-axe axe-core fast-check` for accessibility testing and property-based testing support.
  - Requirements: 8.2, 8.3
  - Design: Testing Strategy > Property-Based Tests

- [x] 4. Create the ADR at `docs/adr/ADR-0001-component-library.md` with sections: Title, Status (Accepted), Context (current state of ad-hoc components, need for accessibility and consistency), Decision (Bits UI as headless layer), Evaluation Criteria (WCAG, headless/unstyled, Svelte 5 runes, active maintenance, bundle size), Rejected Alternatives (shadcn-svelte requires Tailwind, Skeleton UI is opinionated, Melt UI is lower-level), Token Naming Convention (preserve existing names, category grouping), Component File Structure (`$lib/components/ui/{Name}.svelte`, barrel export, scoped styles), and Consequences.
  - Requirements: 6.1, 6.2, 6.3, 6.4, 6.5
  - Design: Key Design Decisions

- [x] 5. Update `web/src/app.css` to import `tokens.css` via `@import` at the top of the file. Remove the duplicate `:root` token definitions from `app.css` (keep only the reset, base styles, and utility classes). Verify the app still renders correctly by running `npm run build` in `web/`.
  - Requirements: 1.6, 5.2
  - Design: Token Schema Structure

- [x] 6. Rewrite `Button.svelte` using the design spec: $props() with variant ('primary'|'secondary'|'danger'|'ghost'), size ('sm'|'md'|'lg'), disabled, type, onclick, children, ...rest. Add $derived() fallback for invalid variant/size. Use tokens only (no hardcoded values). Forward rest props to root element. Apply focus-visible and prefers-reduced-motion styles. Ensure hover transitions use ease timing within 150ms.
  - Requirements: 3.1, 3.2, 3.3, 3.6, 3.7, 3.8, 4.4, 4.5, 4.7, 4.8
  - Design: Component Interfaces > Button, Variant/Size Fallback Strategy

- [x] 7. Rewrite `Input.svelte` using the design spec: $props() with value ($bindable), type, placeholder, label, variant ('default'|'mono'), size ('sm'|'md'|'lg'), disabled, ...rest. Style with tokens only. Focus state: border-color --ouro, box-shadow 2px --ouro-fundo. Include optional label rendering.
  - Requirements: 3.1, 3.2, 3.3, 3.6, 3.8, 4.3, 4.7
  - Design: Component Interfaces > Input

- [x] 8. Rewrite `Badge.svelte` using the design spec: $props() with variant ('default'|'gold'|'pink'|'green'|'red'), children, ...rest. Style each variant with the corresponding token colors. Use --raio-full for pill shape.
  - Requirements: 3.1, 3.2, 3.6, 3.8, 4.1
  - Design: Component Interfaces > Badge

- [x] 9. Rewrite `Alert.svelte` using the design spec: $props() with variant ('error'|'success'|'warning'), inline, children, ...rest. Map variants to feedback tokens (--erro-*, --sucesso-*, --aviso-*). Support inline mode with reduced padding.
  - Requirements: 3.1, 3.2, 3.6, 3.8, 4.1
  - Design: Component Interfaces > Alert

- [x] 10. Rewrite `Card.svelte` using the design spec: $props() with variant ('default'|'highlight'|'success'|'error'), padding ('sm'|'md'|'lg'), children, ...rest. Use --raio for border-radius, --nevoa for background, --linha for border. Apply variant-specific styles (highlight uses --ouro-claro border + gradient, success/error use left-border accent).
  - Requirements: 3.1, 3.2, 3.6, 3.8, 4.1, 4.3
  - Design: Component Interfaces > Card

- [x] 11. Update `web/src/lib/components/ui/index.js` barrel export to include all Phase 2 components. Verify build passes with `npm run build` in `web/`.
  - Requirements: 3.4
  - Design: File Structure

- [x] 12. Implement `Select.svelte` wrapping Bits UI Select: import Select from 'bits-ui', compose Root > Trigger > Content > Item. $props() with value ($bindable), label, options, placeholder, size, disabled, ...rest. Style via scoped CSS targeting Bits UI data-* attributes with tokens. Preserve all ARIA attributes.
  - Requirements: 2.1, 2.3, 2.4, 3.1, 3.5, 3.6, 3.8
  - Design: Component Interfaces > Select, Bits UI Integration Pattern

- [x] 13. Implement `Tabs.svelte` wrapping Bits UI Tabs: compose Root > List > Trigger + Content panels. $props() with tabs array, active ($bindable), children, ...rest. Support keyboard arrow navigation between tabs. Style tab triggers and active state with tokens.
  - Requirements: 2.1, 2.3, 2.4, 3.1, 3.5, 3.6, 3.8
  - Design: Component Interfaces > Tabs

- [x] 14. Implement `Dialog.svelte` wrapping Bits UI Dialog: compose Root > Portal > Overlay > Content > Title > Description > Close. $props() with open ($bindable), title, description, children, ...rest. Style overlay with semi-transparent background, content with --nevoa/--raio-lg/--sombra. Ensure focus trap and Escape-to-close.
  - Requirements: 2.1, 2.3, 2.4, 3.1, 3.5, 3.6, 3.8
  - Design: Component Interfaces > Dialog

- [x] 15. Implement `Tooltip.svelte` wrapping Bits UI Tooltip: compose Root > Trigger > Content. $props() with content, side ('top'|'bottom'|'left'|'right'), children, ...rest. Style content with --tinta background, white text, small text, --raio-sm radius.
  - Requirements: 2.1, 2.3, 2.4, 3.1, 3.6, 3.8
  - Design: Component Interfaces > Tooltip

- [x] 16. Implement `DropdownMenu.svelte` wrapping Bits UI DropdownMenu: compose Root > Trigger > Content > Item. $props() with items array, children (trigger content), ...rest. Style items with hover state using --porcelana. Support destructive items with --erro-texto color. Keyboard arrow navigation.
  - Requirements: 2.1, 2.3, 2.4, 3.1, 3.5, 3.6, 3.8
  - Design: Component Interfaces > DropdownMenu

- [x] 17. Update barrel export `index.js` with new composite components (Select, Tabs, Dialog, Tooltip, DropdownMenu). Run `npm run build` to verify tree-shaking works — unused composites should not appear in route chunks.
  - Requirements: 3.4, 7.1, 7.2
  - Design: File Structure

- [ ] 18. Create unit test file `web/src/lib/components/ui/__tests__/Button.test.js` with tests: renders each variant correctly, renders each size, falls back to defaults for invalid values, forwards rest props to root element, fires onclick, respects disabled state, renders children.
  - Requirements: 8.1, 8.6
  - Design: Testing Strategy > Unit Tests

- [ ] 19. Create unit test files for Input, Badge, Alert, Card (`__tests__/{Component}.test.js`): test variant rendering, size rendering, rest props forwarding, default fallbacks. For Input: test $bindable value, label rendering. For Card: test padding sizes.
  - Requirements: 8.1, 8.6
  - Design: Testing Strategy > Unit Tests

- [ ] 20. Create unit test files for Select, Tabs, Dialog, Tooltip, DropdownMenu: test rendering, open/close state via $bindable, keyboard navigation (arrow keys for Select/Tabs/DropdownMenu, Escape for Dialog/Tooltip), ARIA attribute presence.
  - Requirements: 8.1, 8.5, 8.6
  - Design: Testing Strategy > Unit Tests, Accessibility Tests

- [ ] 21. Create accessibility test file `web/src/lib/components/ui/__tests__/accessibility.test.js`: for each interactive component (Button, Select, Dialog, Tabs, Tooltip, DropdownMenu), render with valid props and run axe-core, assert zero violations with WCAG 2.1 Level AA ruleset.
  - Requirements: 8.2, 8.3
  - Design: Testing Strategy > Accessibility Tests

- [ ] 22. Create property-based test file `web/src/lib/components/ui/__tests__/properties.test.js` with fast-check: Property 2 (variant/size fallback with fc.string()), Property 3 (rest props forwarding with fc.record()), Property 5 (accessibility compliance with random valid prop combos). Minimum 100 iterations per property.
  - Requirements: 8.1, 8.3
  - Design: Testing Strategy > Property-Based Tests, Correctness Properties

- [ ] 23. Create Playwright e2e test `web/tests/component-library.spec.js`: test navigation between pages, Dialog open/close, Select option picking, Tab switching, form submission with Input + Button.
  - Requirements: 8.4
  - Design: Testing Strategy > E2E Tests

- [ ] 24. Update CI configuration to enforce 80% branch coverage threshold for component unit tests. Add vitest coverage configuration with `--coverage.branches 80` and ensure pipeline fails on threshold violation.
  - Requirements: 8.5, 8.6
  - Design: Testing Strategy > CI Integration

- [ ] 25. Create a bundle size baseline: run `npm run build` and record the gzipped sizes of each route chunk in a `web/bundle-baseline.json` file. Create a post-build script (`web/scripts/check-bundle-size.js`) that compares current build output against baseline and fails if any route exceeds 15KB gzipped delta.
  - Requirements: 7.3, 7.4, 7.5
  - Design: Error Handling > Bundle Size Budget

- [ ] 26. Add the bundle size check to the build pipeline — update `package.json` scripts to run `check-bundle-size.js` after `vite build`. Verify it passes with current implementation.
  - Requirements: 7.5
  - Design: Error Handling > Bundle Size Budget

- [ ] 27. Remove legacy duplicate components: delete any components in `web/src/lib/components/` that are now fully replaced by the ui/ versions (EmptyState is duplicated). Update imports in consuming routes/pages to use the barrel export from `$lib/components/ui`. Remove corresponding utility classes from `app.css` that are now handled by component-scoped CSS (e.g., .btn, .card, .badge global classes can be deprecated).
  - Requirements: 5.3, 5.4
  - Design: Migration Adapter Pattern

- [ ] 28. Final verification: run full test suite (`npm run test:unit`, `npm run test`), run build with bundle check, verify no lint errors (`npm run lint`). Confirm all 10 components are exported from barrel, all tests pass, bundle budget is met.
  - Requirements: 7.3, 8.5, 8.6
  - Design: CI Integration

## Task Dependency Graph

```json
{
  "waves": [
    {
      "name": "Phase 1: Foundation",
      "tasks": [1, 2, 3, 4],
      "description": "Tokens, dependencies, and ADR — all independent"
    },
    {
      "name": "Phase 1b: Token Integration",
      "tasks": [5],
      "dependencies": [1],
      "description": "Refactor app.css to import tokens.css"
    },
    {
      "name": "Phase 2: Primitives",
      "tasks": [6, 7, 8, 9, 10],
      "dependencies": [1, 2, 5],
      "description": "Rewrite primitive components using tokens + Svelte 5 patterns"
    },
    {
      "name": "Phase 2b: Barrel Export",
      "tasks": [11],
      "dependencies": [6, 7, 8, 9, 10],
      "description": "Update index.js and verify build"
    },
    {
      "name": "Phase 3: Composites",
      "tasks": [12, 13, 14, 15, 16],
      "dependencies": [2, 5, 11],
      "description": "Implement Bits UI composite wrappers"
    },
    {
      "name": "Phase 3b: Barrel Export",
      "tasks": [17],
      "dependencies": [12, 13, 14, 15, 16],
      "description": "Update barrel export and verify tree-shaking"
    },
    {
      "name": "Phase 4: Testing",
      "tasks": [18, 19, 20, 21, 22],
      "dependencies": [3, 17],
      "description": "Unit, accessibility, and property-based tests"
    },
    {
      "name": "Phase 4b: E2E + CI",
      "tasks": [23, 24],
      "dependencies": [18, 19, 20, 21, 22],
      "description": "Playwright e2e and coverage enforcement"
    },
    {
      "name": "Phase 5: Bundle + Cleanup",
      "tasks": [25, 26],
      "dependencies": [17],
      "description": "Bundle size monitoring infrastructure"
    },
    {
      "name": "Phase 5b: Migration Cleanup",
      "tasks": [27],
      "dependencies": [23, 25, 26],
      "description": "Remove legacy duplicates and deprecated utilities"
    },
    {
      "name": "Phase 5c: Final Verification",
      "tasks": [28],
      "dependencies": [24, 27],
      "description": "Full test suite, build, lint — green across the board"
    }
  ]
}
```

## Notes

- Tasks 1-5 can be done in parallel (no interdependencies within Phase 1)
- Tasks 6-10 can be done in parallel after Phase 1 is complete
- Tasks 12-16 can be done in parallel after task 11
- Tasks 18-22 can be done in parallel after task 17
- Each phase should be a separate PR for easier review
- Migration adapters (Req 5.1) are only needed if legacy prop interfaces change — current Button already uses the same interface, so direct replacement is sufficient
