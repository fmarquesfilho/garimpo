# Requirements Document

## Introduction

Dark mode for the Garimpo web application. The feature adds a second colour scheme that activates automatically via `prefers-color-scheme` and can be overridden manually through a toggle. Dark tokens are defined as CSS custom property overrides in the existing design token system (`tokens.css`), preserving the editorial gold/pink aesthetic while meeting WCAG AA contrast requirements. The implementation must avoid Flash of Unstyled Content on Cloudflare Pages (static adapter, no SSR).

## Glossary

- **Theme_Engine**: The client-side module responsible for detecting, applying, and persisting the active colour scheme.
- **Token_System**: The set of CSS custom properties defined in `tokens.css` that controls all visual values across components.
- **Toggle**: The UI control that allows users to switch between light mode, dark mode, and system-preference mode.
- **FOUC**: Flash of Unstyled Content — a visible flicker when the rendered theme does not match the intended theme on first paint.
- **System_Preference**: The value reported by the `prefers-color-scheme` media query from the user's operating system or browser.
- **Theme_Attribute**: The `data-theme` attribute on the root HTML element (`<html>`) that selects the active token set.

## Requirements

### Requirement 1: Automatic Theme Detection

**User Story:** As a user, I want the application to match my operating system's colour preference automatically, so that the interface feels native without manual configuration.

#### Acceptance Criteria

1. WHEN the System_Preference reports `dark`, THE Theme_Engine SHALL apply the dark token set by setting Theme_Attribute to `"dark"` on the document root element.
2. WHEN the System_Preference reports `light` or has no preference, THE Theme_Engine SHALL apply the light token set by setting Theme_Attribute to `"light"` on the document root element.
3. WHEN the System_Preference changes while the application is open AND no manual override is stored, THE Theme_Engine SHALL update Theme_Attribute within 100ms of the change event.

### Requirement 2: Manual Theme Override

**User Story:** As a user, I want to manually override the system colour preference, so that I can choose the mode I prefer regardless of my OS setting.

#### Acceptance Criteria

1. THE Toggle SHALL present three states: light, dark, and system (follow OS preference).
2. WHEN the user selects a mode via the Toggle, THE Theme_Engine SHALL apply the selected mode immediately and persist the choice to localStorage under the key `"theme"`.
3. WHEN the application loads and localStorage contains a stored theme value, THE Theme_Engine SHALL apply the stored value instead of the System_Preference.
4. WHEN the user selects the system state via the Toggle, THE Theme_Engine SHALL remove the stored theme value from localStorage and revert to System_Preference detection.

### Requirement 3: Flash-of-Unstyled-Content Prevention

**User Story:** As a user, I want the correct theme applied before first paint, so that I do not experience a colour flash on page load.

#### Acceptance Criteria

1. THE Theme_Engine SHALL execute a blocking inline script in the HTML `<head>` that reads localStorage and System_Preference, then sets Theme_Attribute before any stylesheet or body content is parsed.
2. THE Theme_Engine SHALL complete theme resolution in under 5ms on a median mobile device (no network calls, no async operations).
3. IF the blocking script fails or throws an error, THEN THE Theme_Engine SHALL default to the light theme to maintain a usable state.

### Requirement 4: Dark Token Definitions

**User Story:** As a developer, I want dark-mode tokens defined alongside the existing light tokens, so that maintaining and auditing the palette remains straightforward.

#### Acceptance Criteria

1. THE Token_System SHALL define a `:root[data-theme="dark"]` rule block in `tokens.css` that overrides all colour custom properties with dark-mode values.
2. THE Token_System SHALL preserve all non-colour tokens (spacing, typography, radii, shadows) unchanged between light and dark modes.
3. THE Token_System SHALL maintain the gold accent (`--ouro`) and pink accent (`--rosa`) recognisable in both modes by using lighter or more saturated variants rather than simple inversion.

### Requirement 5: WCAG AA Contrast Compliance

**User Story:** As a user with visual impairments, I want all text to be legible in dark mode, so that the application remains accessible.

#### Acceptance Criteria

1. THE Token_System SHALL define dark-mode text and background colour pairings that achieve a minimum contrast ratio of 4.5:1 for normal text (below 18pt).
2. THE Token_System SHALL define dark-mode text and background colour pairings that achieve a minimum contrast ratio of 3:1 for large text (18pt and above) and UI components.
3. THE Token_System SHALL define feedback colour variants (success, error, warning) in dark mode that achieve a minimum contrast ratio of 4.5:1 against their respective background tokens.

### Requirement 6: Smooth Theme Transition

**User Story:** As a user, I want theme changes to animate smoothly, so that the switch feels polished rather than jarring.

#### Acceptance Criteria

1. WHEN Theme_Attribute changes after initial load, THE Token_System SHALL apply a CSS transition of 200ms duration on `background-color` and `color` properties to all elements.
2. WHILE the user has `prefers-reduced-motion: reduce` active, THE Token_System SHALL apply theme changes instantly with no transition.
3. THE Theme_Engine SHALL disable the transition CSS during initial page load to prevent FOUC, and enable it only after the first paint completes.

### Requirement 7: Bits UI Compatibility

**User Story:** As a developer, I want dark mode to work correctly with Bits UI components, so that accessible headless components render consistently in both themes.

#### Acceptance Criteria

1. THE Token_System SHALL scope dark-mode overrides exclusively via `:root[data-theme="dark"]` without conflicting with Bits UI `data-*` state attributes.
2. THE Theme_Engine SHALL set Theme_Attribute only on the `<html>` element to avoid specificity conflicts with component-level data attributes.
3. WHEN a Bits UI component uses `data-state` or `data-highlighted` attributes, THE Token_System SHALL preserve correct styling by referencing only CSS custom properties (not hard-coded colours).

### Requirement 8: Toggle UI Placement

**User Story:** As a user, I want to find the theme toggle easily, so that switching modes does not require digging through settings.

#### Acceptance Criteria

1. THE Toggle SHALL be positioned in the application header bar, visible when the user is authenticated.
2. THE Toggle SHALL display an icon indicating the current active mode (sun for light, moon for dark, monitor for system).
3. THE Toggle SHALL include an accessible label via `aria-label` describing its current state and action.
4. WHEN the user is not authenticated, THE Toggle SHALL be hidden from the interface.

### Requirement 9: Energy Efficiency Documentation

**User Story:** As a product owner, I want the dark-mode decision documented in the backlog, so that the team understands the rationale including OLED energy savings.

#### Acceptance Criteria

1. WHEN dark mode implementation is complete, THE team SHALL create a backlog task documenting the feature rationale, including OLED energy efficiency benefits of dark backgrounds.
2. THE documentation SHALL reference the token architecture decision (`:root[data-theme="dark"]` override approach) and the WCAG AA compliance commitment.
