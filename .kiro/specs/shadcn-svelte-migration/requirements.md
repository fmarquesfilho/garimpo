# Requirements Document

## Introduction

Migration of the Garimpei frontend from custom hand-written CSS components to shadcn-svelte, a component library built on top of Bits UI (already in use). This involves installing Tailwind CSS, configuring shadcn-svelte, replacing 10 custom UI primitives (Button, Alert, Badge, Card, Input, Select, Tabs, Dialog, Tooltip, DropdownMenu) with shadcn-svelte equivalents, and mapping existing design tokens to the shadcn-svelte theming system. The migration is incremental — each component is swapped one at a time while keeping tests green and CI passing.

## Glossary

- **shadcn-svelte**: A port of shadcn/ui for Svelte that provides copy-paste component primitives built on Bits UI, styled with Tailwind CSS
- **Tailwind_CSS**: A utility-first CSS framework that generates classes from configuration; required by shadcn-svelte
- **Bits_UI**: A headless component library for Svelte providing accessible primitives (already a project dependency at v2.18.1)
- **Design_Tokens**: The CSS custom properties defined in `tokens.css` (82 LOC) governing colors, spacing, typography, and surfaces for both light and dark modes
- **Migration_Tool**: The `npx shadcn-svelte@latest` CLI that scaffolds components and configuration into the project
- **Component_Library**: The set of UI primitives and composites exported from `$lib/components/ui/index.js`
- **CI_Pipeline**: The GitHub Actions workflow that runs svelte-check, stylelint, ESLint, knip, vitest, and Playwright E2E tests
- **Theme_Config**: The Tailwind/shadcn-svelte configuration mapping project design tokens to utility classes (typically `tailwind.config.js` and a CSS layer)
- **Primitives**: The 5 simple custom components (Button, Alert, Badge, Card, Input) with hand-written CSS
- **Composites**: The 6 Bits UI-based wrapper components (Select, Tabs, Dialog, Tooltip, DropdownMenu, ThemeToggle)
- **Application_Components**: Higher-level components (DashPanel, MetricCard, PageHeader, etc.) that consume primitives and composites

## Requirements

### Requirement 1: Tailwind CSS Installation and Configuration

**User Story:** As a developer, I want Tailwind CSS installed and configured in the SvelteKit project, so that shadcn-svelte components can use utility classes for styling.

#### Acceptance Criteria

1. WHEN the Migration_Tool initialization is run, THE Theme_Config SHALL include Tailwind CSS v4 configured for the SvelteKit project with content paths covering `src/**/*.{svelte,js,ts}`
2. THE Theme_Config SHALL map all existing Design_Tokens CSS variables to Tailwind utility classes so that existing token values are usable as Tailwind colors, spacing, and typography values
3. WHEN a Svelte component uses a Tailwind utility class, THE build system SHALL generate the corresponding CSS without conflicting with existing component styles
4. THE Theme_Config SHALL preserve the existing dark mode behavior using the `data-theme="dark"` attribute on the root element
5. WHEN the CI_Pipeline runs `npm run lint:css`, THE stylelint configuration SHALL permit Tailwind directives (`@tailwind`, `@apply`, `@layer`) without reporting errors

### Requirement 2: shadcn-svelte Initialization

**User Story:** As a developer, I want shadcn-svelte initialized in the project, so that I can add pre-built accessible components incrementally.

#### Acceptance Criteria

1. WHEN the Migration_Tool `init` command is executed, THE project SHALL contain a `components.json` configuration file specifying the component output path as `$lib/components/ui`
2. THE Migration_Tool configuration SHALL use the project's existing Bits_UI dependency rather than installing a conflicting version
3. WHEN a shadcn-svelte component is added via the CLI, THE component source files SHALL be placed in `$lib/components/ui/` following the shadcn-svelte directory convention
4. THE shadcn-svelte utility file (`cn` function) SHALL be available at `$lib/utils` for class merging across all components

### Requirement 3: Design Token Mapping to shadcn-svelte Theme

**User Story:** As a developer, I want the existing Garimpo design tokens mapped to shadcn-svelte's theming system, so that migrated components maintain the brand identity (gold accent, porcelain neutrals, warm dark mode).

#### Acceptance Criteria

1. THE Theme_Config SHALL define shadcn-svelte CSS variables (`--primary`, `--secondary`, `--destructive`, `--muted`, `--accent`, `--background`, `--foreground`, `--card`, `--border`, `--ring`) using values from the existing Design_Tokens
2. THE Theme_Config SHALL map `--ouro` to the primary color, `--rosa` to the destructive color, `--porcelana` to the muted/background color, and `--tinta` to the foreground color
3. WHEN `data-theme="dark"` is set on the root element, THE Theme_Config SHALL apply the dark mode token overrides defined in `tokens.css` to the shadcn-svelte CSS variables
4. THE Theme_Config SHALL preserve the existing font families (`--display`, `--ui`, `--mono`) as Tailwind font-family utilities
5. THE Theme_Config SHALL preserve existing spacing scale (`--r1` through `--r12`) and border radius tokens (`--raio`, `--raio-sm`, `--raio-lg`, `--raio-full`) as Tailwind utilities

### Requirement 4: Primitive Component Migration

**User Story:** As a developer, I want the 5 custom primitive components (Button, Alert, Badge, Card, Input) replaced with shadcn-svelte equivalents, so that styling is consistent, polished, and maintainable via Tailwind.

#### Acceptance Criteria

1. WHEN the Button component is migrated, THE new Component_Library Button SHALL support the same variant API (`primary`, `secondary`, `danger`, `ghost`) and size API (`sm`, `md`, `lg`) as the current implementation
2. WHEN the Alert component is migrated, THE new Component_Library Alert SHALL support `error`, `success`, and `warning` variants with the same semantic feedback colors from Design_Tokens
3. WHEN the Badge component is migrated, THE new Component_Library Badge SHALL accept the same variant props as the current implementation
4. WHEN the Card component is migrated, THE new Component_Library Card SHALL render with the same surface styling (border, radius, shadow) mapped from Design_Tokens
5. WHEN the Input component is migrated, THE new Component_Library Input SHALL support the same prop interface (label, placeholder, disabled, error state) as the current implementation
6. WHEN any primitive component is migrated, THE new component SHALL maintain the same public prop interface so that consuming components require zero or minimal changes to their template code
7. WHEN any primitive component is migrated, THE component SHALL pass all existing unit tests and E2E tests that exercise it

### Requirement 5: Composite Component Migration

**User Story:** As a developer, I want the Bits UI composite wrappers (Select, Tabs, Dialog, Tooltip, DropdownMenu, ThemeToggle) restyled with shadcn-svelte patterns, so that they match the new design system visually while retaining full accessibility.

#### Acceptance Criteria

1. WHEN the Select composite is migrated, THE new Select SHALL retain the same prop interface (`value`, `label`, `options`, `placeholder`, `size`, `disabled`) and two-way binding behavior
2. WHEN the Tabs composite is migrated, THE new Tabs SHALL retain keyboard navigation and ARIA attributes provided by Bits_UI
3. WHEN the Dialog composite is migrated, THE new Dialog SHALL retain focus trap, escape-to-close, and backdrop click-to-close behavior
4. WHEN the Tooltip composite is migrated, THE new Tooltip SHALL retain the `Tooltip.Provider` ancestor requirement and keyboard/focus activation
5. WHEN the DropdownMenu composite is migrated, THE new DropdownMenu SHALL retain keyboard navigation and proper ARIA menu roles
6. WHEN any composite component is migrated, THE component SHALL use shadcn-svelte Tailwind styling instead of `:global()` CSS rules and scoped styles
7. WHEN any composite component is migrated, THE component SHALL pass axe-core accessibility audits with zero violations

### Requirement 6: Incremental Migration Strategy

**User Story:** As a developer, I want to migrate components one at a time without breaking the application, so that the project stays deployable throughout the migration.

#### Acceptance Criteria

1. WHILE the migration is in progress, THE CI_Pipeline SHALL pass after each individual component migration (svelte-check, stylelint, ESLint, knip, vitest, Playwright)
2. WHILE the migration is in progress, THE existing Application_Components that consume the migrated primitives SHALL continue to render correctly
3. WHEN a component is migrated, THE old hand-written CSS for that component SHALL be removed so that no dead CSS accumulates
4. WHEN all Primitives and Composites have been migrated, THE `tokens.css` file SHALL be evaluated for removal or reduction since colors and spacing will live in the Tailwind/shadcn-svelte theme layer
5. IF a migrated component causes a test failure, THEN THE migration for that component SHALL be reverted or fixed before proceeding to the next component

### Requirement 7: CI Pipeline Compatibility

**User Story:** As a developer, I want all existing CI checks to pass with the new Tailwind + shadcn-svelte setup, so that code quality and accessibility guarantees are maintained.

#### Acceptance Criteria

1. WHEN Tailwind CSS is installed, THE stylelint configuration SHALL be updated to allow Tailwind-specific at-rules and the `color-no-hex` rule SHALL continue to enforce token usage in custom CSS
2. WHEN shadcn-svelte components are added, THE knip configuration SHALL recognize shadcn-svelte's component file structure so that generated files are not flagged as dead code
3. WHEN the migration is complete, THE `npm run check` command (svelte-check) SHALL report zero errors
4. WHEN the migration is complete, THE `npm run lint:js` command (ESLint) SHALL report zero warnings
5. WHEN the migration is complete, THE `npm run test:unit` command (vitest) SHALL pass all 141 existing unit tests
6. WHEN the migration is complete, THE `npm run test` command (Playwright) SHALL pass all 53 existing E2E tests

### Requirement 8: Removal of Legacy CSS

**User Story:** As a developer, I want all hand-written CSS in migrated components removed, so that styling is fully managed by Tailwind utilities and the codebase has a single styling approach.

#### Acceptance Criteria

1. WHEN a primitive component is migrated, THE `<style>` block in that component file SHALL be removed entirely
2. WHEN a composite component is migrated, THE `:global()` CSS rules for Bits UI class overrides SHALL be removed entirely
3. WHEN all component migrations are complete, THE Component_Library SHALL contain zero `<style>` blocks in primitive and composite component files
4. WHEN all component migrations are complete, THE knip audit SHALL report zero dead exports related to migrated components

### Requirement 9: Documentation

**User Story:** As a developer, I want the migration documented, so that the team understands the new component patterns and how to add new components going forward.

#### Acceptance Criteria

1. WHEN the migration is complete, THE project SHALL contain updated documentation describing how to add new shadcn-svelte components via the CLI
2. WHEN the migration is complete, THE project SHALL contain documentation mapping old component APIs to new shadcn-svelte equivalents
3. WHEN the migration is complete, THE project SHALL contain documentation explaining the theme configuration and how to modify brand colors or tokens
4. WHEN the migration is complete, THE existing `docs/componentes.md` file SHALL be updated to reflect the new component library structure
