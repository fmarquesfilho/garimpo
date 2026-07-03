# Requirements Document

## Introduction

This feature formalizes the Garimpo frontend component architecture by adopting Bits UI as a headless component library and establishing a design token system. The goal is to replace ad-hoc utility classes and duplicated component styles with accessible, composable primitives styled via existing CSS custom properties — preserving the editorial gold/porcelain visual identity while gaining WCAG-compliant keyboard navigation, ARIA semantics, and consistent interaction patterns across all UI surfaces.

## Glossary

- **Component_Library**: The set of reusable Svelte 5 UI primitives (Button, Select, Dialog, Tabs, etc.) built on top of Bits UI headless components and styled with Design_Tokens
- **Bits_UI**: A headless, accessible component library built for Svelte 5 using runes; provides behavior and ARIA semantics without opinionated styling
- **Design_Token**: A named CSS custom property that encodes a visual decision (color, spacing, typography, radius, shadow) from the Garimpo design system
- **Token_Schema**: The formal catalog of all Design_Tokens organized by category (color, spacing, typography, surface) with their names, values, and usage rules
- **Headless_Component**: A component that provides interaction behavior and accessibility semantics without visual styling
- **Migration_Adapter**: A wrapper component that provides the same prop interface as an existing UI component while internally delegating to the new Bits_UI-based implementation
- **Visual_Identity**: The Garimpo editorial aesthetic defined by the 60-30-10 color rule (porcelana/tinta/ouro), Fraunces display type, warm gold/pink accents, and rounded surfaces

## Requirements

### Requirement 1: Design Token Schema Definition

**User Story:** As a frontend developer, I want a single source of truth for all design decisions as formal tokens, so that components use consistent values and design changes propagate automatically.

#### Acceptance Criteria

1. THE Token_Schema SHALL define color tokens as CSS custom properties with the following groups: neutrals (--porcelana, --nevoa, --branco, --linha, --tinta-suave, --tinta), ouro accents (--ouro, --ouro-hover, --ouro-claro, --ouro-fundo, --ouro-escuro), rosa accents (--rosa, --rosa-hover, --rosa-fundo), and feedback colors (--sucesso-texto, --sucesso-fundo, --sucesso-borda, --erro-texto, --erro-fundo, --erro-borda, --aviso-texto, --aviso-fundo, --aviso-borda), each preserving the same property name and value defined in app.css
2. THE Token_Schema SHALL define spacing tokens (--r1, --r2, --r3, --r4, --r5, --r6, --r8, --r12) with their corresponding rem values: 0.25rem, 0.5rem, 0.75rem, 1rem, 1.25rem, 1.5rem, 2rem, 3rem
3. THE Token_Schema SHALL define typography tokens for font families (--display, --ui, --mono), font sizes (--text-xs through --text-2xl: 0.7rem, 0.8rem, 0.9rem, 0.95rem, 1.1rem, 1.3rem, 1.5rem), and font weight tokens for semi-bold (600) and bold (700)
4. THE Token_Schema SHALL define surface tokens for border-radius (--raio: 14px, --raio-sm: 8px, --raio-lg: 16px, --raio-full: 999px) and shadow (--sombra)
5. THE Component_Library SHALL reference only Design_Tokens for color, spacing, typography, border-radius, and shadow values rather than hard-coded literals in all component style definitions
6. THE Token_Schema SHALL be implemented as a single CSS file that can be imported via standard CSS @import or framework-level import by any component in the project

### Requirement 2: Bits UI Integration

**User Story:** As a frontend developer, I want a headless component library that provides accessible behavior primitives, so that I can build UI components with correct keyboard navigation and ARIA semantics without reimplementing interaction patterns.

#### Acceptance Criteria

1. THE Component_Library SHALL use Bits UI as its headless primitive layer for interactive components (buttons, selects, dialogs, tabs, tooltips, dropdowns)
2. THE Component_Library SHALL require Bits UI version compatible with Svelte 5 runes ($props, $state, $derived)
3. WHEN a Bits UI primitive is wrapped, THE Component_Library SHALL apply styling exclusively through CSS custom properties from the Token_Schema and may use Bits UI data-state attributes as CSS selectors for state-dependent styles
4. THE Component_Library SHALL preserve all accessibility attributes (aria-*, role, tabindex, id) and data-state attributes rendered by Bits UI without modification or removal
5. IF Bits UI does not provide a primitive for a needed interaction pattern, THEN THE Component_Library SHALL implement the pattern directly following WAI-ARIA Authoring Practices

### Requirement 3: Core Component Implementation

**User Story:** As a frontend developer, I want a standardized set of primitive UI components, so that I can compose interfaces consistently without duplicating interaction logic or styling.

#### Acceptance Criteria

1. THE Component_Library SHALL provide these primitive components: Button, Input, Select, Dialog, Tabs, Tooltip, Dropdown_Menu, Alert, Badge, Card
2. WHEN a component accepts variants, THE Component_Library SHALL expose them via a `variant` prop with typed literal values and a default of the first value: Button ('primary' | 'secondary' | 'danger' | 'ghost', default 'primary'), Alert ('error' | 'success' | 'warning', default 'error'), Badge ('default' | 'gold' | 'pink' | 'green' | 'red', default 'default'), Card ('default' | 'highlight' | 'success' | 'error', default 'default'), Input ('default' | 'mono', default 'default')
3. WHEN a component accepts size options, THE Component_Library SHALL expose them via a `size` prop with typed literal values ('sm' | 'md' | 'lg') defaulting to 'md'
4. THE Component_Library SHALL expose all primitive components via a barrel export from `$lib/components/ui/index.js`
5. WHEN a component renders interactive elements, THE Component_Library SHALL support keyboard navigation as defined by the corresponding WAI-ARIA pattern
6. THE Component_Library SHALL use Svelte 5 runes ($props) for all component prop declarations
7. IF a component receives a `variant` or `size` value not in its typed literal set, THEN THE Component_Library SHALL fall back to the default value for that prop
8. THE Component_Library SHALL forward all unrecognized HTML attributes to the root DOM element of each component via rest props ($props spread)

### Requirement 4: Visual Identity Preservation

**User Story:** As a product owner, I want the component refactor to maintain the existing editorial gold/porcelain aesthetic, so that users experience no visual regression.

#### Acceptance Criteria

1. THE Component_Library SHALL apply the 60-30-10 color rule: porcelana for backgrounds (60%), tinta for text (30%), ouro for interactive accents (10%)
2. THE Component_Library SHALL use --display (Fraunces) for headings (h1, h2, h3) and --ui (Archivo) for body text, buttons, and input labels
3. THE Component_Library SHALL use --raio (14px) as the default border-radius for card-level surfaces (.card, .data-table, .painel) and --raio-sm (8px) for controls (.btn, .input, .btn-icon)
4. WHEN rendering a primary action, THE Component_Library SHALL use --ouro as the background color and white as the text color
5. WHEN a component receives hover interaction, THE Component_Library SHALL transition the affected properties (background, border-color, opacity) to the corresponding hover token (--ouro-hover, --rosa-hover) using CSS ease timing within 150ms
6. THE Component_Library SHALL maintain all existing WCAG AA contrast ratios documented in app.css (≥4.5:1 for normal text, ≥3:1 for large text and non-text UI components)
7. WHEN a focusable element (a, button, [tabindex]) receives focus-visible, THE Component_Library SHALL display a 2px solid outline using --ouro with a 2px offset
8. WHILE the user has prefers-reduced-motion enabled, THE Component_Library SHALL suppress all animations and transitions by setting their duration to 0ms

### Requirement 5: Progressive Migration Strategy

**User Story:** As a frontend developer, I want to migrate existing components incrementally without breaking the application, so that the refactor can be shipped in stages.

#### Acceptance Criteria

1. WHEN a new component replaces an existing one, THE Migration_Adapter SHALL accept the same prop interface (prop names, types, and defaults) as the legacy component and produce equivalent DOM output
2. WHILE migration is in progress, THE Component_Library SHALL scope all new component styles using Svelte component-scoped CSS to prevent conflicts with legacy global utility classes in app.css
3. THE Component_Library SHALL migrate components in four phases: Phase 1 (tokens file), Phase 2 (primitives: Button, Input, Badge, Alert), Phase 3 (composites: Select, Tabs, Dialog, Tooltip, Dropdown_Menu), Phase 4 (page-level: FilterBar, NavDrawer, ProductCard)
4. WHEN a legacy component is fully replaced and all consumers import the new version, THE Migration_Adapter SHALL be deleted and the legacy component file removed in the same commit
5. IF a migration introduces a visual regression detected by Playwright visual comparison or manual review, THEN THE Component_Library SHALL retain the previous styling as a `variant="legacy"` option until all consumers are updated

### Requirement 6: Architecture Decision Record

**User Story:** As a team member, I want the technical decisions documented in an ADR, so that future contributors understand why Bits UI was chosen and how the token system works.

#### Acceptance Criteria

1. THE Component_Library SHALL include an ADR located at `docs/adr/` following the numbering convention ADR-NNNN (e.g., ADR-0001-component-library.md) documenting the selection of Bits UI over alternatives (Melt UI, shadcn-svelte, Skeleton UI)
2. THE ADR SHALL include sections: Title, Status (Accepted), Context, Decision, Evaluation Criteria (accessibility WCAG, headless/unstyled architecture, Svelte 5 runes compatibility, active maintenance, bundle size), and Consequences
3. THE ADR SHALL document the rejected alternatives with rationale: shadcn-svelte (requires Tailwind CSS which the project does not use), Skeleton UI (opinionated styling conflicts with existing visual identity), Melt UI (lower-level builder API requires more boilerplate than Bits UI's component abstraction)
4. THE ADR SHALL define the token naming convention: preserve existing CSS variable names from app.css unchanged (e.g., --ouro, --porcelana, --raio) and document the category grouping (color, spacing, typography, surface) used in the Token_Schema file
5. THE ADR SHALL specify that each component resides in `$lib/components/ui/{ComponentName}.svelte`, is exported from `$lib/components/ui/index.js`, and uses a single-file component pattern with scoped styles

### Requirement 7: Bundle Size and Performance

**User Story:** As a frontend developer, I want the component library to remain lightweight, so that page load performance on Cloudflare Pages is not degraded.

#### Acceptance Criteria

1. THE Component_Library SHALL tree-shake unused Bits UI primitives so that only imported components contribute to the bundle
2. WHEN a component is not used on a page, THE Component_Library SHALL ensure zero JavaScript is shipped for that component via SvelteKit code splitting
3. THE Component_Library SHALL add no more than 15KB gzipped to the total JavaScript bundle size of any single page route compared to the baseline measured from the last production build before migration begins
4. THE Component_Library SHALL not introduce runtime CSS-in-JS; all styles SHALL be static CSS resolved at build time
5. IF the 15KB gzipped budget is exceeded during a production build, THEN THE Component_Library SHALL fail the build step with an error message indicating which route exceeded the budget and by how much

### Requirement 8: Testing and Quality Assurance

**User Story:** As a frontend developer, I want automated tests verifying component behavior and accessibility, so that regressions are caught before deployment.

#### Acceptance Criteria

1. THE Component_Library SHALL include Vitest unit tests for each primitive component defined in Requirement 3 (Button, Input, Select, Dialog, Tabs, Tooltip, Dropdown_Menu, Alert, Badge, Card) verifying prop handling, variant rendering, and event emission with at least one test case per variant per component
2. THE Component_Library SHALL include accessibility tests using axe-core verifying ARIA attributes and keyboard navigation for all interactive components against WCAG 2.1 Level AA rules and the corresponding WAI-ARIA Authoring Practices pattern
3. WHEN a component is rendered in a test environment, THE Component_Library SHALL produce zero violations as reported by axe-core with WCAG 2.1 Level AA ruleset enabled
4. THE Component_Library SHALL include Playwright e2e tests covering at minimum: navigation between pages, opening and closing a Dialog, selecting an option from Select, switching Tabs, and submitting a form with Input and Button
5. IF a component test fails in CI, THEN THE Component_Library SHALL block the deployment pipeline and report the failing test name and component in the CI output
6. THE Component_Library SHALL enforce a minimum of 80% branch coverage for unit tests across all primitive components, failing the CI build if coverage drops below this threshold
