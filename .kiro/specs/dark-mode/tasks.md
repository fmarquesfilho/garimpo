# Implementation Plan

## Overview

Implement dark mode for the Garimpo frontend using the existing design token system. Organized in 4 phases: Token Definitions, Theme Engine, Toggle UI, and Testing/Documentation.

## Tasks

- [x] 1. Add dark colour tokens to `web/src/lib/components/ui/tokens.css` — define `:root[data-theme="dark"]` block with all colour overrides (neutrals, ouro, rosa, feedback, shadow). Preserve spacing/typography/radii unchanged. Verify WCAG AA contrast ≥4.5:1 for all text/background pairs.
  - Requirements: 4.1, 4.2, 4.3, 5.1, 5.2, 5.3
  - Design: Data Models > Dark Token Palette

- [x] 2. Create `web/src/lib/theme.js` — theme engine module with functions: `getStoredTheme()`, `resolveTheme()`, `applyTheme(theme)`, `setTheme(preference)`, `onSystemChange(callback)`. Export a Svelte-compatible store (`theme`). Handle localStorage unavailable gracefully.
  - Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 2.4
  - Design: Components and Interfaces > Theme Engine

- [x] 3. Add blocking inline script to `web/src/app.html` `<head>` — synchronous IIFE that reads localStorage + matchMedia, sets `data-theme` attribute on `<html>` before body parses. Fallback to light on error. Must be under 10 LOC, no async, no imports.
  - Requirements: 3.1, 3.2, 3.3
  - Design: Components and Interfaces > Blocking Inline Script

- [x] 4. Add theme transition CSS to `web/src/app.css` — after first paint, apply `transition: background-color 200ms, color 200ms` to all elements. Disable during initial load (class `no-transitions`). Respect `prefers-reduced-motion: reduce` (0ms duration).
  - Requirements: 6.1, 6.2, 6.3
  - Design: Architecture > Theme Resolution Flow

- [x] 5. Create `web/src/lib/components/ui/ThemeToggle.svelte` — button that cycles through light→dark→system. Shows icon (☀️/🌙/🖥️), has `aria-label` per state. Imports theme store. Styled with tokens (works in both modes).
  - Requirements: 8.1, 8.2, 8.3
  - Design: Components and Interfaces > ThemeToggle Component

- [x] 6. Integrate ThemeToggle in `web/src/routes/+layout.svelte` — add to header bar, visible only when user is authenticated. Initialize theme engine on mount (subscribe to system changes). Remove `no-transitions` class after first paint.
  - Requirements: 8.1, 8.4, 1.3, 6.3
  - Design: File Structure

- [x] 7. Update `web/src/lib/components/ui/index.js` barrel export to include ThemeToggle.
  - Requirements: 8.1
  - Design: File Structure

- [x] 8. Verify Bits UI compatibility — ensure all composite components (Select, Tabs, Dialog, Tooltip, DropdownMenu) render correctly in dark mode. Check that `data-state`/`data-highlighted` styling uses only token references. Fix any hardcoded colours in `:global()` selectors.
  - Requirements: 7.1, 7.2, 7.3
  - Design: Correctness Properties > Property 3

- [x] 9. Update stylelint config — add `:root[data-theme="dark"]` to the override that allows hex colours (same as tokens.css/app.css exception).
  - Requirements: 4.1
  - Design: Data Models > Dark Token Palette

- [x] 10. Write unit tests for `theme.js` — cover all combinations of localStorage × matchMedia, invalid values, setTheme round-trip, onSystemChange callback.
  - Requirements: 1.1, 1.2, 2.2, 2.3, 2.4, 3.3
  - Design: Testing Strategy > Unit Tests

- [x] 11. Write contrast compliance test — parse tokens.css, extract light and dark colour pairs, compute WCAG contrast ratios, assert ≥4.5:1 for all text/background combinations.
  - Requirements: 5.1, 5.2, 5.3
  - Design: Testing Strategy > Contrast Tests

- [ ] 12. Write Playwright e2e test for dark mode — test system preference detection, toggle cycling, persistence across reload, no FOUC verification.
  - Requirements: 1.1, 2.2, 3.1
  - Design: Testing Strategy > E2E Tests

- [x] 13. Run full CI pipeline locally (`npm run ci:local`) and verify: build green, stylelint passes (no hex violations in dark tokens block), svelte-check clean, all tests pass.
  - Requirements: 7.1, 7.3
  - Design: Testing Strategy

- [x] 14. Create backlog task `T-0039-dark-mode.yaml` documenting the feature with OLED energy efficiency rationale, token architecture reference, and WCAG AA compliance commitment. Update `docs/componentes.md` with dark mode usage section.
  - Requirements: 9.1, 9.2
  - Design: N/A (documentation)

## Task Dependency Graph

```json
{
  "waves": [
    {
      "name": "Phase 1: Tokens + Engine",
      "tasks": [1, 2, 3],
      "description": "Dark palette, theme logic, FOUC prevention — independent"
    },
    {
      "name": "Phase 2: Integration",
      "tasks": [4, 5, 6, 7, 9],
      "dependencies": [1, 2, 3],
      "description": "CSS transitions, toggle component, layout integration, stylelint"
    },
    {
      "name": "Phase 3: Compatibility + Testing",
      "tasks": [8, 10, 11, 12],
      "dependencies": [4, 5, 6],
      "description": "Bits UI verification, unit tests, contrast tests, e2e"
    },
    {
      "name": "Phase 4: Verification + Docs",
      "tasks": [13, 14],
      "dependencies": [8, 10, 11, 12],
      "description": "Full CI validation and documentation"
    }
  ]
}
```

## Notes

- Tasks 1, 2, 3 can be done in parallel (no interdependencies)
- Tasks 4, 5, 9 can be done in parallel after Phase 1
- The dark palette values in task 1 should be verified with a contrast checker tool before implementation
- The blocking script (task 3) must be duplicated logic from theme.js — keep both in sync
- Dark mode does NOT change: spacing, typography, radii, font weights, layout
